package auth

import (
	"regexp"
)

var (
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
)

func (s *Service) validateRegisterRequest(req *RegisterRequest) error {
	if err := s.validateEmail(req.Email); err != nil {
		return err
	}
	if err := s.validatePassword(req.Password); err != nil {
		return err
	}
	if err := s.validateFullName(req.FullName); err != nil {
		return err
	}
	if req.Phone == "" {
		return &RegistrationError{
			Code:    "PHONE_REQUIRED",
			Message: "phone is required",
		}
	}
	return nil
}

func (s *Service) validateEmail(email string) error {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if email == "" {
		return &RegistrationError{
			Code:    "EMAIL_REQUIRED",
			Message: "email is required",
		}
	}
	if len(email) > 255 {
		return &RegistrationError{
			Code:    "EMAIL_TOO_LONG",
			Message: "email address is too long",
		}
	}
	if !emailRegex.MatchString(email) {
		return &RegistrationError{
			Code:    "INVALID_EMAIL_FORMAT",
			Message: "please provide a valid email address",
		}
	}
	return nil
}
func (s *Service) validatePassword(password string) error {
	if password == "" {
		return &RegistrationError{
			Code:    "PASSWORD_REQUIRED",
			Message: "Password is required",
		}
	}
	if len(password) < 8 {
		return &RegistrationError{
			Code:    "PASSWORD_TOO_SHORT",
			Message: "Password must be at least 8 characters long",
		}
	}
	if len(password) > 128 {
		return &RegistrationError{
			Code:    "PASSWORD_TOO_LONG",
			Message: "Password must not exceed 128 characters",
		}
	}
	if !regexp.MustCompile(`[A-Z]`).MatchString(password) {
		return &RegistrationError{
			Code:    "PASSWORD_WEAK",
			Message: "password must contain at least one uppercase letter",
		}
	}
	if !regexp.MustCompile(`[a-z]`).MatchString(password) {
		return &RegistrationError{
			Code:    "PASSWORD_WEAK",
			Message: "Password contain at least one lowercase letter",
		}
	}
	if !regexp.MustCompile(`[0-9]`).MatchString(password) {
		return &RegistrationError{
			Code:    "PASSWORD_WEAK",
			Message: "Password contain at least one number",
		}
	}

	if !regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};:'",.<>?/\\|` + "`" + `]`).MatchString(password) {
		return &RegistrationError{
			Code:    "PASSWORD_WEAK",
			Message: "password must contain at least one special character",
		}
	}
	return nil
}

func (s *Service) validateBusinessName(businessName string) error {
	if businessName == "" {
		return &RegistrationError{
			Code:    "BUSINESS_NAME_REQUIRED",
			Message: "business name is required",
		}
	}
	if len(businessName) < 2 {
		return &RegistrationError{
			Code:    "BUSINESS_NAME_TOO_SHORT",
			Message: "business name must be at least 2 characters long",
		}
	}
	if len(businessName) > 255 {
		return &RegistrationError{
			Code:    "BUSINESS_NAME_TOO_LONG",
			Message: "business name must not exceed 255 characters",
		}
	}
	return nil
}

func (s *Service) validateFullName(fullName string) error {
	if fullName == "" {
		return &RegistrationError{
			Code:    "FULLNAME_REQUIRED",
			Message: "fullname is required",
		}
	}
	if len(fullName) < 2 {
		return &RegistrationError{
			Code:    "FULLNAME_TOO_SHORT",
			Message: "fullname must be at least 2 characters long",
		}
	}
	if len(fullName) > 255 {
		return &RegistrationError{
			Code:    "FULLNAME_TOO_LONG",
			Message: "fullname must not exceed 255 characters",
		}
	}
	return nil
}

func (s *Service) validateLocationName(name string) error {
	if name == "" {
		return &RegistrationError{
			Code:    "LOCATION_NAME_REQUIRED",
			Message: "Location name is required",
		}
	}
	if len(name) < 2 {
		return &RegistrationError{
			Code:    "LOCATION_NAME_TOO_SHORT",
			Message: "Location name must be at least 2 characters",
		}
	}
	if len(name) > 150 {
		return &RegistrationError{
			Code:    "LOCATION_NAME_TOO_LONG",
			Message: "Location name must not exceed 100 characters",
		}
	}
	return nil
}
