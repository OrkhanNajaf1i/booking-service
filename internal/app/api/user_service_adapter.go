// File: internal/api/user_service_adapter.go
package api

import (
	"context"
	"fmt"

	"github.com/OrkhanNajaf1i/booking-service/internal/domain/auth"
	"github.com/OrkhanNajaf1i/booking-service/internal/domain/business"
	"github.com/google/uuid"
)

type UserServiceAdapter struct {
	authRepo     auth.AuthRepository
	businessRepo business.Repository
}

func NewUserServiceAdapter(
	authRepo auth.AuthRepository,
	businessRepo business.Repository,
) *UserServiceAdapter {
	return &UserServiceAdapter{
		authRepo:     authRepo,
		businessRepo: businessRepo,
	}
}

// ✅ DÜZƏLİŞ: auth.User → business.User mapping
func (a *UserServiceAdapter) GetUserByID(ctx context.Context, userID uuid.UUID) (*business.User, error) {
	authUser, err := a.authRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("adapter: failed to get user: %w", err)
	}
	if authUser == nil {
		return nil, fmt.Errorf("adapter: user not found")
	}

	// Type mapping: auth.User → business.User
	businessUser := &business.User{
		ID:         authUser.ID,
		Email:      authUser.Email,
		FullName:   authUser.FullName,
		Phone:      authUser.Phone,
		IsOwner:    authUser.IsOwner,
		BusinessID: *authUser.BusinessID,
	}

	return businessUser, nil
}

func (a *UserServiceAdapter) UpdateUserBusinessID(
	ctx context.Context,
	userID uuid.UUID,
	businessID uuid.UUID,
	isOwner bool,
) error {
	existingUser, err := a.authRepo.GetUserByID(ctx, userID)
	if err != nil || existingUser == nil {
		return fmt.Errorf("adapter: user not found: %w", err)
	}

	existingBusiness, err := a.businessRepo.GetBusinessByID(ctx, businessID)
	if err != nil || existingBusiness == nil {
		return fmt.Errorf("adapter: business not found: %w", err)
	}

	if err := a.businessRepo.UpdateUserBusinessID(ctx, userID, businessID); err != nil {
		return fmt.Errorf("adapter: failed to update user business_id: %w", err)
	}

	if isOwner {
		if err := a.businessRepo.UpdateOwner(ctx, businessID, userID); err != nil {
			return fmt.Errorf("adapter: failed to set business owner: %w", err)
		}
	}

	return nil
}
