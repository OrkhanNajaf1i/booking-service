// File: internal/domain/business/service.go
package business

import (
	"context"
	"fmt"
	"strings"

	"github.com/OrkhanNajaf1i/booking-service/internal/domain/catalog"
	"time"

	"github.com/google/uuid"
)

type BusinessService struct {
	repository Repository
	owners     OwnerLinker
	staff      StaffProvisioner
	// schedules qurulmayibsa biznes yene yaranir, sadece baslangic
	// qrafiki olmur — sahib onu ozu qurmalidir.
	schedules ScheduleProvisioner
	locations LocationProvisioner
	staffs    StaffCounter
}

func NewService(
	repository Repository,
	owners OwnerLinker,
	staff StaffProvisioner,
) *BusinessService {
	return &BusinessService{
		repository: repository,
		owners:     owners,
		staff:      staff,
	}
}

// WithSchedules – yeni biznese baslangic qrafiki qurulmasini aktivlesdirir.
func (service *BusinessService) WithSchedules(schedules ScheduleProvisioner) *BusinessService {
	service.schedules = schedules
	return service
}

// WithLocations – ilk filialin yaradilmasini aktivlesdirir.
func (service *BusinessService) WithLocations(locations LocationProvisioner) *BusinessService {
	service.locations = locations
	return service
}

// WithStaffCounter – rejim kecidinde isci sayini yoxlamaq ucun.
func (service *BusinessService) WithStaffCounter(staffs StaffCounter) *BusinessService {
	service.staffs = staffs
	return service
}

func (service *BusinessService) CreateBusiness(
	ctx context.Context,
	ownerID uuid.UUID,
	request *CreateBusinessRequest,
) (*Business, error) {
	if ownerID == uuid.Nil {
		return nil, NewBusinessError("INVALID_OWNER_ID", "Owner ID cannot be empty")
	}

	if request == nil {
		return nil, NewBusinessError("INVALID_REQUEST", "Request cannot be nil")
	}

	if err := service.validateCreateRequest(request); err != nil {
		return nil, err
	}
	business := NewBusiness(
		request.Name,
		request.Industry,
		request.ServiceCategory,
		request.Phone,
		request.BusinessType,
	)
	business.OwnerID = ownerID

	// Kateqoriya mecburidir ve sabit siyahidandir.
	//
	// Serbest metne guvenmek olmur: "berber"i sehv yazan, ya da
	// tamamile basqa soz isleden biznes kesf ekraninda "Diger"e dusur
	// ve musteri onu heç vaxt tapmir.
	if !catalog.IsSelectable(request.CategorySlug) {
		return nil, NewBusinessError(
			"CATEGORY_REQUIRED",
			"Kateqoriya secilmelidir",
		)
	}
	business.CategorySlug = strings.ToLower(strings.TrimSpace(request.CategorySlug))

	if err := service.validateBusiness(business); err != nil {
		return nil, err
	}

	// Filial melumati biznes setri yazilmazdan EVVEL yoxlanir: yarimciq
	// yaradilmis biznes qalmasin.
	if err := service.validateLocationDraft(request.Location); err != nil {
		return nil, err
	}

	if err := service.repository.Create(ctx, business); err != nil {
		return nil, fmt.Errorf("failed to create business: %w", err)
	}

	// Istifadecini biznese baglayiriq. Bu olmasa JWT-ye business_id
	// dusmur ve butun biznes-kontekstli endpoint-ler 400 qaytarir.
	if service.owners != nil {
		if err := service.owners.UpdateUserBusinessID(ctx, ownerID, business.ID, true); err != nil {
			return nil, fmt.Errorf("failed to link owner to business: %w", err)
		}
	}

	// Sahib ozu de xidmet gosteren sexsdir: randevu staff_id-ye baglanir,
	// ona gore profili derhal yaradilir. Eks halda "evvelce isci elave
	// edin" ekranindan kenara cixmaq mumkun olmur.
	if service.staff != nil {
		title := request.ServiceCategory
		if title == "" {
			title = request.Industry
		}
		staffID, err := service.staff.EnsureOwnerProfile(ctx, business.ID, ownerID, title)
		if err != nil {
			return nil, fmt.Errorf("failed to create owner staff profile: %w", err)
		}

		// Baslangic qrafiki: Bazar ertesi–Cume 09:00–18:00.
		//
		// Qrafiksiz biznes musteri terefinde bos gorunur. Bron onsuz da
		// TESDIQ teleb edir, ona gore defaultun bir qeder yanlis olmasi
		// duzeldile bilen seydir; qrafiksiz elan ise oluler.
		if service.schedules != nil {
			if err := service.schedules.EnsureDefaultSchedule(ctx, business.ID, staffID); err != nil {
				return nil, fmt.Errorf("failed to create default schedule: %w", err)
			}
		}
	}

	if service.locations != nil && request.Location != nil {
		draft := *request.Location
		if strings.TrimSpace(draft.Name) == "" {
			draft.Name = defaultLocationName
		}
		if err := service.locations.CreateFirstLocation(ctx, business.ID, draft); err != nil {
			return nil, fmt.Errorf("failed to create first location: %w", err)
		}
	}

	return business, nil
}

// SwitchMode – tek isci ↔ komanda rejimi.
//
// Rejim yalniz etiket deyil: komanda rejimi isci devetini acir, tek
// isci rejimi ise ekrani sadelesdirir. Ona gore kecid asikar
// emeliyyatdir — sahib "komandaya kecirem" deyir, sistem ozbasina qerar
// vermir.
func (service *BusinessService) SwitchMode(
	ctx context.Context,
	businessID, ownerID uuid.UUID,
	mode BusinessType,
) (*Business, error) {
	if businessID == uuid.Nil {
		return nil, NewBusinessError("INVALID_BUSINESS_ID", "Business ID cannot be empty")
	}
	if !mode.IsValid() {
		return nil, NewBusinessError("INVALID_BUSINESS_TYPE", "Rejim duzgun deyil")
	}

	current, err := service.repository.GetByID(ctx, businessID)
	if err != nil {
		return nil, fmt.Errorf("failed to get business: %w", err)
	}
	if current == nil {
		return nil, NewBusinessError("BUSINESS_NOT_FOUND", "Business not found")
	}
	if current.OwnerID != ownerID {
		return nil, NewBusinessError("UNAUTHORIZED", "Yalniz biznes sahibi rejimi deyise biler")
	}
	if current.BusinessType == mode {
		return current, nil
	}

	// Komandadan tek isci rejimine qayidis yalniz sahib tek qalanda.
	// Eks halda ekran "tek isleyirsiniz" yazardi, halbuki randevu
	// qebul eden basqa isciler var.
	if mode == BusinessTypeSolo && service.staffs != nil {
		count, err := service.staffs.CountActiveStaff(ctx, businessID)
		if err != nil {
			return nil, fmt.Errorf("failed to count staff: %w", err)
		}
		if count > 1 {
			return nil, NewBusinessError(
				"TEAM_HAS_STAFF",
				"Komandada basqa isciler var. Evvelce onlari deaktiv edin.",
			)
		}
	}

	if err := service.repository.UpdateType(ctx, businessID, mode); err != nil {
		return nil, fmt.Errorf("failed to update business type: %w", err)
	}

	current.BusinessType = mode
	current.UpdatedAt = time.Now()
	return current, nil
}

func (service *BusinessService) GetBusinessByID(ctx context.Context, id uuid.UUID) (*Business, error) {
	if id == uuid.Nil {
		return nil, NewBusinessError("INVALID_BUSINESS_ID", "Business ID cannot be empty")
	}

	business, err := service.repository.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get business: %w", err)
	}

	if business == nil {
		return nil, NewBusinessError("BUSINESS_NOT_FOUND", "Business not found")
	}

	return business, nil
}

func (service *BusinessService) GetBusinessByOwner(ctx context.Context, ownerID uuid.UUID) (*Business, error) {
	if ownerID == uuid.Nil {
		return nil, NewBusinessError("INVALID_OWNER_ID", "Owner ID cannot be empty")
	}

	business, err := service.repository.GetByOwnerID(ctx, ownerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get business by owner: %w", err)
	}

	if business == nil {
		return nil, NewBusinessError("BUSINESS_NOT_FOUND", "No business found for this owner")
	}

	return business, nil
}
func (service *BusinessService) ListBusinesses(ctx context.Context) ([]*Business, error) {
	businesses, err := service.repository.ListBusinesses(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get business list: %w", err)
	}

	if businesses == nil {
		return nil, NewBusinessError("BUSINESS_NOT_FOUND", "No businesses found")
	}

	return businesses, nil
}

// ListBookableBusinesses – musteri kesfi ucun; yalniz iscisi olan bizneslər.
func (service *BusinessService) ListBookableBusinesses(ctx context.Context) ([]*BookableBusiness, error) {
	businesses, err := service.repository.ListBookable(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list bookable businesses: %w", err)
	}
	return businesses, nil
}
func (service *BusinessService) UpdateBusiness(
	ctx context.Context,
	businessID uuid.UUID,
	ownerID uuid.UUID,
	request *UpdateBusinessRequest,
) error {
	if businessID == uuid.Nil {
		return NewBusinessError("INVALID_BUSINESS_ID", "Business ID cannot be empty")
	}

	if request == nil {
		return NewBusinessError("INVALID_REQUEST", "Request cannot be nil")
	}

	existing, err := service.repository.GetByOwnerID(ctx, ownerID)
	if err != nil {
		return fmt.Errorf("failed to get business: %w", err)
	}
	if existing == nil {
		return NewBusinessError("BUSINESS_NOT_FOUND", "Business not found")
	}

	// Frontdan gələn ID uyğundurmu?
	if existing.ID != businessID {
		return NewBusinessError("UNAUTHORIZED", "Business ID does not match")
	}
	business, err := service.repository.GetByID(ctx, businessID)
	if err != nil {
		return fmt.Errorf("failed to get business: %w", err)
	}

	if business == nil {
		return NewBusinessError("BUSINESS_NOT_FOUND", "Business not found")
	}

	business.Name = request.Name
	business.Industry = request.Industry
	business.Phone = request.Phone
	// business.OwnerID = request.OwnerID
	business.UpdatedAt = time.Now()

	if err := service.validateBusiness(business); err != nil {
		return err
	}

	if err := service.repository.Update(ctx, business); err != nil {
		return fmt.Errorf("failed to update business: %w", err)
	}

	return nil
}
