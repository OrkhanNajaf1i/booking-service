// File: internal/domain/staff/validation.go
package staff

import (
	"regexp"
	"strings"

	"github.com/google/uuid"
)

func (s *StaffService) validateStaffProfile(profile *StaffProfile) error {
	if profile == nil {
		return &StaffError{Code: "INVALID_DATA", Message: "Staff profile data cannot be nil"}
	}

	if profile.UserID == uuid.Nil {
		return &StaffError{Code: "INVALID_USER", Message: "User ID cannot be empty"}
	}

	if profile.BusinessID == uuid.Nil {
		return &StaffError{Code: "INVALID_BUSINESS", Message: "Business ID cannot be empty"}
	}

	if !profile.Role.IsValid() {
		return &StaffError{Code: "INVALID_ROLE", Message: "Invalid staff role"}
	}

	if err := s.validateTitle(profile.Title); err != nil {
		return err
	}

	return nil
}

func (s *StaffService) validateCreateRequest(req *CreateStaffRequest) error {
	if req == nil {
		return &StaffError{Code: "INVALID_REQUEST", Message: "Request cannot be nil"}
	}

	if req.UserID == uuid.Nil {
		return &StaffError{Code: "INVALID_USER", Message: "User ID cannot be empty"}
	}

	if !req.Role.IsValid() {
		return &StaffError{Code: "INVALID_ROLE", Message: "Invalid staff role"}
	}

	if err := s.validateTitle(req.Title); err != nil {
		return err
	}

	return nil
}

func (s *StaffService) validateInviteRequest(req *InviteStaffRequest) error {
	if req == nil {
		return &StaffError{Code: "INVALID_REQUEST", Message: "Request cannot be nil"}
	}

	if strings.TrimSpace(req.Email) == "" && strings.TrimSpace(req.Phone) == "" {
		return &StaffError{Code: "CONTACT_REQUIRED", Message: "Either email or phone is required"}
	}

	if strings.TrimSpace(req.Email) != "" {
		if err := s.validateEmail(req.Email); err != nil {
			return err
		}
	}

	if strings.TrimSpace(req.Phone) != "" {
		if err := s.validatePhone(req.Phone); err != nil {
			return err
		}
	}

	if !req.Role.IsValid() {
		return &StaffError{Code: "INVALID_ROLE", Message: "Invalid staff role"}
	}

	return nil
}

func (s *StaffService) validateTitle(title string) error {
	clean := strings.TrimSpace(title)
	if clean == "" {
		return &StaffError{Code: "TITLE_REQUIRED", Message: "Staff title is required"}
	}
	if len(clean) < 2 {
		return &StaffError{Code: "TITLE_TOO_SHORT", Message: "Staff title must be at least 2 characters"}
	}
	if len(clean) > 50 {
		return &StaffError{Code: "TITLE_TOO_LONG", Message: "Staff title cannot exceed 50 characters"}
	}
	return nil
}

func (s *StaffService) validateEmail(email string) error {
	clean := strings.TrimSpace(email)
	if clean == "" {
		return nil
	}
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(clean) {
		return &StaffError{Code: "EMAIL_INVALID", Message: "Invalid email format"}
	}
	return nil
}

func (s *StaffService) validatePhone(phone string) error {
	clean := strings.TrimSpace(phone)
	if clean == "" {
		return nil
	}
	phoneRegex := regexp.MustCompile(`^\+?[0-9]{7,15}$`)
	if !phoneRegex.MatchString(clean) {
		return &StaffError{Code: "PHONE_INVALID", Message: "Invalid phone format"}
	}
	return nil
}
