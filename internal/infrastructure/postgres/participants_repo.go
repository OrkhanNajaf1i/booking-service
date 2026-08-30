// File: internal/infrastructure/postgres/participants_repo.go
package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/OrkhanNajaf1i/booking-service/internal/domain/booking"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// ParticipantsRepository – bildiris ucun terefleri bir sorguda toplayir.
// booking.ParticipantResolver realizasiyasidir.
type ParticipantsRepository struct {
	database *sqlx.DB
}

func NewParticipantsRepository(database *sqlx.DB) *ParticipantsRepository {
	return &ParticipantsRepository{database: database}
}

// participantsRow – sorgunun xam neticesi.
// CustomerUserID NULL ola biler: musteri sistemde qeydiyyatdan
// kecmeyib, provider onu elle elave edib.
type participantsRow struct {
	BusinessName   string         `db:"business_name"`
	StaffUserID    uuid.UUID      `db:"staff_user_id"`
	StaffName      string         `db:"staff_name"`
	CustomerUserID *uuid.UUID     `db:"customer_user_id"`
	CustomerName   string         `db:"customer_name"`
	ServiceName    sql.NullString `db:"service_name"`
}

// Resolve – isci, musteri, biznes ve (varsa) xidmet adlarini qaytarir.
func (r *ParticipantsRepository) Resolve(
	ctx context.Context,
	businessID, staffID, customerID uuid.UUID,
	serviceID *uuid.UUID,
) (*booking.Participants, error) {
	query := `
		SELECT
			b.name         AS business_name,
			su.id          AS staff_user_id,
			su.full_name   AS staff_name,
			c.user_id      AS customer_user_id,
			c.full_name    AS customer_name,
			s.name         AS service_name
		FROM businesses b
		JOIN staff_profiles sp ON sp.id = $2 AND sp.business_id = b.id
		JOIN users          su ON su.id = sp.user_id
		JOIN customers      c  ON c.id  = $3 AND c.business_id = b.id
		LEFT JOIN services  s  ON s.id  = $4 AND s.business_id = b.id
		WHERE b.id = $1`

	var serviceArg interface{}
	if serviceID != nil && *serviceID != uuid.Nil {
		serviceArg = *serviceID
	} else {
		serviceArg = nil
	}

	var row participantsRow
	err := r.database.GetContext(ctx, &row, query, businessID, staffID, customerID, serviceArg)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("postgres: participants tapilmadi (business=%s staff=%s customer=%s)",
			businessID, staffID, customerID)
	}
	if err != nil {
		return nil, fmt.Errorf("postgres: resolve participants failed: %w", err)
	}

	participants := &booking.Participants{
		BusinessName: row.BusinessName,
		StaffUserID:  row.StaffUserID,
		StaffName:    row.StaffName,
		CustomerName: row.CustomerName,
	}
	if row.CustomerUserID != nil {
		participants.CustomerUserID = *row.CustomerUserID
	}
	if row.ServiceName.Valid {
		participants.ServiceName = row.ServiceName.String
	}

	return participants, nil
}

// BusinessIDForStaff – iscinin aid oldugu biznesin ID-si.
// Musteri bron edende biznes konteksti mehz buradan gelir.
func (r *ParticipantsRepository) BusinessIDForStaff(
	ctx context.Context,
	staffID uuid.UUID,
) (uuid.UUID, error) {
	query := `SELECT business_id FROM staff_profiles WHERE id = $1 AND status = 'active'`

	var businessID uuid.UUID
	err := r.database.GetContext(ctx, &businessID, query, staffID)
	if errors.Is(err, sql.ErrNoRows) {
		return uuid.Nil, fmt.Errorf("postgres: aktiv isci tapilmadi: %s", staffID)
	}
	if err != nil {
		return uuid.Nil, fmt.Errorf("postgres: business for staff failed: %w", err)
	}
	return businessID, nil
}
