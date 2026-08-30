// File: internal/infrastructure/notify/booking_publisher.go
//
// Booking hadiselerini bildiris zerflerine cevirir.
//
// Qayda: hadiseni tetikleyen sexse bildiris getmir. Musteri bron
// edende yalniz provider xeber alir; provider tesdiqleyende yalniz
// musteri xeber alir.
package notify

import (
	"context"
	"fmt"
	"time"

	"github.com/OrkhanNajaf1i/booking-service/internal/domain/booking"
	"github.com/OrkhanNajaf1i/booking-service/internal/domain/notification"
	"github.com/OrkhanNajaf1i/booking-service/internal/logger"
	"github.com/google/uuid"
)

// BookingPublisher – booking.EventPublisher realizasiyasi.
type BookingPublisher struct {
	notifications notification.Service
	log           logger.Logger
	// displayTZ – bildiris metnindeki saatlarin gosterildiyi zona.
	displayTZ *time.Location
}

// NewBookingPublisher – timezone bos olarsa Asia/Baku islenir.
func NewBookingPublisher(
	notifications notification.Service,
	log logger.Logger,
	timezone string,
) *BookingPublisher {
	location, err := time.LoadLocation(timezone)
	if err != nil || location == nil {
		location, err = time.LoadLocation("Asia/Baku")
		if err != nil || location == nil {
			location = time.UTC
		}
	}

	return &BookingPublisher{
		notifications: notifications,
		log:           log,
		displayTZ:     location,
	}
}

// Publish – hadiseni uygun aliciya/alicilara catdirir.
func (p *BookingPublisher) Publish(ctx context.Context, event *booking.Event) error {
	if event == nil || event.Booking == nil {
		return nil
	}
	if event.Participants == nil {
		// Terefler tapilmayibsa kime gonderecegimizi bilmirik.
		p.log.Warn("Bildiris atlandi: terefler bos",
			logger.Field{Key: "booking_id", Value: event.Booking.ID.String()},
			logger.Field{Key: "type", Value: string(event.Type)},
		)
		return nil
	}

	envelopes := p.build(event)
	if len(envelopes) == 0 {
		return nil
	}

	return p.notifications.DispatchMany(ctx, envelopes)
}

// build – hadise novune gore zerfleri qurur.
func (p *BookingPublisher) build(event *booking.Event) []*notification.Envelope {
	current := event.Booking
	parties := event.Participants

	when := p.formatTime(current.StartTime)
	service := parties.ServiceName
	if service == "" {
		service = "randevu"
	}

	base := notification.JSONMap{
		"booking_id":    current.ID.String(),
		"status":        string(current.Status),
		"start_time":    current.StartTime.Format(time.RFC3339),
		"end_time":      current.EndTime.Format(time.RFC3339),
		"staff_id":      current.StaffID.String(),
		"customer_id":   current.CustomerID.String(),
		"customer_name": parties.CustomerName,
		"staff_name":    parties.StaffName,
		"service_name":  service,
		"business_name": parties.BusinessName,
	}
	if current.ProposedStartTime != nil {
		base["proposed_start_time"] = current.ProposedStartTime.Format(time.RFC3339)
	}
	if current.ProposedEndTime != nil {
		base["proposed_end_time"] = current.ProposedEndTime.Format(time.RFC3339)
	}

	businessID := current.BusinessID
	bookingID := current.ID

	envelopes := make([]*notification.Envelope, 0, 2)

	add := func(recipient uuid.UUID, kind notification.Type, title, body string) {
		// Hadiseni eden sexse ozune bildiris getmir.
		if recipient == uuid.Nil || recipient == event.ActorUserID {
			return
		}
		envelopes = append(envelopes, &notification.Envelope{
			Type:       kind,
			UserID:     recipient,
			BusinessID: &businessID,
			BookingID:  &bookingID,
			Title:      title,
			Body:       body,
			Payload:    cloneMap(base),
			CreatedAt:  time.Now(),
		})
	}

	switch event.Type {
	case booking.EventCreated:
		add(parties.StaffUserID, notification.TypeBookingCreated,
			"Yeni bron sorgusu",
			fmt.Sprintf("%s – %s, %s. Tesdiq gozlenilir.", parties.CustomerName, service, when))

	case booking.EventConfirmed:
		add(parties.CustomerUserID, notification.TypeBookingConfirmed,
			"Bronunuz tesdiqlendi",
			fmt.Sprintf("%s ile %s: %s tesdiqlendi.", parties.StaffName, service, when))
		// Provider ozu tesdiqleyibse ona getmir; auto_confirm halinda ise
		// aktor musteri olur ve provider-i xeberdar etmek lazimdir.
		add(parties.StaffUserID, notification.TypeBookingConfirmed,
			"Yeni tesdiqlenmis bron",
			fmt.Sprintf("%s – %s, %s.", parties.CustomerName, service, when))

	case booking.EventRescheduleProposed:
		proposedWhen := when
		if current.ProposedStartTime != nil {
			proposedWhen = p.formatTime(*current.ProposedStartTime)
		}
		body := fmt.Sprintf("%s sizin ucun yeni vaxt teklif etdi: %s", parties.StaffName, proposedWhen)
		if event.Message != "" {
			body += " — " + event.Message
		}
		add(parties.CustomerUserID, notification.TypeBookingRescheduleProposed,
			"Alternativ vaxt teklifi", body)

	case booking.EventRescheduleAccepted:
		add(parties.StaffUserID, notification.TypeBookingRescheduleAccepted,
			"Teklif qebul edildi",
			fmt.Sprintf("%s yeni vaxti qebul etdi: %s", parties.CustomerName, when))

	case booking.EventRescheduleDeclined:
		add(parties.StaffUserID, notification.TypeBookingRescheduleDeclined,
			"Teklif redd edildi",
			fmt.Sprintf("%s teklifi redd etdi. Ilkin vaxt: %s", parties.CustomerName, when))

	case booking.EventCancelled:
		reason := event.Message
		if reason == "" {
			reason = "sebeb gosterilmedi"
		}
		add(parties.CustomerUserID, notification.TypeBookingCancelled,
			"Bron legv edildi",
			fmt.Sprintf("%s – %s legv edildi (%s).", when, service, reason))
		add(parties.StaffUserID, notification.TypeBookingCancelled,
			"Bron legv edildi",
			fmt.Sprintf("%s – %s legv edildi (%s).", parties.CustomerName, when, reason))

	case booking.EventCompleted:
		add(parties.CustomerUserID, notification.TypeBookingCompleted,
			"Randevu tamamlandi",
			fmt.Sprintf("%s – %s tamamlandi.", parties.StaffName, service))

	case booking.EventNoShow:
		add(parties.CustomerUserID, notification.TypeBookingNoShow,
			"Randevuya gelmediniz",
			fmt.Sprintf("%s tarixli randevu gelmeme kimi qeyd edildi.", when))
	}

	return envelopes
}

func (p *BookingPublisher) formatTime(value time.Time) string {
	return value.In(p.displayTZ).Format("02.01.2006 15:04")
}

func cloneMap(source notification.JSONMap) notification.JSONMap {
	cloned := make(notification.JSONMap, len(source))
	for key, value := range source {
		cloned[key] = value
	}
	return cloned
}
