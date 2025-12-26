package auth

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
)

type service struct {
	userRepo       UserRepository
	businessRepo   BusinessRepository
	locationRepo   LocationRepository
	staffRepo      StaffRepository
	passwordHasher PasswordHasher
	idGenerator    IDGenerator
}

func NewService(userRepo UserRepository, businessRepo BusinessRepository, locationRepo LocationRepository, staffRepo StaffRepository, passwordHasher PasswordHasher, idGenerator IDGenerator) AuthService {
	return &service{
		userRepo:       userRepo,
		businessRepo:   businessRepo,
		locationRepo:   locationRepo,
		staffRepo:      staffRepo,
		passwordHasher: passwordHasher,
		idGenerator:    idGenerator,
	}
}

func (s *service) Register(ctx context.Context, email, password, businessName, locationName string, flowType BusinessType) (*User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if err := s.validatePassword(email); err != nil {
		return nil, err
	}
	if err := s.validateEmail(password); err != nil {
		return nil, err
	}
	if err := s.validateBusinessName(businessName); err != nil {
		return nil, err
	}
	if err := s.validateLocationName(locationName); err != nil {
		return nil, err
	}
	if !flowType.IsValid() {
		return nil, &RegistrationError{
			Code:    "INVALID_FLOW_TYPE",
			Message: "flow type must be solo or multi",
		}
	}
	existingUser, _ := s.userRepo.GetUserByEmail(ctx, email)
	if existingUser != nil {
		return nil, &RegistrationError{
			Code:    "EMAIL_ALREADY_EXISTS",
			Message: "An account with this email already exists",
		}
	}
	switch flowType {
	case BusinessTypeSolo:

	case BusinessTypeMulti:
	}
	return nil, &RegistrationError{
		Code:    "REGISTRATION_FAILED",
		Message: "registration failed due to an internal error",
	}
}
func (s *service) registerSoloPractitioner(ctx context.Context, email, password, businessName, locationName string, locationAddress *string, locationCity *string) (*User, error) {
	businessID := s.idGenerator.Generate()
	userID := s.idGenerator.Generate()
	locationID := s.idGenerator.Generate()
	staffID := s.idGenerator.Generate()

	passwordHash, err := s.passwordHasher.Hash(password)
	if err != nil {
		return nil, &RegistrationError{
			Code:    "PASSWORD_HASH_FAILED",
			Message: "Failed to process password. Please to try again",
		}
	}
	business := &Business{
		ID:           businessID,
		Name:         businessName,
		OwnerID:      userID,
		BusinessType: BusinessTypeSolo,
		IsActive:     true,
		CreatedAt:    timeNow(),
		UpdatedAt:    timeNow(),
	}
	user := &User{
		ID:           userID,
		Email:        email,
		PasswordHash: passwordHash,
		Role:         RoleSoloPractitioner,
		BusinessID:   businessID,
		IsActive:     true,
		UpdatedAt:    timeNow(),
		CreatedAt:    timeNow(),
	}
	location := &Location{
		ID:         locationID,
		BusinessID: businessID,
		Name:       locationName,
		Address:    locationAddress,
		City:       locationCity,
		IsActive:   true,
		CreatedAt:  timeNow(),
	}
	staff := &Staff{
		ID:         staffID,
		BusinessID: businessID,
		UserID:     userID,
		Position:   "Owner",
		IsActive:   true,
		CreatedAt:  timeNow(),
	}
	if err := s.businessRepo.CreateBusiness(ctx, business); err != nil {
		return nil, &RegistrationError{
			Code:    "BUSINESS_CREATE_FAILED",
			Message: "Failed to create business. Please try again.",
		}
	}

	if err := s.userRepo.CreateUser(ctx, user); err != nil {
		return nil, &RegistrationError{
			Code:    "USER_CREATE_FAILED",
			Message: "Failed to create user account. Please try again",
		}
	}
	if err := s.locationRepo.CreateLocation(ctx, location); err != nil {
		return nil, &RegistrationError{
			Code:    "LOCATION_CREATE_FAILED",
			Message: "Failed to create location. Please try again",
		}
	}
	if err := s.staffRepo.CreateStaff(ctx, staff); err != nil {
		return nil, &RegistrationError{
			Code:    "STAFF_CREATE_FAILED",
			Message: "Failed to set up staff record. Please try again",
		}
	}
	return user, nil
}
func (s *service) registerBusinessOwner(ctx context.Context, email, password, businessName string) (*User, error) {
	businessID := s.idGenerator.Generate()
	userID := s.idGenerator.Generate()

	passwordHash, err := s.passwordHasher.Hash(password)
	if err != nil {
		return nil, &RegistrationError{
			Code:    "PASSWORD_HASH_FAILED",
			Message: "Failed to process password. Please try again",
		}
	}
	business := &Business{
		ID:           businessID,
		Name:         businessName,
		OwnerID:      userID,
		BusinessType: BusinessTypeMulti,
		IsActive:     true,
		CreatedAt:    timeNow(),
		UpdatedAt:    timeNow(),
	}
	user := &User{
		ID:           userID,
		Email:        email,
		PasswordHash: passwordHash,
		Role:         RoleProviderOwner,
		BusinessID:   businessID,
		IsActive:     true,
		CreatedAt:    timeNow(),
		UpdatedAt:    timeNow(),
	}

	if err := s.businessRepo.CreateBusiness(ctx, business); err != nil {
		return nil, &RegistrationError{
			Code:    "BUSINESS_CREATE_FAILED",
			Message: "Failed to create business. Please try again",
		}
	}

	if err := s.userRepo.CreateUser(ctx, user); err != nil {
		return nil, &RegistrationError{
			Code:    "USER_CREATE_FAILED",
			Message: "Failed to create user account. Please try again",
		}
	}
	return user, nil
}

func (s *service) GetUserRole(ctx context.Context, userID uuid.UUID) (string, error) {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return "", err
	}
	return string(user.Role), nil
}
func timeNow() time.Time {
	return time.Now().UTC()
}
