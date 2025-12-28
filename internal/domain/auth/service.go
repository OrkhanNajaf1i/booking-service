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
	repo            Repository
	passwordHasher  PasswordHasher
	emailService    EmailService
	tokenManager    TokenManager
	businessService BusinessService
}

func NewAuthService(
	repo Repository,
	hasher PasswordHasher,
	email EmailService,
	token TokenManager,
	business BusinessService,
) *Service {
	return &Service{
		repo:            repo,
		passwordHasher:  hasher,
		emailService:    email,
		tokenManager:    token,
		businessService: business,
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
	userID := uuid.New()
	var businessID uuid.UUID
	var locationID uuid.UUID

	switch req.Role {
	case UserTypeSoloPractitioner:
		businessIDStr, err := s.businessService.CreateSoloPractitionerBusiness(
			ctx,
			userID.String(),
			req.BusinessName,
			req.ServiceCategory,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create solo business: %w", err)
		}
		businessID, _ = uuid.Parse(businessIDStr)
		locationIDStr, err := s.businessService.CreateDefaultLocation(ctx, businessID)
		if err != nil {
			return nil, fmt.Errorf("failed to create solo location: %w", err)
		}
		locationID, _ = uuid.Parse(locationIDStr)

	case UserTypeOwner:
		businessIDStr, err := s.businessService.CreateMultiStaffBusiness(
			ctx,
			userID.String(),
			req.BusinessName,
			req.Industry,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create multi business: %w", err)
		}
		businessID, _ = uuid.Parse(businessIDStr)

		locationIDStr, err := s.businessService.CreateDefaultLocation(ctx, businessID)
		if err != nil {
			return nil, fmt.Errorf("failed to create multi location: %w", err)
		}
		locationID, _ = uuid.Parse(locationIDStr)

	case UserTypeCustomer:
		businessID = uuid.Nil
		locationID = uuid.Nil

	default:
		return nil, &RegistrationError{
			Code:    "INVALID_ROLE",
			Message: "Invalid user role",
		}
	}

	now := time.Now()
	user := &User{
		ID:            userID,
		Email:         strings.ToLower(strings.TrimSpace(req.Email)),
		PasswordHash:  hashedPassword,
		FullName:      req.FullName,
		Phone:         req.Phone,
		Role:          req.Role,
		BusinessID:    businessID,
		IsActive:      true,
		IsOwner:       req.Role == UserTypeSoloPractitioner || req.Role == UserTypeOwner,
		EmailVerified: false,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := s.repo.CreateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	if user.Role == UserTypeSoloPractitioner || user.Role == UserTypeOwner {
		profile := &StaffProfile{
			ID:         uuid.New(),
			UserID:     user.ID,
			BusinessID: user.BusinessID,
			LocationID: &locationID,
			Role:       StaffRoleAdministrator,
			Title:      "",
			Department: "",
			Bio:        "",
			HourlyRate: 0,
			Status:     "active",
			JoinedAt:   now,
			CreatedAt:  now,
			UpdatedAt:  now,
		}

		if err := s.repo.CreateStaffProfile(ctx, profile); err != nil {
			return nil, fmt.Errorf("failed to create staff profile: %w", err)
		}
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

func generateRandomToken(length int) string {
	return uuid.New().String() + uuid.New().String()
}
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

// func (s *Service) handleSoloRegistration(ctx context.Context, req *RegisterRequest) (*AuthResponse, error) {
// 	bID := uuid.New()
// 	uID := uuid.New()
// }

// func (s *service) registerSoloPractitioner(ctx context.Context, email, password, businessName, locationName string, locationAddress *string, locationCity *string) (*User, error) {
// 	businessID := s.idGenerator.Generate()
// 	userID := s.idGenerator.Generate()
// 	locationID := s.idGenerator.Generate()
// 	staffID := s.idGenerator.Generate()

// 	passwordHash, err := s.passwordHasher.Hash(password)
// 	if err != nil {
// 		return nil, &RegistrationError{
// 			Code:    "PASSWORD_HASH_FAILED",
// 			Message: "Failed to process password. Please to try again",
// 		}
// 	}
// 	business := &Business{
// 		ID:           businessID,
// 		Name:         businessName,
// 		OwnerID:      userID,
// 		BusinessType: BusinessTypeSolo,
// 		IsActive:     true,
// 		CreatedAt:    timeNow(),
// 		UpdatedAt:    timeNow(),
// 	}
// 	user := &User{
// 		ID:           userID,
// 		Email:        email,
// 		PasswordHash: passwordHash,
// 		Role:         RoleSoloPractitioner,
// 		BusinessID:   businessID,
// 		IsActive:     true,
// 		UpdatedAt:    timeNow(),
// 		CreatedAt:    timeNow(),
// 	}
// 	location := &Location{
// 		ID:         locationID,
// 		BusinessID: businessID,
// 		Name:       locationName,
// 		Address:    locationAddress,
// 		City:       locationCity,
// 		IsActive:   true,
// 		CreatedAt:  timeNow(),
// 	}
// 	staff := &Staff{
// 		ID:         staffID,
// 		BusinessID: businessID,
// 		UserID:     userID,
// 		Position:   "Owner",
// 		IsActive:   true,
// 		CreatedAt:  timeNow(),
// 	}
// 	if err := s.businessRepo.CreateBusiness(ctx, business); err != nil {
// 		return nil, &RegistrationError{
// 			Code:    "BUSINESS_CREATE_FAILED",
// 			Message: "Failed to create business. Please try again.",
// 		}
// 	}

// 	if err := s.userRepo.CreateUser(ctx, user); err != nil {
// 		return nil, &RegistrationError{
// 			Code:    "USER_CREATE_FAILED",
// 			Message: "Failed to create user account. Please try again",
// 		}
// 	}
// 	if err := s.locationRepo.CreateLocation(ctx, location); err != nil {
// 		return nil, &RegistrationError{
// 			Code:    "LOCATION_CREATE_FAILED",
// 			Message: "Failed to create location. Please try again",
// 		}
// 	}
// 	if err := s.staffRepo.CreateStaff(ctx, staff); err != nil {
// 		return nil, &RegistrationError{
// 			Code:    "STAFF_CREATE_FAILED",
// 			Message: "Failed to set up staff record. Please try again",
// 		}
// 	}
// 	return user, nil
// }
// func (s *service) registerBusinessOwner(ctx context.Context, email, password, businessName string) (*User, error) {
// 	businessID := s.idGenerator.Generate()
// 	userID := s.idGenerator.Generate()

// 	passwordHash, err := s.passwordHasher.Hash(password)
// 	if err != nil {
// 		return nil, &RegistrationError{
// 			Code:    "PASSWORD_HASH_FAILED",
// 			Message: "Failed to process password. Please try again",
// 		}
// 	}
// 	business := &Business{
// 		ID:           businessID,
// 		Name:         businessName,
// 		OwnerID:      userID,
// 		BusinessType: BusinessTypeMulti,
// 		IsActive:     true,
// 		CreatedAt:    timeNow(),
// 		UpdatedAt:    timeNow(),
// 	}
// 	user := &User{
// 		ID:           userID,
// 		Email:        email,
// 		PasswordHash: passwordHash,
// 		Role:         RoleProviderOwner,
// 		BusinessID:   businessID,
// 		IsActive:     true,
// 		CreatedAt:    timeNow(),
// 		UpdatedAt:    timeNow(),
// 	}

// 	if err := s.businessRepo.CreateBusiness(ctx, business); err != nil {
// 		return nil, &RegistrationError{
// 			Code:    "BUSINESS_CREATE_FAILED",
// 			Message: "Failed to create business. Please try again",
// 		}
// 	}

// 	if err := s.userRepo.CreateUser(ctx, user); err != nil {
// 		return nil, &RegistrationError{
// 			Code:    "USER_CREATE_FAILED",
// 			Message: "Failed to create user account. Please try again",
// 		}
// 	}
// 	return user, nil
// }

// func (s *service) GetUserRole(ctx context.Context, userID uuid.UUID) (string, error) {
// 	user, err := s.userRepo.GetUserByID(ctx, userID)
// 	if err != nil {
// 		return "", err
// 	}
// 	return string(user.Role), nil
// }
// func timeNow() time.Time {
// 	return time.Now().UTC()
// }
