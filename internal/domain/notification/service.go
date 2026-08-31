// File: internal/domain/notification/service.go
package notification

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/OrkhanNajaf1i/booking-service/internal/logger"
	"github.com/google/uuid"
)

type notificationService struct {
	repo     Repository
	realtime RealtimePublisher
	push     PushSender
	log      logger.Logger
	now      func() time.Time
}

// NewService – bildiris servisi. realtime ve push nil ola biler;
// bu halda hemin kanal sadece atlanir.
func NewService(
	repo Repository,
	realtime RealtimePublisher,
	push PushSender,
	log logger.Logger,
) Service {
	return &notificationService{
		repo:     repo,
		realtime: realtime,
		push:     push,
		log:      log,
		now:      time.Now,
	}
}

// ============================================================
// DISPATCH
// ============================================================

// Dispatch – hadiseni uc kanala paylayir.
//
// Ardicilliq qesdlidir: evvelce baza (bildiris itmesin), sonra WS
// (aninda gorunsun), en sonda outbox (push worker-e qalsin).
// WS ve outbox xetalari hadiseni ucurmur – yalniz loglanir.
func (s *notificationService) Dispatch(ctx context.Context, envelope *Envelope) error {
	if envelope == nil {
		return NewError("INVALID_ENVELOPE", "envelope bosdur")
	}
	if envelope.UserID == uuid.Nil {
		return NewError("INVALID_RECIPIENT", "user_id bosdur")
	}
	if envelope.CreatedAt.IsZero() {
		envelope.CreatedAt = s.now()
	}
	if envelope.Payload == nil {
		envelope.Payload = JSONMap{}
	}

	item := &Notification{
		ID:         uuid.New(),
		BusinessID: envelope.BusinessID,
		UserID:     envelope.UserID,
		BookingID:  envelope.BookingID,
		Type:       envelope.Type,
		Title:      envelope.Title,
		Body:       envelope.Body,
		Payload:    envelope.Payload,
		IsRead:     false,
		CreatedAt:  envelope.CreatedAt,
	}

	if err := s.repo.Create(ctx, item); err != nil {
		return fmt.Errorf("bildiris yazilmadi: %w", err)
	}

	// Client bildirisi ID ile isareleye bilsin.
	envelope.Payload["notification_id"] = item.ID.String()

	if s.realtime != nil {
		if err := s.realtime.PublishToUser(ctx, envelope.UserID, envelope); err != nil {
			s.log.Warn("Realtime yayimi ugursuz",
				logger.Field{Key: "user_id", Value: envelope.UserID.String()},
				logger.Field{Key: "type", Value: string(envelope.Type)},
				logger.Field{Key: "error", Value: err.Error()},
			)
		}
	}

	if !envelope.SkipPush && s.push != nil && s.push.Enabled() {
		outbox := &OutboxItem{
			ID:        uuid.New(),
			UserID:    envelope.UserID,
			BookingID: envelope.BookingID,
			Channel:   ChannelPush,
			Type:      envelope.Type,
			Payload: JSONMap{
				"title":   envelope.Title,
				"body":    envelope.Body,
				"type":    string(envelope.Type),
				"payload": envelope.Payload,
			},
			Status:      OutboxPending,
			ScheduledAt: s.now(),
			CreatedAt:   s.now(),
		}
		if err := s.repo.EnqueueOutbox(ctx, outbox); err != nil {
			s.log.Warn("Push novbeye salinmadi",
				logger.Field{Key: "user_id", Value: envelope.UserID.String()},
				logger.Field{Key: "error", Value: err.Error()},
			)
		}
	}

	return nil
}

// DispatchMany – bir hadisenin bir nece aliciya getmesi ucun.
// Bir alicida xeta olsa qalanlari yene de gonderilir.
func (s *notificationService) DispatchMany(ctx context.Context, envelopes []*Envelope) error {
	var firstErr error
	for _, envelope := range envelopes {
		if err := s.Dispatch(ctx, envelope); err != nil {
			s.log.Error("Bildiris gonderilmedi",
				logger.Field{Key: "type", Value: string(envelope.Type)},
				logger.Field{Key: "error", Value: err.Error()},
			)
			if firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}

// ============================================================
// READ SIDE
// ============================================================

func (s *notificationService) List(
	ctx context.Context,
	userID uuid.UUID,
	unreadOnly bool,
	limit, offset int,
) ([]*Notification, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return s.repo.ListByUser(ctx, userID, unreadOnly, limit, offset)
}

func (s *notificationService) CountUnread(ctx context.Context, userID uuid.UUID) (int, error) {
	return s.repo.CountUnread(ctx, userID)
}

func (s *notificationService) MarkRead(ctx context.Context, userID, notificationID uuid.UUID) error {
	return s.repo.MarkRead(ctx, userID, notificationID)
}

func (s *notificationService) MarkAllRead(ctx context.Context, userID uuid.UUID) error {
	return s.repo.MarkAllRead(ctx, userID)
}

// ============================================================
// DEVICE TOKENS
// ============================================================

func (s *notificationService) RegisterDevice(
	ctx context.Context,
	userID uuid.UUID,
	req *RegisterDeviceRequest,
) (*DeviceToken, error) {
	if req == nil || strings.TrimSpace(req.Token) == "" {
		return nil, ErrInvalidToken
	}
	if !req.Platform.IsValid() {
		return nil, ErrInvalidPlatform
	}

	now := s.now()
	token := &DeviceToken{
		ID:        uuid.New(),
		UserID:    userID,
		Token:     strings.TrimSpace(req.Token),
		Platform:  req.Platform,
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.repo.SaveDeviceToken(ctx, token); err != nil {
		return nil, fmt.Errorf("device token yazilmadi: %w", err)
	}
	return token, nil
}

func (s *notificationService) UnregisterDevice(ctx context.Context, token string) error {
	if strings.TrimSpace(token) == "" {
		return ErrInvalidToken
	}
	return s.repo.DeactivateDeviceToken(ctx, strings.TrimSpace(token))
}

// ============================================================
// OUTBOX (worker)
// ============================================================

// ProcessOutbox – pending push-lari goturur ve gonderir.
// Alicinin aktiv token-i yoxdursa setir "sent" isarelenir –
// yeniden cehd etmeye deymez.
func (s *notificationService) ProcessOutbox(ctx context.Context, batchSize int) (int, error) {
	if batchSize <= 0 || batchSize > 200 {
		batchSize = 50
	}

	items, err := s.repo.ClaimOutbox(ctx, batchSize)
	if err != nil {
		return 0, fmt.Errorf("outbox oxunmadi: %w", err)
	}
	if len(items) == 0 {
		return 0, nil
	}

	sent := 0
	for _, item := range items {
		if s.push == nil || !s.push.Enabled() {
			_ = s.repo.MarkOutboxSent(ctx, item.ID)
			continue
		}

		tokens, err := s.repo.ListActiveDeviceTokens(ctx, item.UserID)
		if err != nil {
			_ = s.repo.MarkOutboxFailed(ctx, item.ID, err.Error())
			continue
		}
		if len(tokens) == 0 {
			_ = s.repo.MarkOutboxSent(ctx, item.ID)
			continue
		}

		raw := make([]string, 0, len(tokens))
		for _, token := range tokens {
			raw = append(raw, token.Token)
		}

		envelope := &Envelope{
			Type:      item.Type,
			UserID:    item.UserID,
			BookingID: item.BookingID,
			Title:     stringFrom(item.Payload, "title"),
			Body:      stringFrom(item.Payload, "body"),
			Payload:   item.Payload,
			CreatedAt: item.CreatedAt,
		}

		if err := s.push.Send(ctx, raw, envelope); err != nil {
			_ = s.repo.MarkOutboxFailed(ctx, item.ID, err.Error())
			s.log.Warn("Push gonderilmedi",
				logger.Field{Key: "outbox_id", Value: item.ID.String()},
				logger.Field{Key: "error", Value: err.Error()},
			)
			continue
		}

		_ = s.repo.MarkOutboxSent(ctx, item.ID)
		sent++
	}

	return sent, nil
}

func stringFrom(payload JSONMap, key string) string {
	if payload == nil {
		return ""
	}
	if value, ok := payload[key].(string); ok {
		return value
	}
	return ""
}
