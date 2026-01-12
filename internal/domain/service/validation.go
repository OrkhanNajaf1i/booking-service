// File: internal/domain/service/validation.go
package service

import (
	"strings"

	"github.com/google/uuid"
)

func (s *ServiceService) validateService(svc *Service) error {
	if svc == nil {
		return &ServiceError{Code: "INVALID_DATA", Message: "Service data cannot be nil"}
	}
	if svc.BusinessID == uuid.Nil {
		return &ServiceError{Code: "INVALID_BUSINESS", Message: "Business ID cannot be empty"}
	}
	if err := s.validateName(svc.Name); err != nil {
		return err
	}
	if err := s.validateDuration(svc.DurationMinutes); err != nil {
		return err
	}
	if err := s.validatePrice(svc.Price); err != nil {
		return err
	}
	return nil
}

func (s *ServiceService) validateCreateRequest(req *CreateServiceRequest) error {
	if req == nil {
		return &ServiceError{Code: "INVALID_REQUEST", Message: "Request cannot be nil"}
	}
	if err := s.validateName(req.Name); err != nil {
		return err
	}
	if err := s.validateDuration(req.DurationMinutes); err != nil {
		return err
	}
	if err := s.validatePrice(req.Price); err != nil {
		return err
	}
	return nil
}

func (s *ServiceService) validateName(name string) error {
	clean := strings.TrimSpace(name)
	if clean == "" {
		return &ServiceError{Code: "NAME_REQUIRED", Message: "Service name is required"}
	}
	if len(clean) < 2 {
		return &ServiceError{Code: "NAME_TOO_SHORT", Message: "Service name must be at least 2 characters"}
	}
	if len(clean) > 100 {
		return &ServiceError{Code: "NAME_TOO_LONG", Message: "Service name cannot exceed 100 characters"}
	}
	return nil
}

func (s *ServiceService) validateDuration(duration int) error {
	if duration <= 0 {
		return &ServiceError{Code: "DURATION_INVALID", Message: "Duration must be greater than zero"}
	}
	if duration > 24*60 {
		return &ServiceError{Code: "DURATION_TOO_LONG", Message: "Duration cannot exceed 1440 minutes"}
	}
	return nil
}

func (s *ServiceService) validatePrice(price float64) error {
	if price < 0 {
		return &ServiceError{Code: "PRICE_INVALID", Message: "Price cannot be negative"}
	}
	return nil
}

func (s *ServiceService) validateAssignServicesRequest(serviceIDs []uuid.UUID) error {
	if len(serviceIDs) == 0 {
		return &ServiceError{Code: "SERVICE_LIST_EMPTY", Message: "At least one service ID is required"}
	}
	if len(serviceIDs) > 100 {
		return &ServiceError{Code: "SERVICE_LIST_TOO_LONG", Message: "Too many services in a single request"}
	}
	for _, id := range serviceIDs {
		if id == uuid.Nil {
			return &ServiceError{Code: "INVALID_SERVICE", Message: "Service ID cannot be empty"}
		}
	}
	return nil
}
