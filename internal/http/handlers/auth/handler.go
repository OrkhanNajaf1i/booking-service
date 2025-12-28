package auth

import (
	"encoding/json"
	"net/http"

	"github.com/OrkhanNajaf1i/booking-service/internal/domain/auth"
)

type Handler struct {
	authService *auth.Service
}

func NewAuthHandler(authService *auth.Service) *Handler {
	return &Handler{
		authService: authService,
	}
}
func (h *Handler) sendJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
func (h *Handler) sendError(w http.ResponseWriter, status int, code string) {
	errorDTO := GetErrorResponse(code)
	h.sendJSON(w, status, errorDTO)
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var httpReq RegisterHTTPRequest
	if err := json.NewDecoder(r.Body).Decode(&httpReq); err != nil {
		h.sendError(w, http.StatusBadRequest, "VALIDATION_ERROR")
		return
	}
	domainReq := ToDomainRegister(&httpReq)
	authResp, err := h.authService.Register(r.Context(), domainReq)
	if err != nil {
		if regErr, ok := err.(*auth.RegistrationError); ok {
			switch regErr.Code {
			case "EMAIL_EXISTS":
				h.sendError(w, http.StatusConflict, regErr.Code)
			case "PASSWORD_TOO_SHORT", "INVALID_EMAIL_FORMAT", "FULLNAME_REQUIRED":
				h.sendError(w, http.StatusBadRequest, regErr.Code)
			default:
				h.sendError(w, http.StatusBadRequest, regErr.Code)
			}
			return
		}
		h.sendError(w, http.StatusInternalServerError, "INTERNAL_ERROR")
		return
	}
	httpResp := FromDomainAuthResponse(authResp)
	h.sendJSON(w, http.StatusCreated, httpResp)
}
func (h *Handler) Login(w http.ResponseWriter, r http.Request) {
	var httpReq LoginHTTPRequest
	if err := json.NewDecoder(r.Body).Decode(&httpReq); err != nil {
		h.sendError(w, http.StatusBadRequest, "VALIDATION_ERROR")
		return
	}
	domainReq := &auth.LoginRequest{
		Email:    httpReq.Email,
		Password: httpReq.Password,
	}
	authRes, err := h.authService.Login(r.Context(), domainReq)
	if err != nil {
		if authErr, ok := err.(*auth.RegistrationError); ok {
			switch authErr.Code {
			case "INVALID_CREDENTIALS":
				h.sendError(w, http.StatusUnauthorized, authErr.Code) // 401
			case "USER_INACTIVE":
				h.sendError(w, http.StatusForbidden, authErr.Code) // 403
			default:
				h.sendError(w, http.StatusBadRequest, authErr.Code)
			}
			return
		}
		h.sendError(w, http.StatusInternalServerError, "INTERNAL_ERROR")
		return
	}
	httpRes := FromDomainAuthResponse(authRes)
	h.sendJSON(w, http.StatusOK, httpRes)
}
