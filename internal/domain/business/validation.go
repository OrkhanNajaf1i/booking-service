package business

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

func (s *Service) ValidateBusiness(b *Business) error {
	if b == nil {
		return fmt.Errorf("business data cannot be nil")
	}

	if err := s.validateBusinessName(b.Name); err != nil {
		return err
	}
	if err := s.validatePhone(b.Phone); err != nil {
		return err
	}
	if err := s.validateBusinessTypeRules(b); err != nil {
		return err
	}

	return nil
}

func (s *Service) ValidateStaffProfile(sp *StaffProfile) error {
	if sp == nil {
		return fmt.Errorf("staff profile data cannot be nil")
	}

	if err := s.validateUUID(sp.UserID, "User ID"); err != nil {
		return err
	}

	if err := s.validateUUID(sp.BusinessID, "Business ID"); err != nil {
		return err
	}

	if err := s.validateStaffRole(sp.Role); err != nil {
		return err
	}

	if err := s.validateTitle(sp.Title); err != nil {
		return err
	}

	if err := s.validateStatus(sp.Status); err != nil {
		return err
	}

	return nil
}

func (s *Service) ValidateBusinessInvite(bi *BusinessInvite) error {
	if bi == nil {
		return fmt.Errorf("business invite data cannot be nil")
	}
	if err := s.validateUUID(bi.BusinessID, "Business ID"); err != nil {
		return err
	}
	if strings.TrimSpace(bi.InvitedEmail) == "" && strings.TrimSpace(bi.InvitedPhone) == "" {
		return fmt.Errorf("either email or phone is required")
	}
	if strings.TrimSpace(bi.InvitedEmail) != "" {
		if err := s.validateEmail(bi.InvitedEmail); err != nil {
			return err
		}
	}
	if strings.TrimSpace(bi.InvitedPhone) != "" {
		if err := s.validatePhone(bi.InvitedPhone); err != nil {
			return err
		}
	}
	if err := s.validateStaffRole(bi.Role); err != nil {
		return err
	}
	if strings.TrimSpace(bi.Token) == "" {
		return fmt.Errorf("invite token is required")
	}
	return nil
}

func (s *Service) ValidateLocation(loc *Location) error {
	if loc == nil {
		return fmt.Errorf("location data cannot be nil")
	}
	if err := s.validateUUID(loc.BusinessID, "Business ID"); err != nil {
		return err
	}
	if err := s.validateLocationName(loc.Name); err != nil {
		return err
	}
	return nil
}

func (s *Service) validateBusinessName(name string) error {
	cleanName := strings.TrimSpace(name)
	if cleanName == "" {
		return fmt.Errorf("business name is required")
	}
	if len(cleanName) < 2 {
		return fmt.Errorf("business name must be at least 2 characters")
	}
	if len(cleanName) > 100 {
		return fmt.Errorf("business name cannot exceed 100 characters")
	}

	return nil
}

func (s *Service) validateLocationName(name string) error {
	cleanName := strings.TrimSpace(name)
	if cleanName == "" {
		return fmt.Errorf("location name is required")
	}
	if len(cleanName) < 2 {
		return fmt.Errorf("location name must be at least 2 characters")
	}
	if len(cleanName) > 100 {
		return fmt.Errorf("location name cannot exceed 100 characters")
	}

	return nil
}

func (s *Service) validateTitle(title string) error {
	cleanTitle := strings.TrimSpace(title)
	if cleanTitle == "" {
		return fmt.Errorf("staff title is required")
	}
	if len(cleanTitle) < 2 {
		return fmt.Errorf("staff title must be at least 2 characters")
	}
	if len(cleanTitle) > 50 {
		return fmt.Errorf("staff title cannot exceed 50 characters")
	}

	return nil
}

func (s *Service) validatePhone(phone string) error {
	cleanPhone := strings.TrimSpace(phone)
	if cleanPhone == "" {
		return fmt.Errorf("phone number is required")
	}
	phoneRegex := regexp.MustCompile(`^\+?[0-9]{7,15}$`)
	if !phoneRegex.MatchString(cleanPhone) {
		return fmt.Errorf("invalid phone format (e.g., +994501234567)")
	}

	return nil
}

func (s *Service) validateEmail(email string) error {
	cleanEmail := strings.TrimSpace(email)
	if cleanEmail == "" {
		return fmt.Errorf("email is required")
	}
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(cleanEmail) {
		return fmt.Errorf("invalid email format")
	}
	return nil
}

func (s *Service) validateUUID(id uuid.UUID, fieldName string) error {
	if id == uuid.Nil {
		return fmt.Errorf("%s is required", fieldName)
	}
	return nil
}

func (s *Service) validateStatus(status string) error {
	cleanStatus := strings.ToLower(strings.TrimSpace(status))
	validStatuses := map[string]bool{
		"active":   true,
		"inactive": true,
		"pending":  true,
	}

	if !validStatuses[cleanStatus] {
		return fmt.Errorf("invalid status: %s (valid: active, inactive, pending)", status)
	}

	return nil
}

func (s *Service) validateStaffRole(role StaffRole) error {
	switch role {
	case StaffRoleAdmin, StaffRoleManager, StaffRoleStaff:
		return nil
	default:
		return fmt.Errorf("invalid staff role: %s (valid: admin, manager, staff)", role)
	}
}

func (s *Service) validateBusinessTypeRules(b *Business) error {
	if !b.BusinessType.IsValid() {
		return fmt.Errorf("invalid business type: %s", b.BusinessType)
	}

	switch b.BusinessType {
	case BusinessTypeSolo:
		if err := s.validateServiceCategory(b.ServiceCategory); err != nil {
			return err
		}

	case BusinessTypeMulti:
		if err := s.validateIndustry(b.Industry); err != nil {
			return err
		}

	default:
		return fmt.Errorf("unknown business type: %s", b.BusinessType)
	}

	return nil
}

func (s *Service) validateServiceCategory(category string) error {
	cleanCat := strings.TrimSpace(category)

	if cleanCat == "" {
		return fmt.Errorf("service_category is required for solo business")
	}
	if len(cleanCat) < 3 {
		return fmt.Errorf("service_category must be at least 3 characters")
	}
	if len(cleanCat) > 50 {
		return fmt.Errorf("service_category cannot exceed 50 characters")
	}

	return nil
}

func (s *Service) validateIndustry(industry string) error {
	cleanInd := strings.TrimSpace(industry)

	if cleanInd == "" {
		return fmt.Errorf("industry is required for multi-staff business")
	}
	if len(cleanInd) < 3 {
		return fmt.Errorf("industry must be at least 3 characters")
	}
	if len(cleanInd) > 50 {
		return fmt.Errorf("industry cannot exceed 50 characters")
	}

	return nil
}

func (s *Service) ValidateCreateBusinessRequest(req *CreateBusinessRequest, businessType BusinessType) error {
	if req == nil {
		return fmt.Errorf("request cannot be nil")
	}

	if err := s.validateBusinessName(req.Name); err != nil {
		return err
	}

	if err := s.validatePhone(req.Phone); err != nil {
		return err
	}

	switch businessType {
	case BusinessTypeSolo:
		if err := s.validateServiceCategory(req.ServiceCategory); err != nil {
			return err
		}

	case BusinessTypeMulti:
		if err := s.validateIndustry(req.Industry); err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) ValidateInviteStaffRequest(req *InviteStaffRequest) error {
	if req == nil {
		return fmt.Errorf("request cannot be nil")
	}

	if strings.TrimSpace(req.Email) == "" && strings.TrimSpace(req.Phone) == "" {
		return fmt.Errorf("either email or phone is required")
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

	if err := s.validateStaffRole(req.Role); err != nil {
		return err
	}

	return nil
}
