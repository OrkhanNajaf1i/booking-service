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
	owners     OwnerLinker
	staff      StaffProvisioner
}

func NewService(
	repository Repository,
	owners OwnerLinker,
	staff StaffProvisioner,
) *BusinessService {
	return &BusinessService{
		repository: repository,
		owners:     owners,
		staff:      staff,
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

	// Istifadecini biznese baglayiriq. Bu olmasa JWT-ye business_id
	// dusmur ve butun biznes-kontekstli endpoint-ler 400 qaytarir.
	if service.owners != nil {
		if err := service.owners.UpdateUserBusinessID(ctx, ownerID, business.ID, true); err != nil {
			return nil, fmt.Errorf("failed to link owner to business: %w", err)
		}
	}

	// Sahib ozu de xidmet gosteren sexsdir: randevu staff_id-ye baglanir,
	// ona gore profili derhal yaradilir. Eks halda "evvelce isci elave
	// edin" ekranindan kenara cixmaq mumkun olmur.
	if service.staff != nil {
		title := request.ServiceCategory
		if title == "" {
			title = request.Industry
		}
		if err := service.staff.EnsureOwnerProfile(ctx, business.ID, ownerID, title); err != nil {
			return nil, fmt.Errorf("failed to create owner staff profile: %w", err)
		}
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
func (service *BusinessService) ListBusinesses(ctx context.Context) ([]*Business, error) {
	businesses, err := service.repository.ListBusinesses(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get business list: %w", err)
	}

	if businesses == nil {
		return nil, NewBusinessError("BUSINESS_NOT_FOUND", "No businesses found")
	}

	return businesses, nil
}
func (service *BusinessService) UpdateBusiness(
	ctx context.Context,
	businessID uuid.UUID,
	ownerID uuid.UUID,
	request *UpdateBusinessRequest,
) error {
	if businessID == uuid.Nil {
		return NewBusinessError("INVALID_BUSINESS_ID", "Business ID cannot be empty")
	}

	if request == nil {
		return NewBusinessError("INVALID_REQUEST", "Request cannot be nil")
	}

	existing, err := service.repository.GetByOwnerID(ctx, ownerID)
	if err != nil {
		return fmt.Errorf("failed to get business: %w", err)
	}
	if existing == nil {
		return NewBusinessError("BUSINESS_NOT_FOUND", "Business not found")
	}

	// Frontdan gələn ID uyğundurmu?
	if existing.ID != businessID {
		return NewBusinessError("UNAUTHORIZED", "Business ID does not match")
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
	// business.OwnerID = request.OwnerID
	business.UpdatedAt = time.Now()

	if err := service.validateBusiness(business); err != nil {
		return err
	}

	if err := service.repository.Update(ctx, business); err != nil {
		return fmt.Errorf("failed to update business: %w", err)
	}

	return nil
}
