package business

import (
	"context"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	repo        Repository
	userService UserService
}

func NewService(repo Repository, userService UserService) *Service {
	return &Service{
		repo:        repo,
		userService: userService,
	}
}
func (s *Service) CreateSoloBusiness(ctx context.Context, ownerID uuid.UUID, request *CreateBusinessRequest) (uuid.UUID, error) {
	if ownerID == uuid.Nil {
		return uuid.Nil, fmt.Errorf("OwnerID cannot be empty")
	}
	if request == nil {
		return uuid.Nil, fmt.Errorf("request cannot be nil")
	}
	if err := s.ValidateCreateBusinessRequest(request, BusinessTypeSolo); err != nil {
		return uuid.Nil, fmt.Errorf("request validation failed: %w", err)
	}
	business := &Business{
		ID:              uuid.New(),
		Name:            request.Name,
		OwnerID:         ownerID,
		Industry:        request.Industry,
		ServiceCategory: request.ServiceCategory,
		Phone:           request.Phone,
		BusinessType:    BusinessTypeSolo,
		IsActive:        true,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	if err := s.ValidateBusiness(business); err != nil {
		return uuid.Nil, fmt.Errorf("business validation failed: %w", err)
	}
	if err := s.repo.CreateBusiness(ctx, business); err != nil {
		return uuid.Nil, fmt.Errorf("failed to create business in DB: %w", err)
	}
	if err := s.userService.UpdateUserBusinessID(ctx, ownerID, business.ID, true); err != nil {
		return uuid.Nil, fmt.Errorf("failed to assign owner to business: %w", err)
	}
	locationID, err := s.CreateDefaultLocation(ctx, business.ID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to create default location: %w", err)
	}
	profile := &StaffProfile{
		ID:         uuid.New(),
		UserID:     ownerID,
		BusinessID: business.ID,
		LocationID: &locationID,
		Role:       StaffRoleAdmin,
		Title:      "Owner",
		Status:     "active",
		JoinedAt:   time.Now(),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	if err := s.ValidateStaffProfile(profile); err != nil {
		return uuid.Nil, fmt.Errorf("staff profile validation failed: %w", err)
	}
	if err := s.repo.CreateStaffProfile(ctx, profile); err != nil {
		return uuid.Nil, fmt.Errorf("failed to create owner staff profile: %w", err)
	}
	return business.ID, nil
}
func (s *Service) CreateMultiBusiness(ctx context.Context, ownerID uuid.UUID, request *CreateBusinessRequest) (uuid.UUID, error) {
	if ownerID == uuid.Nil {
		return uuid.Nil, fmt.Errorf("ownerID cannot be empty")
	}
	if request == nil {
		return uuid.Nil, fmt.Errorf("request cannot be nil")
	}
	if err := s.ValidateCreateBusinessRequest(request, BusinessTypeMulti); err != nil {
		return uuid.Nil, fmt.Errorf("request validation failed: %w", err)
	}
	business := &Business{
		ID:              uuid.New(),
		Name:            request.Name,
		OwnerID:         ownerID,
		Industry:        request.Industry,
		ServiceCategory: request.ServiceCategory,
		Phone:           request.Phone,
		BusinessType:    BusinessTypeMulti,
		IsActive:        true,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	if err := s.ValidateBusiness(business); err != nil {
		return uuid.Nil, fmt.Errorf("business validation failed: %w", err)
	}
	if err := s.repo.CreateBusiness(ctx, business); err != nil {
		return uuid.Nil, fmt.Errorf("failed to create business in DB: %w", err)
	}
	if err := s.repo.CreateBusiness(ctx, business); err != nil {
		return uuid.Nil, fmt.Errorf("failed to create business in DB: %w", err)
	}
	if err := s.userService.UpdateUserBusinessID(ctx, ownerID, business.ID, true); err != nil {
		return uuid.Nil, fmt.Errorf("failed to assign owner to business: %w", err)
	}
	profile := &StaffProfile{
		ID:         uuid.New(),
		UserID:     ownerID,
		BusinessID: business.ID,
		Role:       StaffRoleAdmin,
		Title:      "Owner",
		Status:     "active",
		JoinedAt:   time.Now(),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	if err := s.ValidateStaffProfile(profile); err != nil {
		return uuid.Nil, fmt.Errorf("staff profile validation failed: %w", err)
	}
	if err := s.repo.CreateStaffProfile(ctx, profile); err != nil {
		return uuid.Nil, fmt.Errorf("failed to create owner staff profile: %w", err)
	}
	return business.ID, nil
}
func (s *Service) InviteStaff(ctx context.Context, businessID uuid.UUID, request *InviteStaffRequest) (string, error) {
	if businessID == uuid.Nil {
		return "", fmt.Errorf("businessID cannot be empty")
	}
	if request == nil {
		return "", fmt.Errorf("request cannot be nil")
	}
	if err := s.ValidateInviteStaffRequest(request); err != nil {
		return "", fmt.Errorf("invite request validation failed: %w", err)
	}
	tokenBytes := make([]byte, 32)
	token := hex.EncodeToString(tokenBytes)
	invite := &BusinessInvite{
		ID:           uuid.New(),
		BusinessID:   businessID,
		InvitedEmail: request.Email,
		InvitedPhone: request.Phone,
		Role:         request.Role,
		LocationID:   request.LocationID,
		Token:        token,
		ExpiresAt:    time.Now().Add(7 * 24 * time.Hour),
		Used:         false,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if err := s.ValidateBusinessInvite(invite); err != nil {
		return "", fmt.Errorf("invite validaton failed: %w", err)
	}
	if err := s.repo.CreateInvite(ctx, invite); err != nil {
		return "", fmt.Errorf("failed to save invite to DB: %w", err)
	}
	return token, nil
}
func (s *Service) JoinWithInvite(ctx context.Context, token, password string) error {
	if token == "" {
		return fmt.Errorf("invite token is required")
	}
	userIDStr, ok := ctx.Value("user_id").(string)
	if !ok || userIDStr == "" {
		return fmt.Errorf("user_id missing in context (JWT required)")
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return fmt.Errorf("invalid user_id in context: %w", err)
	}
	invite, err := s.repo.GetInviteByToken(ctx, token, uuid.Nil)
	if err != nil {
		return fmt.Errorf("failed to retrieve invite: %w", err)
	}
	if invite == nil {
		return fmt.Errorf("invite not found")
	}
	if invite.Used {
		return fmt.Errorf("invite has already been used")
	}
	if time.Now().After(invite.ExpiresAt) {
		return fmt.Errorf("invite token has expired")
	}
	if err := s.userService.UpdateUserBusinessID(ctx, userID, invite.BusinessID, false); err != nil {
		return fmt.Errorf("failed to link user to business: %w", err)
	}
	profile := &StaffProfile{
		ID:         uuid.New(),
		UserID:     userID,
		BusinessID: invite.BusinessID,
		LocationID: invite.LocationID,
		Role:       invite.Role,
		Title:      "Staff Member",
		Status:     "active",
		JoinedAt:   time.Now(),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := s.ValidateStaffProfile(profile); err != nil {
		return fmt.Errorf("staff profile validation failed: %w", err)
	}
	if err := s.repo.CreateStaffProfile(ctx, profile); err != nil {
		return fmt.Errorf("failed to create staff profile: %w", err)
	}
	if err := s.repo.UseInvite(ctx, invite.ID); err != nil {
		return fmt.Errorf("failed to mark invite as used: %w", err)
	}
	return nil
}

func (s *Service) CreateDefaultLocation(
	ctx context.Context,
	businessID uuid.UUID,
) (uuid.UUID, error) {
	if businessID == uuid.Nil {
		return uuid.Nil, fmt.Errorf("business ID cannot be nil")
	}

	business, err := s.repo.GetBusinessByID(ctx, businessID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to verify business: %w", err)
	}
	if business == nil {
		return uuid.Nil, fmt.Errorf("business not found")
	}

	location := &Location{
		ID:         uuid.New(),
		BusinessID: businessID,
		Name:       "Default Location",
		IsActive:   true,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	if err := s.ValidateLocation(location); err != nil {
		return uuid.Nil, fmt.Errorf("location validation failed: %w", err)
	}
	if err := s.repo.CreateLocation(ctx, location); err != nil {
		return uuid.Nil, fmt.Errorf("failed to create default location: %w", err)
	}

	return location.ID, nil
}
