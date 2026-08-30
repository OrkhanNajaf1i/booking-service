// File: internal/app/worker/expiry.go
package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/OrkhanNajaf1i/booking-service/internal/domain/booking"
	"github.com/OrkhanNajaf1i/booking-service/internal/domain/notification"
	"github.com/OrkhanNajaf1i/booking-service/internal/logger"
	"github.com/google/uuid"
)

// PendingExpiryJob – cavabsiz qalmis bronlari legv edir.
//
// Provider sorguya baxmasa, bron "pending" qalir ve slotu bloklayir:
// hemin vaxti basqa musteriye vermek mumkun olmur. Muddet her biznesin
// schedule_settings-inde teyin olunur (default 24 saat).
//
// Legv edilende her iki teref xeberdar edilir — musteri niye cavab
// almadigini bilmelidir.
type PendingExpiryJob struct {
	bookings      booking.Repository
	participants  booking.ParticipantResolver
	notifications notification.Service
	log           logger.Logger
	displayTZ     *time.Location
}

func NewPendingExpiryJob(
	bookings booking.Repository,
	participants booking.ParticipantResolver,
	notifications notification.Service,
	log logger.Logger,
	timezone string,
) *PendingExpiryJob {
	location, err := time.LoadLocation(timezone)
	if err != nil || location == nil {
		location = time.UTC
	}

	return &PendingExpiryJob{
		bookings:      bookings,
		participants:  participants,
		notifications: notifications,
		log:           log,
		displayTZ:     location,
	}
}

// Run – bir dovre; legv edilen bron sayini qaytarir.
func (j *PendingExpiryJob) Run(ctx context.Context) (int, error) {
	expired, err := j.bookings.ExpireStalePending(ctx, 50)
	if err != nil {
		return 0, fmt.Errorf("cavabsiz bronlar legv edilmedi: %w", err)
	}
	if len(expired) == 0 {
		return 0, nil
	}

	envelopes := make([]*notification.Envelope, 0, len(expired)*2)

	for _, item := range expired {
		j.log.Info("Cavabsiz bron legv edildi",
			logger.Field{Key: "booking_id", Value: item.ID.String()},
			logger.Field{Key: "start_time", Value: item.StartTime.Format(time.RFC3339)},
		)

		parties, resolveErr := j.participants.Resolve(
			ctx, item.BusinessID, item.StaffID, item.CustomerID, item.ServiceID,
		)
		if resolveErr != nil || parties == nil {
			continue
		}

		when := item.StartTime.In(j.displayTZ).Format("02.01.2006 15:04")
		businessID := item.BusinessID
		bookingID := item.ID

		payload := notification.JSONMap{
			"booking_id": bookingID.String(),
			"start_time": item.StartTime.Format(time.RFC3339),
			"reason":     "expired",
		}

		add := func(recipient uuid.UUID, title, body string) {
			if recipient == uuid.Nil {
				return
			}
			envelopes = append(envelopes, &notification.Envelope{
				Type:       notification.TypeBookingCancelled,
				UserID:     recipient,
				BusinessID: &businessID,
				BookingID:  &bookingID,
				Title:      title,
				Body:       body,
				Payload:    payload,
				CreatedAt:  time.Now(),
			})
		}

		add(parties.CustomerUserID,
			"Bron legv edildi",
			fmt.Sprintf("%s tarixli sorgunuz vaxtinda cavablandirilmadi ve legv olundu. "+
				"Basqa vaxt sece bilersiniz.", when))

		add(parties.StaffUserID,
			"Cavabsiz bron legv edildi",
			fmt.Sprintf("%s – %s sorgusuna cavab verilmediyi ucun legv olundu.",
				parties.CustomerName, when))
	}

	if len(envelopes) > 0 {
		if err := j.notifications.DispatchMany(ctx, envelopes); err != nil {
			j.log.Warn("Legv bildirisleri tam gonderilmedi",
				logger.Field{Key: "error", Value: err.Error()},
			)
		}
	}

	return len(expired), nil
}
