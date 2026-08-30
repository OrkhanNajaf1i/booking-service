// File: internal/infrastructure/adapters/service_duration.go
//
// Kicik korpu adapterler: bir domenin porta ehtiyacini basqa domenin
// mövcud repository-si ile odeyir. Bele olanda availability paketi
// xidmet kataloquna birbasa asili olmur.
package adapters

import (
	"context"

	"github.com/OrkhanNajaf1i/booking-service/internal/domain/service"
	"github.com/google/uuid"
)

// ServiceDuration – availability.ServiceDurationProvider realizasiyasi.
type ServiceDuration struct {
	repo service.Repository
}

func NewServiceDuration(repo service.Repository) *ServiceDuration {
	return &ServiceDuration{repo: repo}
}

// GetDurationMins – xidmetin dakika ile mueddeti.
func (a *ServiceDuration) GetDurationMins(
	ctx context.Context,
	businessID, serviceID uuid.UUID,
) (int, error) {
	found, err := a.repo.GetByID(ctx, serviceID, businessID)
	if err != nil {
		return 0, err
	}
	if found == nil {
		return 0, &service.ServiceError{Code: "NOT_FOUND", Message: "Service not found"}
	}
	return found.DurationMinutes, nil
}
