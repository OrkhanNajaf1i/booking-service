// File: internal/infrastructure/postgres/notification_repo.go
package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/OrkhanNajaf1i/booking-service/internal/domain/notification"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type NotificationRepository struct {
	database *sqlx.DB
}

func NewNotificationRepository(database *sqlx.DB) *NotificationRepository {
	return &NotificationRepository{database: database}
}

// ============================================================
// IN-APP
// ============================================================

const notificationColumns = `
	id, business_id, user_id, booking_id, type, title, body,
	payload, is_read, read_at, created_at`

func (r *NotificationRepository) Create(ctx context.Context, item *notification.Notification) error {
	query := `
		INSERT INTO notifications (
			id, business_id, user_id, booking_id, type, title, body,
			payload, is_read, read_at, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	_, err := r.database.ExecContext(
		ctx, query,
		item.ID, item.BusinessID, item.UserID, item.BookingID,
		string(item.Type), item.Title, item.Body,
		item.Payload, item.IsRead, item.ReadAt, item.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("postgres: create notification failed: %w", err)
	}
	return nil
}

func (r *NotificationRepository) ListByUser(
	ctx context.Context,
	userID uuid.UUID,
	unreadOnly bool,
	limit, offset int,
) ([]*notification.Notification, error) {
	query := `SELECT ` + notificationColumns + ` FROM notifications WHERE user_id = $1`
	if unreadOnly {
		query += ` AND is_read = FALSE`
	}
	query += ` ORDER BY created_at DESC LIMIT $2 OFFSET $3`

	rows := make([]*notification.Notification, 0, limit)
	if err := r.database.SelectContext(ctx, &rows, query, userID, limit, offset); err != nil {
		return nil, fmt.Errorf("postgres: list notifications failed: %w", err)
	}
	return rows, nil
}

func (r *NotificationRepository) CountUnread(ctx context.Context, userID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND is_read = FALSE`

	var count int
	if err := r.database.GetContext(ctx, &count, query, userID); err != nil {
		return 0, fmt.Errorf("postgres: count unread failed: %w", err)
	}
	return count, nil
}

func (r *NotificationRepository) MarkRead(ctx context.Context, userID, notificationID uuid.UUID) error {
	query := `
		UPDATE notifications
		SET is_read = TRUE, read_at = NOW()
		WHERE user_id = $1 AND id = $2 AND is_read = FALSE`

	if _, err := r.database.ExecContext(ctx, query, userID, notificationID); err != nil {
		return fmt.Errorf("postgres: mark read failed: %w", err)
	}
	return nil
}

func (r *NotificationRepository) MarkAllRead(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE notifications
		SET is_read = TRUE, read_at = NOW()
		WHERE user_id = $1 AND is_read = FALSE`

	if _, err := r.database.ExecContext(ctx, query, userID); err != nil {
		return fmt.Errorf("postgres: mark all read failed: %w", err)
	}
	return nil
}

// ============================================================
// DEVICE TOKENS
// ============================================================

// SaveDeviceToken – eyni token basqa istifadeciye kecmis ola biler
// (paylasilan telefon, cixis-giris), ona gore user_id de yenilenir.
func (r *NotificationRepository) SaveDeviceToken(ctx context.Context, token *notification.DeviceToken) error {
	query := `
		INSERT INTO device_tokens (id, user_id, token, platform, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (token) DO UPDATE SET
			user_id    = EXCLUDED.user_id,
			platform   = EXCLUDED.platform,
			is_active  = TRUE,
			updated_at = EXCLUDED.updated_at
		RETURNING id`

	var id uuid.UUID
	err := r.database.QueryRowxContext(
		ctx, query,
		token.ID, token.UserID, token.Token, string(token.Platform),
		token.IsActive, token.CreatedAt, token.UpdatedAt,
	).Scan(&id)
	if err != nil {
		return fmt.Errorf("postgres: save device token failed: %w", err)
	}

	token.ID = id
	return nil
}

func (r *NotificationRepository) DeactivateDeviceToken(ctx context.Context, token string) error {
	query := `UPDATE device_tokens SET is_active = FALSE, updated_at = NOW() WHERE token = $1`
	if _, err := r.database.ExecContext(ctx, query, token); err != nil {
		return fmt.Errorf("postgres: deactivate device token failed: %w", err)
	}
	return nil
}

func (r *NotificationRepository) ListActiveDeviceTokens(
	ctx context.Context,
	userID uuid.UUID,
) ([]*notification.DeviceToken, error) {
	query := `
		SELECT id, user_id, token, platform, is_active, created_at, updated_at
		FROM device_tokens
		WHERE user_id = $1 AND is_active = TRUE`

	rows := make([]*notification.DeviceToken, 0, 4)
	if err := r.database.SelectContext(ctx, &rows, query, userID); err != nil {
		return nil, fmt.Errorf("postgres: list device tokens failed: %w", err)
	}
	return rows, nil
}

// ============================================================
// OUTBOX
// ============================================================

func (r *NotificationRepository) EnqueueOutbox(ctx context.Context, item *notification.OutboxItem) error {
	query := `
		INSERT INTO notification_outbox (
			id, user_id, booking_id, channel, type, payload,
			status, attempts, scheduled_at, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err := r.database.ExecContext(
		ctx, query,
		item.ID, item.UserID, item.BookingID, string(item.Channel), string(item.Type),
		item.Payload, string(item.Status), item.Attempts, item.ScheduledAt, item.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("postgres: enqueue outbox failed: %w", err)
	}
	return nil
}

// ClaimOutbox – bir nece worker paralel islese de eyni setiri iki defe
// goturmesin deye FOR UPDATE SKIP LOCKED istifade olunur. Goturulen
// setirler derhal "attempts+1" ile isarelenir.
func (r *NotificationRepository) ClaimOutbox(
	ctx context.Context,
	limit int,
) ([]*notification.OutboxItem, error) {
	query := `
		WITH claimed AS (
			SELECT id
			FROM notification_outbox
			WHERE status = 'pending' AND scheduled_at <= NOW()
			ORDER BY scheduled_at
			LIMIT $1
			FOR UPDATE SKIP LOCKED
		)
		UPDATE notification_outbox o
		SET attempts = o.attempts + 1
		FROM claimed
		WHERE o.id = claimed.id
		RETURNING o.id, o.user_id, o.booking_id, o.channel, o.type, o.payload,
		          o.status, o.attempts, o.last_error, o.scheduled_at, o.sent_at, o.created_at`

	rows := make([]*notification.OutboxItem, 0, limit)
	if err := r.database.SelectContext(ctx, &rows, query, limit); err != nil {
		return nil, fmt.Errorf("postgres: claim outbox failed: %w", err)
	}
	return rows, nil
}

func (r *NotificationRepository) MarkOutboxSent(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE notification_outbox SET status = 'sent', sent_at = $2 WHERE id = $1`
	if _, err := r.database.ExecContext(ctx, query, id, time.Now()); err != nil {
		return fmt.Errorf("postgres: mark outbox sent failed: %w", err)
	}
	return nil
}

// MarkOutboxFailed – 5 cehdden sonra setir "failed" olur ve daha
// goturulmur; ondan evvel exponential geriye cekilme ile yeniden novbeye dusur.
func (r *NotificationRepository) MarkOutboxFailed(ctx context.Context, id uuid.UUID, reason string) error {
	query := `
		UPDATE notification_outbox
		SET last_error   = $2,
		    status       = CASE WHEN attempts >= 5 THEN 'failed' ELSE 'pending' END,
		    scheduled_at = NOW() + (LEAST(attempts, 5) * INTERVAL '1 minute')
		WHERE id = $1`

	if _, err := r.database.ExecContext(ctx, query, id, reason); err != nil {
		return fmt.Errorf("postgres: mark outbox failed: %w", err)
	}
	return nil
}
