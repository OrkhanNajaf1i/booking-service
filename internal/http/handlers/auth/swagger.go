// File: internal/http/handlers/auth/swagger.go
package auth

import "net/http"

type AuthSwagger interface {

	// @Summary      İstifadəçi qeydiyyatı
	// @Description  Yeni istifadəçi hesabı yaradır (account-first).
	// @Tags         Auth
	// @Accept       json
	// @Produce      json
	// @Param        request body RegisterHTTPRequest true "Qeydiyyat məlumatları"
	// @Success      201  {object}  AuthResponseDTO
	// @Failure      400  {object}  ErrorResponseDTO
	// @Failure      409  {object}  ErrorResponseDTO
	// @Router       /auth/register [post]
	Register(w http.ResponseWriter, r *http.Request)

	// @Summary      Giriş (Login)
	// @Description  Email və şifrə ilə giriş edərək JWT tokenlərini alır.
	// @Tags         Auth
	// @Accept       json
	// @Produce      json
	// @Param        request body LoginHTTPRequest true "Giriş məlumatları"
	// @Success      200  {object}  AuthResponseDTO
	// @Failure      401  {object}  ErrorResponseDTO
	// @Router       /auth/login [post]
	Login(w http.ResponseWriter, r *http.Request)

	// @Summary      Token Yeniləmə (Refresh)
	// @Description  Refresh token vasitəsilə yeni access token alır.
	// @Tags         Auth
	// @Accept       json
	// @Produce      json
	// @Param        request body RefreshTokenHTTPRequest true "Refresh token"
	// @Success      200  {object}  SuccessResponseDTO
	// @Failure      401  {object}  ErrorResponseDTO
	// @Router       /auth/refresh [post]
	RefreshAccessToken(w http.ResponseWriter, r *http.Request)

	// @Summary      Çıxış (Logout)
	// @Description  Refresh tokeni ləğv edərək sistemdən çıxış edir.
	// @Tags         Auth
	// @Accept       json
	// @Produce      json
	// @Param        request body RefreshTokenHTTPRequest true "Refresh token"
	// @Success      200  {object}  SuccessResponseDTO
	// @Router       /auth/logout [post]
	Logout(w http.ResponseWriter, r *http.Request)
}
