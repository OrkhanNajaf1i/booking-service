// File: internal/domain/staff/entity.go
package staff

import (
	"time"

	"github.com/google/uuid"
)

type StaffRole string

const (
	StaffRoleAdmin   StaffRole = "admin"
	StaffRoleManager StaffRole = "manager"
	StaffRoleStaff   StaffRole = "staff"
)

func (sr StaffRole) IsValid() bool {
	return sr == StaffRoleAdmin || sr == StaffRoleManager || sr == StaffRoleStaff
}

type StaffStatus string

const (
	StaffStatusActive   StaffStatus = "active"
	StaffStatusInactive StaffStatus = "inactive"
	StaffStatusPending  StaffStatus = "pending"
)

type StaffProfile struct {
	ID         uuid.UUID   `db:"id" json:"id"`
	UserID     uuid.UUID   `db:"user_id" json:"user_id"`
	BusinessID uuid.UUID   `db:"business_id" json:"business_id"`
	LocationID *uuid.UUID  `db:"location_id" json:"location_id,omitempty"`
	Role       StaffRole   `db:"role" json:"role"`
	Title      string      `db:"title" json:"title"`
	Department string      `db:"department" json:"department"`
	Bio        string      `db:"bio" json:"bio"`
	HourlyRate float64     `db:"hourly_rate" json:"hourly_rate"`
	Status     StaffStatus `db:"status" json:"status"`
	JoinedAt   time.Time   `db:"joined_at" json:"joined_at"`
	CreatedAt  time.Time   `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time   `db:"updated_at" json:"updated_at"`
}

func NewStaffProfile(userID, businessID uuid.UUID, role StaffRole, title string) *StaffProfile {
	now := time.Now()
	return &StaffProfile{
		ID:         uuid.New(),
		UserID:     userID,
		BusinessID: businessID,
		Role:       role,
		Title:      title,
		Status:     StaffStatusActive,
		JoinedAt:   now,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

type StaffWithUser struct {
	ID         uuid.UUID   `db:"id" json:"id"`
	UserID     uuid.UUID   `db:"user_id" json:"user_id"`
	FullName   string      `db:"full_name" json:"full_name"`
	Email      string      `db:"email" json:"email"`
	Phone      string      `db:"phone" json:"phone"`
	Avatar     *string     `db:"avatar" json:"avatar,omitempty"`
	Role       StaffRole   `db:"role" json:"role"`
	Title      string      `db:"title" json:"title"`
	Department string      `db:"department" json:"department"`
	LocationID *uuid.UUID  `db:"location_id" json:"location_id,omitempty"`
	Status     StaffStatus `db:"status" json:"status"`
	JoinedAt   time.Time   `db:"joined_at" json:"joined_at"`
}

type BusinessInvite struct {
	ID           uuid.UUID  `db:"id" json:"id"`
	BusinessID   uuid.UUID  `db:"business_id" json:"business_id"`
	InvitedEmail string     `db:"invited_email" json:"invited_email"`
	InvitedPhone string     `db:"invited_phone" json:"invited_phone"`
	Role         StaffRole  `db:"role" json:"role"`
	LocationID   *uuid.UUID `db:"location_id" json:"location_id,omitempty"`
	Token        string     `db:"token" json:"token"`
	ExpiresAt    time.Time  `db:"expires_at" json:"expires_at"`
	Used         bool       `db:"used" json:"used"`
	CreatedAt    time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at" json:"updated_at"`
}

type CreateStaffRequest struct {
	UserID     uuid.UUID  `json:"user_id"`
	Role       StaffRole  `json:"role"`
	Title      string     `json:"title"`
	LocationID *uuid.UUID `json:"location_id,omitempty"`
}

type UpdateStaffRequest struct {
	Role       StaffRole  `json:"role"`
	Title      string     `json:"title"`
	Department string     `json:"department"`
	Bio        string     `json:"bio"`
	HourlyRate float64    `json:"hourly_rate"`
	LocationID *uuid.UUID `json:"location_id,omitempty"`
}

type InviteStaffRequest struct {
	Email      string     `json:"email"`
	Phone      string     `json:"phone"`
	Role       StaffRole  `json:"role"`
	LocationID *uuid.UUID `json:"location_id,omitempty"`
}

type AcceptInviteRequest struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}

type StaffError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *StaffError) Error() string {
	return e.Message
}
