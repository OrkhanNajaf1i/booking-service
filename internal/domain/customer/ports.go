// File: internal/domain/customer/ports.go
package customer

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Repository interface {
	// CRUD Operations
	Create(ctx context.Context, customer *Customer) error
	GetByID(ctx context.Context, businessID, id uuid.UUID) (*Customer, error)
	GetByEmail(ctx context.Context, businessID uuid.UUID, email string) (*Customer, error)
	GetByUserID(ctx context.Context, businessID, userID uuid.UUID) (*Customer, error)
	List(ctx context.Context, businessID uuid.UUID, limit, offset int) ([]*Customer, error)
	Count(ctx context.Context, businessID uuid.UUID) (int, error)
	Update(ctx context.Context, customer *Customer) error
	Delete(ctx context.Context, businessID, id uuid.UUID) error // Soft delete

	// Statistics (Booking modulundan tərəfindən çağrılır)
	IncrementBookingCount(ctx context.Context, customerID uuid.UUID) error
	UpdateLastBookingTime(ctx context.Context, customerID uuid.UUID, timestamp time.Time) error
}

// UserProfile - ResolveSelf ucun lazim olan istifadeci melumati.
type UserProfile struct {
	ID       uuid.UUID
	FullName string
	Email    string
	Phone    string
}

// UserProfileProvider - istifadeci profilini oxuyan port.
// Auth repository bu imzani odeyir.
type UserProfileProvider interface {
	GetUserProfile(ctx context.Context, userID uuid.UUID) (*UserProfile, error)
}

// Service - Business Logic (Domain Service)
type Service interface {
	CreateCustomer(ctx context.Context, businessID uuid.UUID, req *CreateCustomerRequest) (*Customer, error)
	GetCustomer(ctx context.Context, businessID, id uuid.UUID) (*Customer, error)
	GetCustomerByUserID(ctx context.Context, businessID, userID uuid.UUID) (*Customer, error)

	// ResolveSelf - musteri tetbiqi ucun: login olmus istifadecinin hemin
	// biznesdeki musteri kartini tapir, yoxdursa yaradir.
	// Bron yaratmaq ucun customer_id teleb olundugundan bu addim vacibdir.
	ResolveSelf(ctx context.Context, businessID, userID uuid.UUID) (*Customer, error)
	ListCustomers(ctx context.Context, businessID uuid.UUID, page, pageSize int) (*CustomersListResponse, error)
	UpdateCustomer(ctx context.Context, businessID, id uuid.UUID, req *UpdateCustomerRequest) (*Customer, error)
	DeleteCustomer(ctx context.Context, businessID, id uuid.UUID) error
}
