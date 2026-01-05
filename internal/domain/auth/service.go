package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	repo           AuthRepository
	passwordHasher PasswordHasher
	emailService   EmailService
	tokenManager   TokenManager
}

func NewAuthService(
	repo AuthRepository,
	hasher PasswordHasher,
	email EmailService,
	token TokenManager,
) *Service {
	return &Service{
		repo:           repo,
		passwordHasher: hasher,
		emailService:   email,
		tokenManager:   token,
	}
}

func (s *Service) Register(ctx context.Context, req *RegisterRequest) (*AuthResponse, error) {
	if err := s.validateRegisterRequest(req); err != nil {
		return nil, err
	}
	exists, err := s.repo.EmailExists(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("email exists check failed: %w", err)
	}
	if exists {
		return nil, &RegistrationError{
			Code:    "EMAIL_EXISTS",
			Message: "Email already registered",
		}
	}
	hashedPassword, err := s.passwordHasher.HashPassword(req.Password)
	if err != nil {
		return nil, &RegistrationError{
			Code:    "PASSWORD_HASHING_FAILED",
			Message: "Failed to process password",
		}
	}
	now := time.Now()
	userID := uuid.New()

	user := &User{
		ID:            userID,
		Email:         strings.ToLower(strings.TrimSpace(req.Email)), // Email normalize
		PasswordHash:  hashedPassword,
		FullName:      strings.TrimSpace(req.FullName),
		Phone:         strings.TrimSpace(req.Phone),
		Role:          UserTypeCustomer,
		BusinessID:    nil,
		IsActive:      true,
		IsOwner:       false,
		EmailVerified: false,
		Avatar:        nil,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := s.repo.CreateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	return s.generateAuthResponse(ctx, user)
}

func (s *Service) Login(ctx context.Context, req *LoginRequest) (*AuthResponse, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil || user == nil {
		return nil, &RegistrationError{Code: "INVALID_CREDENTIALS", Message: "Invalid email or password"}
	}
	if err := s.passwordHasher.VerifyPassword(user.PasswordHash, req.Password); err != nil {
		return nil, &RegistrationError{
			Code:    "INVALID_CREDENTIALS",
			Message: "Invalid email or password",
		}
	}
	if !user.IsActive {
		return nil, &RegistrationError{Code: "USER_INACTIVE", Message: "Account is inactive"}
	}
	return s.generateAuthResponse(ctx, user)
}
func (s *Service) RefreshAccessToken(ctx context.Context, plainToken string) (string, error) {
	hashedToken := hashToken(plainToken)
	rt, err := s.repo.GetRefreshToken(ctx, hashedToken)
	if err != nil || rt == nil {
		return "", &RegistrationError{Code: "INVALID_REFRESH_TOKEN", Message: "Invalid refresh token"}
	}
	if time.Now().After(rt.ExpiresAt) {
		return "", &RegistrationError{Code: "REFRESH_TOKEN_EXPIRED", Message: "Refresh token expired"}
	}
	if rt.Revoked {
		return "", &RegistrationError{Code: "REFRESH_TOKEN_REVOKED", Message: "Refresh token revoked"}
	}

	user, err := s.repo.GetUserByID(ctx, rt.UserID)
	if err != nil || user == nil {
		return "", &RegistrationError{Code: "USER_NOT_FOUND", Message: "User not found"}
	}

	claims := &JWTClaims{
		UserID:     user.ID,
		Email:      user.Email,
		Role:       user.Role,
		BusinessID: user.BusinessID,
		IsOwner:    user.IsOwner,
		ExpiresAt:  time.Now().Add(15 * time.Minute).Unix(),
	}

	accessToken, err := s.tokenManager.GenerateAccessToken(claims)
	if err != nil {
		return "", fmt.Errorf("failed to generate access token: %w", err)
	}

	return accessToken, nil
}

func (s *Service) ForgotPassword(ctx context.Context, req *ForgotPasswordRequest) error {
	email := strings.ToLower(strings.TrimSpace(req.Email))
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil || user == nil {
		return nil
	}
	resetToken, err := generateSecureRandomToken(32)
	if err != nil {
		return fmt.Errorf("reset token generation failed: %w", err)
	}
	reset := &PasswordReset{
		ID:        uuid.New(),
		Email:     user.Email,
		Token:     hashToken(resetToken),
		ExpiresAt: time.Now().Add(24 * time.Hour),
		Used:      false,
		CreatedAt: time.Now(),
	}
	if err := s.repo.SavePasswordReset(ctx, reset); err != nil {
		return &RegistrationError{Code: "RESET_TOKEN_SAVE_FAILED", Message: "Failed to save reset token"}
	}
	resetURL := fmt.Sprintf("https://bronet.com/reset-password?token=%s", resetToken)
	if err := s.emailService.SendPasswordResetEmail(user.Email, resetURL); err != nil {
		fmt.Printf("Email send failed for %s: %v", user.Email, err)
	}
	return nil
}
func (s *Service) generateAuthResponse(ctx context.Context, user *User) (*AuthResponse, error) {
	accesClaims := &JWTClaims{
		UserID:     user.ID,
		Email:      user.Email,
		Role:       user.Role,
		BusinessID: user.BusinessID,
		IsOwner:    user.IsOwner,
		ExpiresAt:  time.Now().Add(15 * time.Minute).Unix(),
	}

	accessToken, err := s.tokenManager.GenerateAccessToken(accesClaims)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshTokenString, err := s.tokenManager.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	refreshToken := &RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		Token:     hashToken(refreshTokenString),
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		CreatedAt: time.Now(),
		Revoked:   false,
	}

	if err := s.repo.SaveRefreshToken(ctx, refreshToken); err != nil {
		return nil, fmt.Errorf("failed to save refresh token: %w", err)
	}

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshTokenString,
		User:         user,
		ExpiresIn:    900,
		TokenType:    "Bearer",
	}, nil
}

func (s *Service) ResetPassword(ctx context.Context, req *ResetPasswordRequest) error {
	hashedToken := hashToken(req.Token)
	reset, err := s.repo.GetPasswordReset(ctx, hashedToken)
	if err != nil || reset == nil {
		return &RegistrationError{Code: "INVALID_TOKEN", Message: "Invalid or expired token"}
	}
	now := time.Now()
	if now.After(reset.ExpiresAt) {
		return &RegistrationError{Code: "TOKEN_EXPIRED", Message: "Token expired (24 hours)"}
	}
	if reset.Used {
		return &RegistrationError{Code: "TOKEN_ALREADY_USED", Message: "Token already used"}
	}
	hashedPassword, err := s.passwordHasher.HashPassword(req.Password)
	if err != nil {
		return &RegistrationError{Code: "PASSWORD_HASH_FAILED", Message: "Password processing failed"}
	}
	user, err := s.repo.GetUserByEmail(ctx, reset.Email)
	if err != nil || user == nil {
		return &RegistrationError{Code: "USER_NOT_FOUND", Message: "User not found"}
	}
	if err := s.repo.UpdatePassword(ctx, user.ID.String(), hashedPassword); err != nil {
		return &RegistrationError{Code: "PASSWORD_UPDATE_FAILED", Message: "Failed to update password"}
	}
	reset.Used = true
	reset.UpdatedAt = now
	if err := s.repo.SavePasswordReset(ctx, reset); err != nil {
		fmt.Printf("Reset token revoke failed: %v", err)
	}
	return nil
}
func (s *Service) RevokeRefreshToken(ctx context.Context, plainToken string) error {
	if plainToken == "" {
		return &RegistrationError{
			Code:    "INVALID_REFRESH_TOKEN",
			Message: "Invalid refresh token",
		}
	}

	hashedToken := hashToken(plainToken)

	rt, err := s.repo.GetRefreshToken(ctx, hashedToken)
	if err != nil || rt == nil {
		return &RegistrationError{
			Code:    "INVALID_REFRESH_TOKEN",
			Message: "Invalid refresh token",
		}
	}

	if err := s.repo.RevokeRefreshToken(ctx, rt.ID); err != nil {
		return fmt.Errorf("failed to revoke refresh token: %w", err)
	}

	return nil
}

//	func generateRandomToken(length int) string {
//		return uuid.New().String() + uuid.New().String()
//	}
func generateSecureRandomToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
func hashToken(token string) string {
	h := sha256.New()
	h.Write([]byte(token))
	return fmt.Sprintf("%x", h.Sum(nil))
}
