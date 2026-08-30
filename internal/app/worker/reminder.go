// File: internal/app/worker/reminder.go
package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/OrkhanNajaf1i/booking-service/internal/domain/notification"
	"github.com/OrkhanNajaf1i/booking-service/internal/logger"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// reminderLeadTime – randevudan ne qeder evvel xatirladilsin.
const reminderLeadTime = 60 * time.Minute

// ReminderJob – yaxinlasan randevular ucun xatirlatma yaradir.
type ReminderJob struct {
	db            *sqlx.DB
	notifications notification.Service
	log           logger.Logger
	displayTZ     *time.Location
}

func NewReminderJob(
	db *sqlx.DB,
	notifications notification.Service,
	log logger.Logger,
	timezone string,
) *ReminderJob {
	location, err := time.LoadLocation(timezone)
	if err != nil || location == nil {
		location = time.UTC
	}

	return &ReminderJob{
		db:            db,
		notifications: notifications,
		log:           log,
		displayTZ:     location,
	}
}

// reminderRow – xatirlatma ucun lazim olan minimal data.
type reminderRow struct {
	BookingID      uuid.UUID  `db:"booking_id"`
	BusinessID     uuid.UUID  `db:"business_id"`
	BusinessName   string     `db:"business_name"`
	StartTime      time.Time  `db:"start_time"`
	StaffUserID    uuid.UUID  `db:"staff_user_id"`
	StaffName      string     `db:"staff_name"`
	CustomerUserID *uuid.UUID `db:"customer_user_id"`
	CustomerName   string     `db:"customer_name"`
}

// Run – novbeti bir saat erzinde baslayacaq tesdiqlenmis randevular ucun
// hele xatirlatma gonderilmeyenlerini tapib bildiris yaradir.
//
// Tekrarin qarsisi notifications cedveline baxmaqla alinir: eyni booking
// ucun "booking.reminder" tipli setir varsa kecilir. Ayrica bayraq
// sutunu saxlamaga ehtiyac qalmır.
func (j *ReminderJob) Run(ctx context.Context) (int, error) {
	query := `
		SELECT
			b.id           AS booking_id,
			b.business_id  AS business_id,
			bus.name       AS business_name,
			b.start_time   AS start_time,
			su.id          AS staff_user_id,
			su.full_name   AS staff_name,
			c.user_id      AS customer_user_id,
			c.full_name    AS customer_name
		FROM bookings b
		JOIN businesses     bus ON bus.id = b.business_id
		JOIN staff_profiles sp  ON sp.id  = b.staff_id
		JOIN users          su  ON su.id  = sp.user_id
		JOIN customers      c   ON c.id   = b.customer_id
		WHERE b.status = 'confirmed'
		  AND b.start_time > NOW()
		  AND b.start_time <= NOW() + $1::interval
		  AND NOT EXISTS (
			  SELECT 1 FROM notifications n
			  WHERE n.booking_id = b.id AND n.type = 'booking.reminder'
		  )
		ORDER BY b.start_time
		LIMIT 100`

	interval := fmt.Sprintf("%d minutes", int(reminderLeadTime.Minutes()))

	rows := make([]reminderRow, 0, 16)
	if err := j.db.SelectContext(ctx, &rows, query, interval); err != nil {
		return 0, fmt.Errorf("xatirlatma sorgusu ugursuz: %w", err)
	}
	if len(rows) == 0 {
		return 0, nil
	}

	envelopes := make([]*notification.Envelope, 0, len(rows)*2)
	for _, row := range rows {
		when := row.StartTime.In(j.displayTZ).Format("15:04")
		businessID := row.BusinessID
		bookingID := row.BookingID

		payload := notification.JSONMap{
			"booking_id": bookingID.String(),
			"start_time": row.StartTime.Format(time.RFC3339),
			"staff_name": row.StaffName,
		}

		if row.CustomerUserID != nil && *row.CustomerUserID != uuid.Nil {
			envelopes = append(envelopes, &notification.Envelope{
				Type:       notification.TypeBookingReminder,
				UserID:     *row.CustomerUserID,
				BusinessID: &businessID,
				BookingID:  &bookingID,
				Title:      "Randevu xatirlatmasi",
				Body: fmt.Sprintf("%s saat %s-de %s ile randevunuz var.",
					row.BusinessName, when, row.StaffName),
				Payload:   payload,
				CreatedAt: time.Now(),
			})
		}

		envelopes = append(envelopes, &notification.Envelope{
			Type:       notification.TypeBookingReminder,
			UserID:     row.StaffUserID,
			BusinessID: &businessID,
			BookingID:  &bookingID,
			Title:      "Yaxinlasan randevu",
			Body:       fmt.Sprintf("Saat %s – %s.", when, row.CustomerName),
			Payload:    payload,
			CreatedAt:  time.Now(),
		})
	}

	if err := j.notifications.DispatchMany(ctx, envelopes); err != nil {
		// Bir hisse gonderilmis ola biler; sayi yene de qaytaririq.
		j.log.Warn("Bezi xatirlatmalar gonderilmedi",
			logger.Field{Key: "error", Value: err.Error()},
		)
	}

	return len(envelopes), nil
}
