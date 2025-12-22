package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/OrkhanNajaf1i/booking-service/internal/domain/business"
)

type BusinessHandler struct {
	svc *business.Service
}

func NewBusinessHandler(svc *business.Service) *BusinessHandler {
	return &BusinessHandler{svc: svc}
}

type createBusinessRequest struct {
	Name  string `json:"name"`
	Phone string `json:"phone"`
}

func (h *BusinessHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createBusinessRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	b, err := h.svc.CreateBusiness(r.Context(), req.Name, req.Phone)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(b)
}

func (h *BusinessHandler) GetBusinessByID(w http.ResponseWriter, r *http.Request) {

}
