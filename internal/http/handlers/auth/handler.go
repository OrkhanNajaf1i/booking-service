package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/OrkhanNajaf1i/booking-service/internal/domain/auth"
	"github.com/OrkhanNajaf1i/booking-service/internal/logger"
)

type Handler struct {
	authService *auth.Service
	logger      logger.Logger
}

func NewAuthHandler(authService *auth.Service, logger logger.Logger) *Handler {
	return &Handler{
		authService: authService,
		logger:      logger,
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
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	h.logger.Info("Register request received",
		logger.Field{Key: "method", Value: r.Method},
		logger.Field{Key: "path", Value: r.URL.Path},
		logger.Field{Key: "remote_addr", Value: r.RemoteAddr},
	)
	var httpReq RegisterHTTPRequest
	if err := json.NewDecoder(r.Body).Decode(&httpReq); err != nil {
		h.logger.Error("Failed to decode register request", logger.Field{Key: "error", Value: err.Error()})
		h.sendError(w, http.StatusBadRequest, "VALIDATION_ERROR")
		return
	}
	if httpReq.Email == "" || httpReq.Password == "" || httpReq.FullName == "" || httpReq.Phone == "" {
		h.sendError(w, http.StatusBadRequest, "VALIDATION_ERROR")
		return
	}
	domainReq := ToDomainRegister(&httpReq)
	authResp, err := h.authService.Register(ctx, domainReq)
	var bid string = "null"
	if authResp.User.BusinessID != nil {
		bid = authResp.User.BusinessID.String()
	}
	if err != nil {
		h.logger.Error("Register: Service error",
			logger.Field{Key: "error", Value: err.Error()},
			logger.Field{Key: "email", Value: httpReq.Email},
		)

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
	h.logger.Info("Register: user created successfully",
		logger.Field{Key: "user_id", Value: authResp.User.ID.String()},
		logger.Field{Key: "email", Value: authResp.User.Email},
		logger.Field{Key: "role", Value: string(authResp.User.Role)},
		logger.Field{Key: "business_id", Value: bid},
		logger.Field{Key: "is_owner", Value: authResp.User.IsOwner},
	)
	httpResp := FromDomainAuthResponse(authResp)
	h.sendJSON(w, http.StatusCreated, httpResp)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	h.logger.Info("Login request received",
		logger.Field{Key: "method", Value: r.Method},
		logger.Field{Key: "path", Value: r.URL.Path},
		logger.Field{Key: "remote_addr", Value: r.RemoteAddr},
	)

	var httpReq LoginHTTPRequest
	if err := json.NewDecoder(r.Body).Decode(&httpReq); err != nil {
		h.logger.Error("Login: JSON parse failed",
			logger.Field{Key: "error", Value: err.Error()},
			logger.Field{Key: "remote_addr", Value: r.RemoteAddr},
		)
		h.sendError(w, http.StatusBadRequest, "VALIDATION_ERROR")
		return
	}
	domainReq := &auth.LoginRequest{
		Email:    httpReq.Email,
		Password: httpReq.Password,
	}
	authResponse, err := h.authService.Login(ctx, domainReq)
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
	var bid string = "null"
	if authResponse.User.BusinessID != nil {
		bid = authResponse.User.BusinessID.String()
	}
	h.logger.Info("Login: User authenticated successfully",
		logger.Field{Key: "user_id", Value: authResponse.User.ID.String()},
		logger.Field{Key: "email", Value: authResponse.User.Email},
		logger.Field{Key: "role", Value: string(authResponse.User.Role)},
		logger.Field{Key: "business_id", Value: bid},
	)
	httpRes := FromDomainAuthResponse(authResponse)
	h.sendJSON(w, http.StatusOK, httpRes)
}

func (h *Handler) RefreshAccessToken(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	h.logger.Info("RefreshToken request received",
		logger.Field{Key: "remote_addr", Value: r.RemoteAddr},
	)

	var httpReq RefreshTokenHTTPRequest
	if err := json.NewDecoder(r.Body).Decode(&httpReq); err != nil {
		h.logger.Error("RefreshToken: JSON parse failed",
			logger.Field{Key: "error", Value: err.Error()},
		)
		h.sendError(w, http.StatusBadRequest, "VALIDATION_ERROR")
		return
	}

	accessToken, err := h.authService.RefreshAccessToken(ctx, httpReq.RefreshToken)
	if err != nil {
		if authErr, ok := err.(*auth.RegistrationError); ok {
			switch authErr.Code {
			case "INVALID_REFRESH_TOKEN",
				"REFRESH_TOKEN_EXPIRED",
				"REFRESH_TOKEN_REVOKED",
				"USER_NOT_FOUND":
				h.sendError(w, http.StatusUnauthorized, authErr.Code)
			default:
				h.sendError(w, http.StatusBadRequest, authErr.Code)
			}
			return
		}

		h.logger.Error("RefreshToken: service error",
			logger.Field{Key: "error", Value: err.Error()},
		)
		h.sendError(w, http.StatusInternalServerError, "INTERNAL_ERROR")
		return
	}

	h.logger.Info("RefreshToken: Token refreshed successfully")

	successResponse := SuccessResponseDTO{
		Success: true,
		Data: map[string]interface{}{
			"access_token": accessToken,
			"expires_in":   900,
			"token_type":   "Bearer",
		},
		Message: "Token yeniləndi",
	}
	h.sendJSON(w, http.StatusOK, successResponse)
}

func (h *Handler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	h.logger.Info("ForgotPassword request received", logger.Field{Key: "remote_addr", Value: r.RemoteAddr})
	var httpReq ForgotPasswordHTTPRequest
	if err := json.NewDecoder(r.Body).Decode(&httpReq); err != nil {
		h.logger.Error("ForgotPassword: JSON parse failed",
			logger.Field{Key: "error", Value: err.Error()},
		)
		h.sendError(w, http.StatusBadRequest, "VALIDATION_ERROR")
		return
	}
	domainReq := &auth.ForgotPasswordRequest{
		Email: httpReq.Email,
	}
	err := h.authService.ForgotPassword(ctx, domainReq)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, "INTERNAL_ERROR")
		return
	}
	successResp := SuccessResponseDTO{
		Success: true,
		Message: "Parol sıfırlama linki email-ə göndərildi",
		Data:    nil,
	}
	h.logger.Info("ForgotPassword: Reset email process completed",
		logger.Field{Key: "email", Value: httpReq.Email},
	)
	h.sendJSON(w, http.StatusOK, successResp)
}
func (h *Handler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	h.logger.Info("ResetPassword request received",
		logger.Field{Key: "remote_addr", Value: r.RemoteAddr},
	)

	var httpReq ResetPasswordHTTPRequest
	if err := json.NewDecoder(r.Body).Decode(&httpReq); err != nil {
		h.logger.Error("ResetPassword: JSON parse failed",
			logger.Field{Key: "error", Value: err.Error()},
		)
		h.sendError(w, http.StatusBadRequest, "VALIDATION_ERROR")
		return
	}

	domainReq := &auth.ResetPasswordRequest{
		Token:    httpReq.Token,
		Password: httpReq.Password,
	}

	err := h.authService.ResetPassword(ctx, domainReq)
	if err != nil {
		h.logger.Warn("ResetPassword: Password reset failed",
			logger.Field{Key: "error", Value: err.Error()},
		)

		if authErr, ok := err.(*auth.RegistrationError); ok {
			switch authErr.Code {
			case "INVALID_TOKEN",
				"TOKEN_EXPIRED",
				"TOKEN_ALREADY_USED",
				"PASSWORD_TOO_SHORT",
				"PASSWORD_WEAK",
				"USER_NOT_FOUND":
				h.sendError(w, http.StatusBadRequest, authErr.Code)
			default:
				h.sendError(w, http.StatusBadRequest, authErr.Code)
			}
			return
		}

		h.sendError(w, http.StatusInternalServerError, "INTERNAL_ERROR")
		return
	}

	h.logger.Info("ResetPassword: Password reset successful")

	successResp := SuccessResponseDTO{
		Success: true,
		Message: "Parol uğurla sıfırlandı",
		Data:    nil,
	}
	h.sendJSON(w, http.StatusOK, successResp)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	var req RefreshTokenHTTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "VALIDATION_ERROR")
		return
	}

	if req.RefreshToken == "" {
		h.sendError(w, http.StatusBadRequest, "VALIDATION_ERROR")
		return
	}

	if err := h.authService.RevokeRefreshToken(ctx, req.RefreshToken); err != nil {
		if authErr, ok := err.(*auth.RegistrationError); ok {
			switch authErr.Code {
			case "INVALID_REFRESH_TOKEN":
				h.sendError(w, http.StatusBadRequest, authErr.Code)
			default:
				h.sendError(w, http.StatusBadRequest, authErr.Code)
			}
			return
		}

		h.logger.Error("Failed to revoke token",
			logger.Field{Key: "error", Value: err.Error()},
		)
		h.sendError(w, http.StatusInternalServerError, "INTERNAL_ERROR")
		return
	}

	success := SuccessResponseDTO{
		Success: true,
		Message: "Çıxış edildi",
		Data:    nil,
	}
	h.sendJSON(w, http.StatusOK, success)
}
