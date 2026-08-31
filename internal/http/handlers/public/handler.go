// File: internal/http/handlers/public/handler.go
//
// Musteri terefinin kesf endpoint-leri.
//
// Bunlar auth telebi etmir: musteri hele qeydiyyatdan kecmemis de
// xestexana/berber secib bos vaxtlara baxa bilmelidir. Bron yaratmaq
// ucun ise artiq login lazimdir (POST /bookings qorunur).
//
// Yalniz oxuma emeliyyatlaridir ve hec bir sexsi melumat qaytarmir:
// biznes adi, isci adi/vezifesi, xidmet ve hesablanmis bos vaxtlar.
package public

import (
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	availabilityDomain "github.com/OrkhanNajaf1i/booking-service/internal/domain/availability"
	businessDomain "github.com/OrkhanNajaf1i/booking-service/internal/domain/business"
	"github.com/OrkhanNajaf1i/booking-service/internal/domain/catalog"
	serviceDomain "github.com/OrkhanNajaf1i/booking-service/internal/domain/service"
	staffDomain "github.com/OrkhanNajaf1i/booking-service/internal/domain/staff"
	"github.com/OrkhanNajaf1i/booking-service/internal/logger"
	"github.com/google/uuid"
)

type Handler struct {
	businesses   businessDomain.Service
	staff        staffDomain.Service
	services     serviceDomain.ServiceUseCase
	availability availabilityDomain.Service
	log          logger.Logger
}

func NewHandler(
	businesses businessDomain.Service,
	staff staffDomain.Service,
	services serviceDomain.ServiceUseCase,
	availability availabilityDomain.Service,
	log logger.Logger,
) *Handler {
	return &Handler{
		businesses:   businesses,
		staff:        staff,
		services:     services,
		availability: availability,
		log:          log,
	}
}

// ============================================================
// BIZNESLER
// ============================================================

// BusinessCard – siyahida gosterilen minimal melumat.
type BusinessCard struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`

	// Category sabit kateqoriya slug-idir; qruplasdirma bunun uzre gedir.
	Category     string `json:"category"`
	CategoryName string `json:"category_name"`
	CategoryIcon string `json:"category_icon"`

	// ServiceCategory sahibin ozunun yazdigi metndir — kartda alt basliq
	// kimi gorunur, cunki "Kardioloq" "Hekim"-den daha melumatlidir.
	ServiceCategory string `json:"service_category"`
	Industry        string `json:"industry"`
	Phone           string `json:"phone"`
	BusinessType    string `json:"business_type"`

	// Yer melumati — en yaxin (koordinat verilmeyibse birinci) filialdan.
	City    string `json:"city,omitempty"`
	Address string `json:"address,omitempty"`

	// DistanceKm yalniz sorguda lat/lng olanda dolur.
	DistanceKm *float64 `json:"distance_km,omitempty"`
}

// ListBusinesses – GET /api/v1/public/businesses
//
// @Summary      Aktiv bizneslerin siyahisi
// @Description  Musteri xestexana / klinika / berberxana secmek ucun. Auth telem olunmur.
// @Tags         Public
// @Produce      json
// @Param        category  query string false "Kateqoriya slug-i (mes. dentist)"
// @Param        q         query string false "Ad / sahe uzre axtaris"
// @Param        lat       query number false "Musterinin en dairesi — yaxinliq siralamasi ucun"
// @Param        lng       query number false "Musterinin uzunluq dairesi"
// @Param        radius_km query number false "lat/lng verilibse: bu mesafeden uzaqlari kesir"
// @Success      200 {object} SuccessResponse
// @Router       /public/businesses [get]
func (h *Handler) ListBusinesses(w http.ResponseWriter, r *http.Request) {
	all, err := h.businesses.ListBookableBusinesses(r.Context())
	if err != nil {
		h.writeInternal(w, err)
		return
	}

	query := r.URL.Query()
	category := strings.TrimSpace(query.Get("category"))
	search := normalizeSearch(query.Get("q"))

	from, hasOrigin := parseOrigin(query)
	radiusKm, hasRadius := parseFloat(query.Get("radius_km"))

	cards := make([]BusinessCard, 0, len(all))
	for _, item := range all {
		// Deaktiv bizneslər musteriye gosterilmir.
		if !item.IsActive {
			continue
		}

		resolved := catalog.ResolveWith(item.CategorySlug, item.ServiceCategory, item.Industry)

		// Kateqoriya slug uzre suzulur; kohne muraciet uyumlulugu ucun
		// gorunen ad da qebul edilir.
		if category != "" &&
			!strings.EqualFold(resolved.Slug, category) &&
			!strings.EqualFold(resolved.Name, category) {
			continue
		}

		if search != "" && !matchesSearch(item, resolved, search) {
			continue
		}

		card := BusinessCard{
			ID:              item.ID,
			Name:            item.Name,
			Category:        resolved.Slug,
			CategoryName:    resolved.Name,
			CategoryIcon:    resolved.Icon,
			ServiceCategory: item.ServiceCategory,
			Industry:        item.Industry,
			Phone:           item.Phone,
			BusinessType:    string(item.BusinessType),
		}

		if hasOrigin {
			distance, ok := item.NearestDistanceKm(from.lat, from.lng)
			switch {
			case ok:
				if hasRadius && distance > radiusKm {
					continue
				}
				rounded := math.Round(distance*10) / 10
				card.DistanceKm = &rounded
			case hasRadius:
				// Harada oldugu bilinmeyen biznes mesafe suzgecinden
				// kecmemelidir — istifadeci "yaxinliqdakilar" isteyib.
				continue
			}
		}

		fillLocation(&card, item, from, hasOrigin)
		cards = append(cards, card)
	}

	// Yaxinliq isteyibse en yaxin evvel; mesafesiz olanlar sona.
	if hasOrigin {
		sort.SliceStable(cards, func(i, j int) bool {
			left, right := cards[i].DistanceKm, cards[j].DistanceKm
			switch {
			case left == nil:
				return false
			case right == nil:
				return true
			default:
				return *left < *right
			}
		})
	}

	writeSuccess(w, http.StatusOK, "", cards)
}

// originPoint – musterinin noqtesi.
type originPoint struct{ lat, lng float64 }

// parseOrigin – lat ve lng birlikde ve etibarli olmalidir. Yarimciq
// deyer sessizce goz ardi edilir: siyahi yene qaytarilir, sadece
// mesafesiz. Kesf ekrani koordinat olmadan da islemelidir.
func parseOrigin(query url.Values) (originPoint, bool) {
	lat, okLat := parseFloat(query.Get("lat"))
	lng, okLng := parseFloat(query.Get("lng"))
	if !okLat || !okLng {
		return originPoint{}, false
	}
	if lat < -90 || lat > 90 || lng < -180 || lng > 180 {
		return originPoint{}, false
	}
	return originPoint{lat: lat, lng: lng}, true
}

func parseFloat(raw string) (float64, bool) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return 0, false
	}
	value, err := strconv.ParseFloat(trimmed, 64)
	if err != nil {
		return 0, false
	}
	return value, true
}

// fillLocation – kartda gosterilecek filial. Musterinin noqtesi
// bellidirse en yaxin filial, deyilse birinci filial goturulur.
func fillLocation(
	card *BusinessCard,
	item *businessDomain.BookableBusiness,
	from originPoint,
	hasOrigin bool,
) {
	if len(item.Locations) == 0 {
		return
	}

	chosen := item.Locations[0]
	if hasOrigin {
		best := -1.0
		for _, location := range item.Locations {
			if !location.HasCoordinates() {
				continue
			}
			distance := businessDomain.HaversineKm(
				from.lat, from.lng, *location.Latitude, *location.Longitude,
			)
			if best < 0 || distance < best {
				best = distance
				chosen = location
			}
		}
	}

	card.City = chosen.City
	card.Address = chosen.Address
}

// CategoryCard – kesf ekraninda gosterilen kateqoriya.
type CategoryCard struct {
	Slug  string `json:"slug"`
	Name  string `json:"name"`
	Icon  string `json:"icon"`
	Count int    `json:"count"`
}

// ListCategories – GET /api/v1/public/categories
//
// Musteri once sahe secir ("Berber", "Dis hekimi"), sonra hemin
// sahedeki biznesler siyahilanir. Kateqoriya ayrica cedvel deyil —
// bizneslerin serbest metni sabit taksonomiyaya baglanir (bax:
// domain/catalog). Belelikle "Dis Hekimi" ve "Stomatoloq" bir yerde
// gorunur, siyahi biznes sayi artdiqca parcalanmir.
//
// @Summary      Xidmet kateqoriyalari
// @Description  Her kateqoriyada nece aktiv biznes oldugunu da qaytarir.
// @Tags         Public
// @Produce      json
// @Param        lat       query number false "Verilibse yalniz radius daxilindekiler sayilir"
// @Param        lng       query number false "Musterinin uzunluq dairesi"
// @Param        radius_km query number false "Yaxinliq radiusu"
// @Success      200 {object} SuccessResponse
// @Router       /public/categories [get]
func (h *Handler) ListCategories(w http.ResponseWriter, r *http.Request) {
	all, err := h.businesses.ListBookableBusinesses(r.Context())
	if err != nil {
		h.writeInternal(w, err)
		return
	}

	query := r.URL.Query()
	from, hasOrigin := parseOrigin(query)
	radiusKm, hasRadius := parseFloat(query.Get("radius_km"))

	counts := make(map[string]int)
	for _, item := range all {
		if !item.IsActive {
			continue
		}

		// Yaxinliq secilibse sayac da hemin radiusu gostermelidir —
		// eks halda "3 hekim" yazir, acanda bos cixir.
		if hasOrigin && hasRadius {
			distance, ok := item.NearestDistanceKm(from.lat, from.lng)
			if !ok || distance > radiusKm {
				continue
			}
		}

		counts[catalog.ResolveWith(item.CategorySlug, item.ServiceCategory, item.Industry).Slug]++
	}

	// Bos kateqoriya musteriye gosterilmir: acanda bos ekran cixarir.
	// all=true ise sahibin secim siyahisi ucundur — orada butun
	// kateqoriyalar lazimdir, sayindan asili olmayaraq. Orada "Diger"
	// de yoxdur: o, pese deyil.
	includeEmpty := query.Get("all") == "true"

	source := catalog.All()
	if includeEmpty {
		source = catalog.Selectable()
	}

	cards := make([]CategoryCard, 0, len(source))
	for _, category := range source {
		count := counts[category.Slug]
		if count == 0 && !includeEmpty {
			continue
		}
		cards = append(cards, CategoryCard{
			Slug:  category.Slug,
			Name:  category.Name,
			Icon:  category.Icon,
			Count: count,
		})
	}

	// Kesf ekraninda coxlu biznesi olan sahe once gorunsun; secim
	// siyahisinda ise taksonomiyanin oz sirasi saxlanilir ki, sahib
	// her defe eyni yerde tapsin.
	//
	// "Diger" isteniilen halda EN SONDA qalir: o, pese deyil, sadece
	// kateqoriyasi secilmemis kohne setirlerin yeridir — sayi cox olsa
	// da siyahinin basina cixmamalidir.
	if !includeEmpty {
		sort.Slice(cards, func(i, j int) bool {
			leftOther := cards[i].Slug == catalog.SlugOther
			rightOther := cards[j].Slug == catalog.SlugOther
			if leftOther != rightOther {
				return rightOther
			}
			if cards[i].Count != cards[j].Count {
				return cards[i].Count > cards[j].Count
			}
			return cards[i].Name < cards[j].Name
		})
	}

	writeSuccess(w, http.StatusOK, "", cards)
}

// matchesSearch – ad, sahibin yazdigi sahe, industry, kateqoriya adi ve
// filial unvani uzre axtaris. Azerbaycan herfleri ASCII-ye endirilir ki,
// "dis" yazan adam "Diş Həkimi"ni tapsin.
func matchesSearch(
	item *businessDomain.BookableBusiness,
	resolved catalog.Category,
	needle string,
) bool {
	fields := []string{item.Name, item.ServiceCategory, item.Industry, resolved.Name}
	for _, location := range item.Locations {
		fields = append(fields, location.City, location.Address, location.Name)
	}

	for _, field := range fields {
		if strings.Contains(normalizeSearch(field), needle) {
			return true
		}
	}
	return false
}

// normalizeSearch – axtaris ucun kicik herf + Azerbaycan herflerinin
// ASCII qarsiligi.
func normalizeSearch(text string) string {
	var builder strings.Builder
	builder.Grow(len(text))

	for _, symbol := range strings.ToLower(strings.TrimSpace(text)) {
		switch symbol {
		case 'ə':
			builder.WriteByte('e')
		case 'ı':
			builder.WriteByte('i')
		case 'ş':
			builder.WriteByte('s')
		case 'ç':
			builder.WriteByte('c')
		case 'ğ':
			builder.WriteByte('g')
		case 'ö':
			builder.WriteByte('o')
		case 'ü':
			builder.WriteByte('u')
		default:
			builder.WriteRune(symbol)
		}
	}
	return builder.String()
}

// GetBusiness – GET /api/v1/public/businesses/{id}
//
// @Summary      Biznes detallari
// @Tags         Public
// @Produce      json
// @Param        id path string true "Biznes ID"
// @Success      200 {object} SuccessResponse
// @Router       /public/businesses/{id} [get]
func (h *Handler) GetBusiness(w http.ResponseWriter, r *http.Request) {
	businessID, ok := h.pathID(w, r)
	if !ok {
		return
	}

	found, err := h.businesses.GetBusinessByID(r.Context(), businessID)
	if err != nil || found == nil || !found.IsActive {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "biznes tapilmadi")
		return
	}

	resolved := catalog.ResolveWith(found.CategorySlug, found.ServiceCategory, found.Industry)

	writeSuccess(w, http.StatusOK, "", BusinessCard{
		ID:              found.ID,
		Name:            found.Name,
		Category:        resolved.Slug,
		CategoryName:    resolved.Name,
		CategoryIcon:    resolved.Icon,
		Industry:        found.Industry,
		ServiceCategory: found.ServiceCategory,
		Phone:           found.Phone,
		BusinessType:    string(found.BusinessType),
	})
}

// ============================================================
// ISCILER
// ============================================================

// StaffCard – musteriye gosterilen isci melumati.
// E-poct/telefon qesden daxil edilmir.
type StaffCard struct {
	ID         uuid.UUID `json:"id"`
	FullName   string    `json:"full_name"`
	Title      string    `json:"title"`
	Department string    `json:"department"`
	Avatar     *string   `json:"avatar,omitempty"`
}

// ListStaff – GET /api/v1/public/businesses/{id}/staff
//
// @Summary      Biznesin isciləri
// @Description  Musteri hansi hekim/berber ile bron edeceyini secir.
// @Tags         Public
// @Produce      json
// @Param        id path string true "Biznes ID"
// @Success      200 {object} SuccessResponse
// @Router       /public/businesses/{id}/staff [get]
func (h *Handler) ListStaff(w http.ResponseWriter, r *http.Request) {
	businessID, ok := h.pathID(w, r)
	if !ok {
		return
	}

	members, err := h.staff.ListStaff(r.Context(), businessID)
	if err != nil {
		h.writeInternal(w, err)
		return
	}

	cards := make([]StaffCard, 0, len(members))
	for _, member := range members {
		if string(member.Status) != "active" {
			continue
		}
		cards = append(cards, StaffCard{
			ID:         member.ID,
			FullName:   member.FullName,
			Title:      member.Title,
			Department: member.Department,
			Avatar:     member.Avatar,
		})
	}

	writeSuccess(w, http.StatusOK, "", cards)
}

// ============================================================
// XIDMETLER
// ============================================================

type ServiceCard struct {
	ID              uuid.UUID `json:"id"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	DurationMinutes int       `json:"duration_minutes"`
	Price           float64   `json:"price"`
}

// ListServices – GET /api/v1/public/businesses/{id}/services
//
// @Summary      Biznesin xidmetleri
// @Tags         Public
// @Produce      json
// @Param        id       path  string true  "Biznes ID"
// @Param        staff_id query string false "Yalniz bu iscinin xidmetleri"
// @Success      200 {object} SuccessResponse
// @Router       /public/businesses/{id}/services [get]
func (h *Handler) ListServices(w http.ResponseWriter, r *http.Request) {
	businessID, ok := h.pathID(w, r)
	if !ok {
		return
	}

	var (
		items []*serviceDomain.Service
		err   error
	)

	if raw := r.URL.Query().Get("staff_id"); raw != "" {
		staffID, parseErr := uuid.Parse(raw)
		if parseErr != nil {
			writeError(w, http.StatusBadRequest, "INVALID_STAFF_ID", "staff_id duzgun deyil")
			return
		}
		items, err = h.services.GetStaffServices(r.Context(), businessID, staffID)
	} else {
		items, err = h.services.ListServices(r.Context(), businessID)
	}

	if err != nil {
		h.writeInternal(w, err)
		return
	}

	cards := make([]ServiceCard, 0, len(items))
	for _, item := range items {
		if !item.IsActive {
			continue
		}
		cards = append(cards, ServiceCard{
			ID:              item.ID,
			Name:            item.Name,
			Description:     item.Description,
			DurationMinutes: item.DurationMinutes,
			Price:           item.Price,
		})
	}

	writeSuccess(w, http.StatusOK, "", cards)
}

// ============================================================
// BOS VAXTLAR
// ============================================================

// GetAvailability – GET /api/v1/public/availability
//
// Qorunan /availability endpoint-i biznes konteksti JWT-den alir;
// burada ise musterinin biznesi olmadigi ucun business_id acıq verilir.
//
// @Summary      Bos vaxtlar (auth telem olunmur)
// @Tags         Public
// @Produce      json
// @Param        business_id query string true  "Biznes ID"
// @Param        staff_id    query string true  "Isci ID"
// @Param        service_id  query string false "Xidmet ID"
// @Param        from        query string false "YYYY-MM-DD"
// @Param        to          query string false "YYYY-MM-DD"
// @Success      200 {object} SuccessResponse
// @Router       /public/availability [get]
func (h *Handler) GetAvailability(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	businessID, err := uuid.Parse(query.Get("business_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BUSINESS_ID", "business_id duzgun deyil")
		return
	}

	staffID, err := uuid.Parse(query.Get("staff_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_STAFF_ID", "staff_id duzgun deyil")
		return
	}

	var serviceID *uuid.UUID
	if raw := query.Get("service_id"); raw != "" {
		parsed, parseErr := uuid.Parse(raw)
		if parseErr != nil {
			writeError(w, http.StatusBadRequest, "INVALID_SERVICE_ID", "service_id duzgun deyil")
			return
		}
		serviceID = &parsed
	}

	from, err := parseDate(query.Get("from"), time.Now())
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_FROM", "from tarixi YYYY-MM-DD olmalidir")
		return
	}

	to, err := parseDate(query.Get("to"), from.AddDate(0, 0, 6))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_TO", "to tarixi YYYY-MM-DD olmalidir")
		return
	}

	result, err := h.availability.GetAvailability(r.Context(), businessID, &availabilityDomain.AvailabilityQuery{
		StaffID:   staffID,
		ServiceID: serviceID,
		FromDate:  from,
		ToDate:    to,
	})
	if err != nil {
		var domainErr *availabilityDomain.Error
		if errors.As(err, &domainErr) {
			writeError(w, http.StatusBadRequest, domainErr.Code, domainErr.Message)
			return
		}
		h.writeInternal(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, "", result)
}

// ============================================================
// HELPERS
// ============================================================

func (h *Handler) pathID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	businessID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "biznes id duzgun deyil")
		return uuid.Nil, false
	}
	return businessID, true
}

func (h *Handler) writeInternal(w http.ResponseWriter, err error) {
	h.log.Error("Public endpoint xetasi", logger.Field{Key: "error", Value: err.Error()})
	writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "daxili xeta")
}

func parseDate(raw string, fallback time.Time) (time.Time, error) {
	if raw == "" {
		return fallback, nil
	}
	return time.Parse("2006-01-02", raw)
}

// ============================================================
// RESPONSE FORMATI
// ============================================================

type SuccessResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeSuccess(w http.ResponseWriter, status int, message string, data interface{}) {
	writeJSON(w, status, SuccessResponse{Success: true, Message: message, Data: data})
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, ErrorResponse{Success: false, Code: code, Message: message})
}
