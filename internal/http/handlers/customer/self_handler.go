// File: internal/http/handlers/customer/self_handler.go
package customer

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/OrkhanNajaf1i/booking-service/internal/http/middleware"
	"github.com/OrkhanNajaf1i/booking-service/internal/logger"
	"github.com/google/uuid"
)

// ResolveSelfRequest – musteri tetbiqi hansi biznesde bron edeceyini bildirir.
type ResolveSelfRequest struct {
	BusinessID string `json:"business_id"`
}

// ResolveSelf – POST /api/v1/customers/self
//
// Musteri tetbiqinde JWT-de yalniz user_id olur, booking ise customer_id
// teleb edir. Bu endpoint istifadecinin hemin biznesdeki musteri kartini
// tapir, yoxdursa profil melumati ile yaradir.
//
// @Summary      Oz musteri kartini tap/yarat
// @Description  Login olmus istifadecinin verilmis biznesdeki musteri kartini qaytarir. Kart yoxdursa yaradilir.
// @Tags         Customer
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body ResolveSelfRequest true "Biznes ID"
// @Success      200 {object} SuccessResponseDTO
// @Failure      400 {object} ErrorResponseDTO
// @Router       /customers/self [post]
func (h *Handler) ResolveSelf(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	userID, err := middleware.UserIDFrom(r)
	if err != nil {
		h.sendError(w, http.StatusUnauthorized, "UNAUTHORIZED")
		return
	}

	var req ResolveSelfRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "INVALID_REQUEST")
		return
	}

	businessID, err := uuid.Parse(req.BusinessID)
	if err != nil || businessID == uuid.Nil {
		// Provider rolundaki istifadeci ucun JWT-deki biznes fallback olur.
		businessID = middleware.OptionalBusinessIDFrom(r)
		if businessID == uuid.Nil {
			h.sendError(w, http.StatusBadRequest, "INVALID_REQUEST")
			return
		}
	}

	found, err := h.service.ResolveSelf(ctx, businessID, userID)
	if err != nil {
		h.logger.Warn("ResolveSelf ugursuz",
			logger.Field{Key: "user_id", Value: userID.String()},
			logger.Field{Key: "business_id", Value: businessID.String()},
			logger.Field{Key: "error", Value: err.Error()},
		)
		h.sendError(w, http.StatusBadRequest, "CUSTOMER_NOT_FOUND")
		return
	}

	h.sendSuccess(w, http.StatusOK, "Musteri karti hazirdir", found)
}
