// File: internal/infrastructure/postgres/otp_repo.go
package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/OrkhanNajaf1i/booking-service/internal/domain/otp"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type OTPRepository struct {
	database *sqlx.DB
}

func NewOTPRepository(database *sqlx.DB) *OTPRepository {
	return &OTPRepository{database: database}
}

func (r *OTPRepository) Create(ctx context.Context, verification *otp.Verification) error {
	query := `
		INSERT INTO phone_verifications (
			id, phone_e164, code_hash, channel, attempts, expires_at, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.database.ExecContext(
		ctx, query,
		verification.ID,
		verification.PhoneE164,
		verification.CodeHash,
		string(verification.Channel),
		verification.Attempts,
		verification.ExpiresAt,
		verification.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("postgres: failed to insert phone verification: %w", err)
	}

	return nil
}

// LatestActive – hemin nomre ucun hele islenmemis son kod.
//
// Kohnelmis kodlar da qaytarilir: "vaxti bitib, yeniden isteyin"
// mesaji "bele kod yoxdur"dan daha aydindir.
func (r *OTPRepository) LatestActive(
	ctx context.Context,
	phoneE164 string,
) (*otp.Verification, error) {
	query := `
		SELECT id, phone_e164, code_hash, channel, attempts,
		       expires_at, consumed_at, created_at
		FROM phone_verifications
		WHERE phone_e164 = $1 AND consumed_at IS NULL
		ORDER BY created_at DESC
		LIMIT 1
	`

	var verification otp.Verification
	err := r.database.GetContext(ctx, &verification, query, phoneE164)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("postgres: failed to get phone verification: %w", err)
	}

	return &verification, nil
}

func (r *OTPRepository) CountSince(
	ctx context.Context,
	phoneE164 string,
	since time.Time,
) (int, error) {
	var count int

	err := r.database.GetContext(
		ctx, &count,
		`SELECT COUNT(*) FROM phone_verifications
		 WHERE phone_e164 = $1 AND created_at >= $2`,
		phoneE164, since,
	)
	if err != nil {
		return 0, fmt.Errorf("postgres: failed to count phone verifications: %w", err)
	}

	return count, nil
}

func (r *OTPRepository) IncrementAttempts(ctx context.Context, id uuid.UUID) error {
	_, err := r.database.ExecContext(
		ctx,
		`UPDATE phone_verifications SET attempts = attempts + 1 WHERE id = $1`,
		id,
	)
	if err != nil {
		return fmt.Errorf("postgres: failed to increment attempts: %w", err)
	}

	return nil
}

func (r *OTPRepository) MarkConsumed(ctx context.Context, id uuid.UUID) error {
	_, err := r.database.ExecContext(
		ctx,
		`UPDATE phone_verifications SET consumed_at = NOW() WHERE id = $1`,
		id,
	)
	if err != nil {
		return fmt.Errorf("postgres: failed to consume verification: %w", err)
	}

	return nil
}
