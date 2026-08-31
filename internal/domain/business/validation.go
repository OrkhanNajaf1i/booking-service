// File: internal/domain/business/validation.go
package business

import (
	"regexp"
	"strings"
)

func (service *BusinessService) validateBusiness(business *Business) error {
	if business == nil {
		return NewBusinessError("INVALID_DATA", "Business data cannot be nil")
	}

	if err := service.validateBusinessName(business.Name); err != nil {
		return err
	}

	if err := service.validatePhone(business.Phone); err != nil {
		return err
	}

	if !business.BusinessType.IsValid() {
		return NewBusinessError("INVALID_BUSINESS_TYPE", "Invalid business type")
	}

	// Sahe metni isteye baglidir — qruplasdirmani category_slug teyin
	// edir. Verilibse uzunluguna baxilir.
	if strings.TrimSpace(business.ServiceCategory) != "" {
		if err := service.validateServiceCategory(business.ServiceCategory); err != nil {
			return err
		}
	}

	return nil
}

func (service *BusinessService) validateCreateRequest(request *CreateBusinessRequest) error {
	if request == nil {
		return NewBusinessError("INVALID_REQUEST", "Request cannot be nil")
	}

	if err := service.validateBusinessName(request.Name); err != nil {
		return err
	}

	if err := service.validatePhone(request.Phone); err != nil {
		return err
	}

	if !request.BusinessType.IsValid() {
		return NewBusinessError("INVALID_BUSINESS_TYPE", "Invalid business type")
	}

	// Sahe metni artiq isteye baglidir: qruplasdirmani sabit kateqoriya
	// secimi teyin edir (bax: CreateBusiness). Metn yalniz kartda alt
	// basliq kimi gorunur, ona gore verilibse uzunluguna baxilir,
	// verilmeyibse buraxilir.
	if strings.TrimSpace(request.ServiceCategory) != "" {
		if err := service.validateServiceCategory(request.ServiceCategory); err != nil {
			return err
		}
	}

	return nil
}

func (service *BusinessService) validateBusinessName(name string) error {
	cleanName := strings.TrimSpace(name)

	if cleanName == "" {
		return NewBusinessError("BUSINESS_NAME_REQUIRED", "Business name is required")
	}

	if len(cleanName) < 2 {
		return NewBusinessError("BUSINESS_NAME_TOO_SHORT", "Business name must be at least 2 characters")
	}

	if len(cleanName) > 100 {
		return NewBusinessError("BUSINESS_NAME_TOO_LONG", "Business name cannot exceed 100 characters")
	}

	return nil
}

func (service *BusinessService) validatePhone(phone string) error {
	cleanPhone := strings.TrimSpace(phone)

	if cleanPhone == "" {
		return NewBusinessError("PHONE_REQUIRED", "Phone number is required")
	}

	phoneRegex := regexp.MustCompile(`^\+?[0-9]{7,15}$`)
	if !phoneRegex.MatchString(cleanPhone) {
		return NewBusinessError("PHONE_INVALID", "Invalid phone format (example: +994501234567)")
	}

	return nil
}

// validateServiceCategory – yalniz metn verilibse cagirilir.
func (service *BusinessService) validateServiceCategory(category string) error {
	cleanCategory := strings.TrimSpace(category)

	if len(cleanCategory) < 3 {
		return NewBusinessError("SERVICE_CATEGORY_TOO_SHORT", "Service category must be at least 3 characters")
	}

	if len(cleanCategory) > 50 {
		return NewBusinessError("SERVICE_CATEGORY_TOO_LONG", "Service category cannot exceed 50 characters")
	}

	return nil
}

func (service *BusinessService) validateIndustry(industry string) error {
	cleanIndustry := strings.TrimSpace(industry)

	if cleanIndustry == "" {
		return NewBusinessError("INDUSTRY_REQUIRED", "Industry is required for multi-staff business")
	}

	if len(cleanIndustry) < 3 {
		return NewBusinessError("INDUSTRY_TOO_SHORT", "Industry must be at least 3 characters")
	}

	if len(cleanIndustry) > 50 {
		return NewBusinessError("INDUSTRY_TOO_LONG", "Industry cannot exceed 50 characters")
	}

	return nil
}
