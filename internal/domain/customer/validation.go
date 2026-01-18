// File: internal/domain/customer/validation.go
package customer

import (
	"regexp"
	"strings"
)

var (
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	phoneRegex = regexp.MustCompile(`^\+?[0-9]{10,15}$`)
)

// Validate - CreateCustomerRequest-i validate et
func (r *CreateCustomerRequest) Validate() error {
	r.FullName = strings.TrimSpace(r.FullName)
	r.Email = strings.TrimSpace(strings.ToLower(r.Email))
	r.Phone = strings.TrimSpace(r.Phone)

	if len(r.FullName) < 2 || len(r.FullName) > 100 {
		return ErrInvalidCustomerData
	}

	if !emailRegex.MatchString(r.Email) {
		return ErrInvalidCustomerData
	}

	if !phoneRegex.MatchString(r.Phone) {
		return ErrInvalidCustomerData
	}

	return nil
}

// Validate - UpdateCustomerRequest-i validate et
func (r *UpdateCustomerRequest) Validate() error {
	if r.FullName != nil {
		*r.FullName = strings.TrimSpace(*r.FullName)
		if len(*r.FullName) < 2 || len(*r.FullName) > 100 {
			return ErrInvalidCustomerData
		}
	}

	if r.Email != nil {
		*r.Email = strings.TrimSpace(strings.ToLower(*r.Email))
		if !emailRegex.MatchString(*r.Email) {
			return ErrInvalidCustomerData
		}
	}

	if r.Phone != nil {
		*r.Phone = strings.TrimSpace(*r.Phone)
		if !phoneRegex.MatchString(*r.Phone) {
			return ErrInvalidCustomerData
		}
	}

	if r.Status != nil && !r.Status.IsValid() {
		return ErrInvalidStatus
	}

	if r.Notes != nil && len(*r.Notes) > 500 {
		return ErrInvalidCustomerData
	}

	return nil
}
