// File: internal/domain/location/validation.go
package location

import (
	"regexp"
	"strings"
)

func (s *LocationService) validateLocation(loc *Location) error {
	if loc == nil {
		return &LocationError{Code: "INVALID_DATA", Message: "Location data cannot be nil"}
	}

	if err := s.validateLocationName(loc.Name); err != nil {
		return err
	}

	if loc.Phone != nil {
		if err := s.validatePhone(*loc.Phone); err != nil {
			return err
		}
	}

	if err := s.validateCoordinates(loc.Latitude, loc.Longitude); err != nil {
		return err
	}

	return nil
}

func (s *LocationService) validateCreateRequest(req *CreateLocationRequest) error {
	if req == nil {
		return &LocationError{Code: "INVALID_REQUEST", Message: "Request cannot be nil"}
	}

	if err := s.validateLocationName(req.Name); err != nil {
		return err
	}

	if req.Phone != nil {
		if err := s.validatePhone(*req.Phone); err != nil {
			return err
		}
	}

	return nil
}

func (s *LocationService) validateLocationName(name string) error {
	clean := strings.TrimSpace(name)
	if clean == "" {
		return &LocationError{Code: "NAME_REQUIRED", Message: "Location name is required"}
	}
	if len(clean) < 2 {
		return &LocationError{Code: "NAME_TOO_SHORT", Message: "Location name must be at least 2 characters"}
	}
	if len(clean) > 100 {
		return &LocationError{Code: "NAME_TOO_LONG", Message: "Location name cannot exceed 100 characters"}
	}
	return nil
}

func (s *LocationService) validatePhone(phone string) error {
	clean := strings.TrimSpace(phone)
	if clean == "" {
		return nil // Optional field
	}
	phoneRegex := regexp.MustCompile(`^\+?[0-9]{7,15}$`)
	if !phoneRegex.MatchString(clean) {
		return &LocationError{Code: "PHONE_INVALID", Message: "Invalid phone format"}
	}
	return nil
}

// Koordinat ya tam verilir, ya da hec verilmir — yarimciq deyer
// xeritede yanlis noqte gosterer. Bazada da eyni CHECK var (009),
// amma orada pozulma 500 kimi qayidir; istifadeci 400 gormelidir.
func (s *LocationService) validateCoordinates(lat, lng *float64) error {
	if (lat == nil) != (lng == nil) {
		return &LocationError{
			Code:    "COORDINATES_INCOMPLETE",
			Message: "Latitude and longitude must be provided together",
		}
	}
	if lat == nil {
		return nil // koordinat verilmeyib — elle yazma yolu
	}
	if *lat < -90 || *lat > 90 {
		return &LocationError{
			Code:    "LATITUDE_OUT_OF_RANGE",
			Message: "Latitude must be between -90 and 90",
		}
	}
	if *lng < -180 || *lng > 180 {
		return &LocationError{
			Code:    "LONGITUDE_OUT_OF_RANGE",
			Message: "Longitude must be between -180 and 180",
		}
	}
	return nil
}
