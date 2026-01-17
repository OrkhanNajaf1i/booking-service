// File: internal/domain/customer/service.go
package customer

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type service struct {
	repo Repository
}

// NewService - Service instance-ı yaratır
func NewService(repo Repository) Service {
	return &service{
		repo: repo,
	}
}

// CreateCustomer - Biznes tərəfindən müştəri yaratmaq
func (s *service) CreateCustomer(ctx context.Context, businessID uuid.UUID, req *CreateCustomerRequest) (*Customer, error) {
	// 1. Request validation
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// 2. Email uniqueness check (yalnız HƏMİN business daxilində)
	existing, _ := s.repo.GetByEmail(ctx, businessID, req.Email)
	if existing != nil {
		return nil, ErrEmailAlreadyExists
	}

	// 3. Entity yaratır
	customer := NewCustomer(businessID, req.FullName, req.Email, req.Phone)
	customer.Notes = req.Notes

	// 4. Database-ə yaz
	if err := s.repo.Create(ctx, customer); err != nil {
		return nil, fmt.Errorf("failed to create customer: %w", err)
	}

	return customer, nil
}

// GetCustomer - Konkret müştərini gətirir
func (s *service) GetCustomer(ctx context.Context, businessID, id uuid.UUID) (*Customer, error) {
	customer, err := s.repo.GetByID(ctx, businessID, id)
	if err != nil {
		return nil, ErrCustomerNotFound
	}

	// businessID yoxlaması (multi-tenancy security)
	if customer.BusinessID != businessID {
		return nil, ErrAccessDenied
	}

	return customer, nil
}

// GetCustomerByUserID - User ID-dən müştəri tap
func (s *service) GetCustomerByUserID(ctx context.Context, businessID, userID uuid.UUID) (*Customer, error) {
	customer, err := s.repo.GetByUserID(ctx, businessID, userID)
	if err != nil {
		return nil, ErrCustomerNotFound
	}

	if customer.BusinessID != businessID {
		return nil, ErrAccessDenied
	}

	return customer, nil
}

// ListCustomers - Pagination-li siyahı
func (s *service) ListCustomers(ctx context.Context, businessID uuid.UUID, page, pageSize int) (*CustomersListResponse, error) {
	// Default values
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	// Məlumatları gət
	customers, err := s.repo.List(ctx, businessID, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list customers: %w", err)
	}

	// Total count
	total, err := s.repo.Count(ctx, businessID)
	if err != nil {
		return nil, fmt.Errorf("failed to count customers: %w", err)
	}

	// Response DTO-ya çevirt
	responses := make([]*CustomerResponse, len(customers))
	for i, c := range customers {
		responses[i] = c.ToResponse()
	}

	totalPages := (total + pageSize - 1) / pageSize

	return &CustomersListResponse{
		Data:       responses,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// UpdateCustomer - Müştəri məlumatlarını yeniləyir
func (s *service) UpdateCustomer(ctx context.Context, businessID, id uuid.UUID, req *UpdateCustomerRequest) (*Customer, error) {
	// 1. Validation
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// 2. Mövcud müştərini tap
	customer, err := s.repo.GetByID(ctx, businessID, id)
	if err != nil {
		return nil, ErrCustomerNotFound
	}

	// Multi-tenancy check
	if customer.BusinessID != businessID {
		return nil, ErrAccessDenied
	}

	// 3. Sahələri yenilə (partial update)
	if req.FullName != nil {
		customer.FullName = *req.FullName
	}

	if req.Phone != nil {
		customer.Phone = *req.Phone
	}

	if req.Email != nil {
		// Email dəyişirsə, yeni emailin unique olub-olmadığını yoxla
		if *req.Email != customer.Email {
			existing, _ := s.repo.GetByEmail(ctx, businessID, *req.Email)
			if existing != nil && existing.ID != customer.ID {
				return nil, ErrEmailAlreadyExists
			}
			customer.Email = *req.Email
		}
	}

	if req.Notes != nil {
		customer.Notes = *req.Notes
	}

	if req.Status != nil {
		customer.Status = *req.Status
	}

	customer.UpdatedAt = time.Now()

	// 4. Database-ə yaz
	if err := s.repo.Update(ctx, customer); err != nil {
		return nil, fmt.Errorf("failed to update customer: %w", err)
	}

	return customer, nil
}

// DeleteCustomer - Soft delete (status = inactive)
func (s *service) DeleteCustomer(ctx context.Context, businessID, id uuid.UUID) error {
	// Yoxla
	customer, err := s.repo.GetByID(ctx, businessID, id)
	if err != nil {
		return ErrCustomerNotFound
	}

	if customer.BusinessID != businessID {
		return ErrAccessDenied
	}

	// Soft delete
	return s.repo.Delete(ctx, businessID, id)
}
