// File: internal/http/handlers/auth/phone_handler.go
//
// Telefonla giris: kod istə → kodu tesdiqle.
//
// Bu axin musteri tetbiqi ucundur. Admin panel e-poct/sifre ile
// qalir (bax: CLAUDE.md, rol ayrimi).
package auth

import (
	"encoding/json"
	"net/http"

	authDomain "github.com/OrkhanNajaf1i/booking-service/internal/domain/auth"
	"github.com/OrkhanNajaf1i/booking-service/internal/domain/otp"
)

// PhoneAuthHandler – kod istəmə ve tesdiq.
type PhoneAuthHandler struct {
	otp  *otp.Service
	auth *authDomain.Service
}

func NewPhoneAuthHandler(otpService *otp.Service, authService *authDomain.Service) *PhoneAuthHandler {
	return &PhoneAuthHandler{otp: otpService, auth: authService}
}

type requestCodeRequest struct {
	Phone string `json:"phone"`
}

type requestCodeResponse struct {
	Phone       string `json:"phone"`
	MaskedPhone string `json:"masked_phone"`
	ExpiresIn   int    `json:"expires_in"`
	ResendAfter int    `json:"resend_after"`
	Channel     string `json:"channel"`
	// DebugCode YALNIZ inkisaf rejiminde (log kanali) dolur.
	DebugCode string `json:"debug_code,omitempty"`
}

type verifyCodeRequest struct {
	Phone string `json:"phone"`
	Code  string `json:"code"`
	// FullName yalniz hesab ilk defe yaradilanda islenir.
	FullName string `json:"full_name"`
}

// RequestCode – POST /api/v1/auth/phone/request
//
// @Summary      Telefona təsdiq kodu göndərir
// @Description  Nömrə normallaşdırılır, 6 rəqəmli kod göndərilir. Kod 5 dəqiqə etibarlıdır.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body requestCodeRequest true "Telefon nömrəsi"
// @Success      200 {object} requestCodeResponse
// @Failure      400 {object} ErrorResponse "Nömrə yanlışdır"
// @Failure      429 {object} ErrorResponse "Çox sayda sorğu"
// @Router       /auth/phone/request [post]
func (h *PhoneAuthHandler) RequestCode(w http.ResponseWriter, r *http.Request) {
	var request requestCodeRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writePhoneError(w, http.StatusBadRequest, "INVALID_BODY", "Sorğu oxunmadı")
		return
	}

	result, err := h.otp.RequestCode(r.Context(), request.Phone)
	if err != nil {
		writeOTPError(w, err)
		return
	}

	writePhoneJSON(w, http.StatusOK, requestCodeResponse{
		Phone:       result.PhoneE164,
		MaskedPhone: result.MaskedPhone,
		ExpiresIn:   result.ExpiresIn,
		ResendAfter: result.ResendAfter,
		Channel:     string(result.Channel),
		DebugCode:   result.DebugCode,
	})
}

// VerifyCode – POST /api/v1/auth/phone/verify
//
// @Summary      Kodu təsdiqləyir və token verir
// @Description  Kod düzgündürsə hesab tapılır, yoxdursa yaradılır və giriş tokenləri qaytarılır.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body verifyCodeRequest true "Nömrə və kod"
// @Success      200 {object} authDomain.AuthResponse
// @Failure      400 {object} ErrorResponse "Kod yanlış və ya vaxtı bitib"
// @Router       /auth/phone/verify [post]
func (h *PhoneAuthHandler) VerifyCode(w http.ResponseWriter, r *http.Request) {
	var request verifyCodeRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writePhoneError(w, http.StatusBadRequest, "INVALID_BODY", "Sorğu oxunmadı")
		return
	}

	phone, err := h.otp.VerifyCode(r.Context(), request.Phone, request.Code)
	if err != nil {
		writeOTPError(w, err)
		return
	}

	response, err := h.auth.LoginWithPhone(r.Context(), phone, request.FullName)
	if err != nil {
		if authErr, ok := err.(*authDomain.RegistrationError); ok {
			status := http.StatusBadRequest
			if authErr.Code == "PHONE_LOGIN_NOT_ALLOWED" {
				status = http.StatusForbidden
			}
			writePhoneError(w, status, authErr.Code, authErr.Message)
			return
		}
		writePhoneError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Giriş alınmadı")
		return
	}

	writePhoneJSON(w, http.StatusOK, response)
}

// ── Cavab köməkçiləri ────────────────────────────────────────

func writeOTPError(w http.ResponseWriter, err error) {
	otpErr, ok := otp.AsError(err)
	if !ok {
		writePhoneError(w, http.StatusInternalServerError, "INTERNAL_ERROR",
			"Əməliyyat alınmadı")
		return
	}

	// Limit xetalari 429 olmalidir ki, tetbiq onlari ayird edib
	// geri sayim gostere bilsin.
	status := http.StatusBadRequest
	switch otpErr.Code {
	case "TOO_MANY_REQUESTS", "RESEND_TOO_SOON", "TOO_MANY_ATTEMPTS":
		status = http.StatusTooManyRequests
	}

	writePhoneError(w, status, otpErr.Code, otpErr.Message)
}

func writePhoneJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writePhoneError(w http.ResponseWriter, status int, code, message string) {
	writePhoneJSON(w, status, map[string]any{
		"success": false,
		"code":    code,
		"message": message,
	})
}
