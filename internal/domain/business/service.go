// File: internal/domain/business/service.go
package business

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type BusinessService struct {
	repository Repository
}

func NewService(repository Repository) *BusinessService {
	return &BusinessService{
		repository: repository,
	}
}

func (service *BusinessService) CreateBusiness(
	ctx context.Context,
	ownerID uuid.UUID,
	request *CreateBusinessRequest,
) (*Business, error) {
	if ownerID == uuid.Nil {
		return nil, NewBusinessError("INVALID_OWNER_ID", "Owner ID cannot be empty")
	}

	if request == nil {
		return nil, NewBusinessError("INVALID_REQUEST", "Request cannot be nil")
	}

	if err := service.validateCreateRequest(request); err != nil {
		return nil, err
	}
	business := NewBusiness(
		request.Name,
		request.Industry,
		request.ServiceCategory,
		request.Phone,
		request.BusinessType,
	)
	business.OwnerID = ownerID

	if err := service.validateBusiness(business); err != nil {
		return nil, err
	}

	if err := service.repository.Create(ctx, business); err != nil {
		return nil, fmt.Errorf("failed to create business: %w", err)
	}

	return business, nil
}

func (service *BusinessService) GetBusinessByID(ctx context.Context, id uuid.UUID) (*Business, error) {
	if id == uuid.Nil {
		return nil, NewBusinessError("INVALID_BUSINESS_ID", "Business ID cannot be empty")
	}

	business, err := service.repository.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get business: %w", err)
	}

	if business == nil {
		return nil, NewBusinessError("BUSINESS_NOT_FOUND", "Business not found")
	}

	return business, nil
}

func (service *BusinessService) GetBusinessByOwner(ctx context.Context, ownerID uuid.UUID) (*Business, error) {
	if ownerID == uuid.Nil {
		return nil, NewBusinessError("INVALID_OWNER_ID", "Owner ID cannot be empty")
	}

	business, err := service.repository.GetByOwnerID(ctx, ownerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get business by owner: %w", err)
	}

	if business == nil {
		return nil, NewBusinessError("BUSINESS_NOT_FOUND", "No business found for this owner")
	}

	return business, nil
}

func (service *BusinessService) UpdateBusiness(
	ctx context.Context,
	businessID uuid.UUID,
	request *UpdateBusinessRequest,
) error {
	if businessID == uuid.Nil {
		return NewBusinessError("INVALID_BUSINESS_ID", "Business ID cannot be empty")
	}

	if request == nil {
		return NewBusinessError("INVALID_REQUEST", "Request cannot be nil")
	}

	business, err := service.repository.GetByID(ctx, businessID)
	if err != nil {
		return fmt.Errorf("failed to get business: %w", err)
	}

	if business == nil {
		return NewBusinessError("BUSINESS_NOT_FOUND", "Business not found")
	}

	business.Name = request.Name
	business.Industry = request.Industry
	business.Phone = request.Phone
	business.UpdatedAt = time.Now()

	if err := service.validateBusiness(business); err != nil {
		return err
	}

	if err := service.repository.Update(ctx, business); err != nil {
		return fmt.Errorf("failed to update business: %w", err)
	}

	return nil
}
