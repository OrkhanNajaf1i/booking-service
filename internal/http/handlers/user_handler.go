package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/OrkhanNajaf1i/booking-service/internal/domain/user"
	domainuser "github.com/OrkhanNajaf1i/booking-service/internal/domain/user"
	"github.com/google/uuid"
)

type UserHandler struct {
	svc *domainuser.Service
}

func NewUserHandler(svc *user.Service) *UserHandler {
	return &UserHandler{svc: svc}
}

type createUserRequest struct {
	BusinessID string `json:"business_id"`
	Name       string `json:"name"`
	Phone      string `json:"phone"`
}

func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	bID, err := uuid.Parse(req.BusinessID)
	if err != nil {
		http.Error(w, "invalid business_id", http.StatusBadRequest)
		return
	}
	u, err := h.svc.CreateUser(r.Context(), bID, req.Name, req.Phone)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(u)
}
func (h *UserHandler) GetByPhone(w http.ResponseWriter, r *http.Request) {
	phone := r.URL.Query().Get("phone")
	if phone == "" {
		http.Error(w, "phone is required", http.StatusBadRequest)
	}
	u, err := h.svc.GetUserByPhone(r.Context(), phone)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(u)
}
