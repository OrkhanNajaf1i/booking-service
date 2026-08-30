// File: internal/infrastructure/postgres/booking_repo.go
package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/OrkhanNajaf1i/booking-service/internal/domain/booking"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
)

// pgExclusionViolation – iki randevunun ust-uste dusmesi zamani
// bookings_no_overlap constraint-inin verdiyi SQLSTATE.
const pgExclusionViolation = "23P01"

type BookingRepository struct {
	database *sqlx.DB
}

func NewBookingRepository(database *sqlx.DB) *BookingRepository {
	return &BookingRepository{database: database}
}

const bookingColumns = `
	id, business_id, customer_id, staff_id, service_id, location_id, slot_id,
	start_time, end_time, duration_mins, status, notes,
	proposed_start_time, proposed_end_time, proposed_by, proposal_note, proposed_at,
	cancel_reason, cancelled_by, confirmed_at,
	created_at, updated_at`

// ============================================================
// CREATE
// ============================================================

// Create – bronu yazir.
//
// Eyni isci ucun ust-uste dusen aktiv bron olarsa Postgres exclusion
// constraint-i sorgunu redd edir. Iki musteri eyni saniyede eyni vaxta
// basanda mehz burada ikincisi ErrSlotTaken alir – availability
// yoxlamasi ile yazilis arasindaki yarisi baglayan yegane yer budur.
func (r *BookingRepository) Create(ctx context.Context, b *booking.Booking) error {
	query := `
		INSERT INTO bookings (
			id, business_id, customer_id, staff_id, service_id, location_id, slot_id,
			start_time, end_time, duration_mins, status, notes,
			proposed_start_time, proposed_end_time, proposed_by, proposal_note, proposed_at,
			cancel_reason, cancelled_by, confirmed_at,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7,
			$8, $9, $10, $11, $12,
			$13, $14, $15, $16, $17,
			$18, $19, $20,
			$21, $22
		)`

	_, err := r.database.ExecContext(
		ctx, query,
		b.ID, b.BusinessID, b.CustomerID, b.StaffID, b.ServiceID, b.LocationID, b.SlotID,
		b.StartTime, b.EndTime, b.DurationMins, string(b.Status), b.Notes,
		b.ProposedStartTime, b.ProposedEndTime, b.ProposedBy, b.ProposalNote, b.ProposedAt,
		b.CancelReason, b.CancelledBy, b.ConfirmedAt,
		b.CreatedAt, b.UpdatedAt,
	)
	if err != nil {
		if isOverlapViolation(err) {
			return booking.ErrSlotTaken
		}
		return fmt.Errorf("postgres: create booking failed: %w", err)
	}
	return nil
}

// isOverlapViolation – xeta ust-uste dusme constraint-indendirmi?
func isOverlapViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == pgExclusionViolation
	}
	// btree_gist qurulmayibsa constraint yoxdur; bu halda unikal slot
	// indeksinden gelen xetani da eyni cur qiymetlendiririk.
	return strings.Contains(strings.ToLower(err.Error()), "bookings_no_overlap")
}

// ============================================================
// READ
// ============================================================

func (r *BookingRepository) GetByID(
	ctx context.Context,
	businessID, bookingID uuid.UUID,
) (*booking.Booking, error) {
	query := `SELECT ` + bookingColumns + ` FROM bookings WHERE business_id = $1 AND id = $2`

	var found booking.Booking
	err := r.database.GetContext(ctx, &found, query, businessID, bookingID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("postgres: get booking failed: %w", err)
	}
	return &found, nil
}

// GetByIDForUser – musteri tetbiqi ucun: istifadeci hemin bronun
// musterisidirse tapilir (business konteksti teleb olunmur).
func (r *BookingRepository) GetByIDForUser(
	ctx context.Context,
	userID, bookingID uuid.UUID,
) (*booking.Booking, error) {
	query := `
		SELECT ` + prefixColumns("b", bookingColumns) + `
		FROM bookings b
		JOIN customers c ON c.id = b.customer_id
		WHERE b.id = $1 AND c.user_id = $2`

	var found booking.Booking
	err := r.database.GetContext(ctx, &found, query, bookingID, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("postgres: get booking for user failed: %w", err)
	}
	return &found, nil
}

// List – biznes daxilinde filtrli siyahi.
func (r *BookingRepository) List(
	ctx context.Context,
	businessID uuid.UUID,
	filter *booking.ListFilter,
) ([]*booking.Booking, error) {
	conditions := []string{"business_id = $1"}
	args := []interface{}{businessID}

	conditions, args = applyBookingFilter(conditions, args, filter, "")

	query := fmt.Sprintf(
		`SELECT %s FROM bookings WHERE %s ORDER BY start_time DESC LIMIT $%d OFFSET $%d`,
		bookingColumns,
		strings.Join(conditions, " AND "),
		len(args)+1, len(args)+2,
	)
	args = append(args, filter.Limit, filter.Offset)

	rows := make([]*booking.Booking, 0, filter.Limit)
	if err := r.database.SelectContext(ctx, &rows, query, args...); err != nil {
		return nil, fmt.Errorf("postgres: list bookings failed: %w", err)
	}
	return rows, nil
}

// ListForCustomerUser – istifadecinin butun bizneslerdeki bronlari.
func (r *BookingRepository) ListForCustomerUser(
	ctx context.Context,
	userID uuid.UUID,
	filter *booking.ListFilter,
) ([]*booking.Booking, error) {
	conditions := []string{"c.user_id = $1"}
	args := []interface{}{userID}

	conditions, args = applyBookingFilter(conditions, args, filter, "b.")

	query := fmt.Sprintf(
		`SELECT %s FROM bookings b
		 JOIN customers c ON c.id = b.customer_id
		 WHERE %s ORDER BY b.start_time DESC LIMIT $%d OFFSET $%d`,
		prefixColumns("b", bookingColumns),
		strings.Join(conditions, " AND "),
		len(args)+1, len(args)+2,
	)
	args = append(args, filter.Limit, filter.Offset)

	rows := make([]*booking.Booking, 0, filter.Limit)
	if err := r.database.SelectContext(ctx, &rows, query, args...); err != nil {
		return nil, fmt.Errorf("postgres: list customer bookings failed: %w", err)
	}
	return rows, nil
}

func (r *BookingRepository) CountByStatus(
	ctx context.Context,
	businessID uuid.UUID,
	status booking.BookingStatus,
) (int, error) {
	query := `SELECT COUNT(*) FROM bookings WHERE business_id = $1 AND status = $2`

	var count int
	if err := r.database.GetContext(ctx, &count, query, businessID, string(status)); err != nil {
		return 0, fmt.Errorf("postgres: count bookings failed: %w", err)
	}
	return count, nil
}

// ============================================================
// UPDATE
// ============================================================

func (r *BookingRepository) Update(ctx context.Context, b *booking.Booking) error {
	query := `
		UPDATE bookings SET
			staff_id            = $3,
			service_id          = $4,
			location_id         = $5,
			start_time          = $6,
			end_time            = $7,
			duration_mins       = $8,
			status              = $9,
			notes               = $10,
			proposed_start_time = $11,
			proposed_end_time   = $12,
			proposed_by         = $13,
			proposal_note       = $14,
			proposed_at         = $15,
			cancel_reason       = $16,
			cancelled_by        = $17,
			confirmed_at        = $18,
			updated_at          = $19
		WHERE business_id = $1 AND id = $2`

	result, err := r.database.ExecContext(
		ctx, query,
		b.BusinessID, b.ID,
		b.StaffID, b.ServiceID, b.LocationID,
		b.StartTime, b.EndTime, b.DurationMins, string(b.Status), b.Notes,
		b.ProposedStartTime, b.ProposedEndTime, b.ProposedBy, b.ProposalNote, b.ProposedAt,
		b.CancelReason, b.CancelledBy, b.ConfirmedAt,
		b.UpdatedAt,
	)
	if err != nil {
		if isOverlapViolation(err) {
			return booking.ErrSlotTaken
		}
		return fmt.Errorf("postgres: update booking failed: %w", err)
	}

	affected, err := result.RowsAffected()
	if err == nil && affected == 0 {
		return booking.ErrNotFound
	}
	return nil
}

// ============================================================
// HELPERS
// ============================================================

// applyBookingFilter – ortaq filter sertlerini elave edir.
// alias bos ve ya "b." ola biler (JOIN-li sorgular ucun).
func applyBookingFilter(
	conditions []string,
	args []interface{},
	filter *booking.ListFilter,
	alias string,
) ([]string, []interface{}) {
	if filter == nil {
		return conditions, args
	}

	if filter.StaffID != nil {
		args = append(args, *filter.StaffID)
		conditions = append(conditions, fmt.Sprintf("%sstaff_id = $%d", alias, len(args)))
	}
	if filter.CustomerID != nil {
		args = append(args, *filter.CustomerID)
		conditions = append(conditions, fmt.Sprintf("%scustomer_id = $%d", alias, len(args)))
	}
	if filter.Status != nil {
		args = append(args, string(*filter.Status))
		conditions = append(conditions, fmt.Sprintf("%sstatus = $%d", alias, len(args)))
	}
	if filter.From != nil {
		args = append(args, *filter.From)
		conditions = append(conditions, fmt.Sprintf("%sstart_time >= $%d", alias, len(args)))
	}
	if filter.To != nil {
		args = append(args, *filter.To)
		conditions = append(conditions, fmt.Sprintf("%sstart_time < $%d", alias, len(args)))
	}

	return conditions, args
}

// prefixColumns – "id, business_id" -> "b.id, b.business_id".
// JOIN-li sorgularda sutun adlarinin qarismamasi ucun.
func prefixColumns(alias, columns string) string {
	parts := strings.Split(columns, ",")
	prefixed := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		prefixed = append(prefixed, alias+"."+trimmed)
	}
	return strings.Join(prefixed, ", ")
}

// FindCustomerOverlap – istifadecinin verilmis araliqla kesisen aktiv bronu.
//
// customers cedvelinde her biznes ucun ayri setir olur, ona gore
// user_id uzre JOIN edirik: eyni sexsin ferqli bizneslerdeki randevulari
// da nezere alinir.
func (r *BookingRepository) FindCustomerOverlap(
	ctx context.Context,
	customerUserID uuid.UUID,
	start, end time.Time,
	excludeBookingID uuid.UUID,
) (*booking.Booking, error) {
	query := `
		SELECT ` + prefixColumns("b", bookingColumns) + `
		FROM bookings b
		JOIN customers c ON c.id = b.customer_id
		WHERE c.user_id = $1
		  AND b.status IN ('pending', 'confirmed', 'reschedule_proposed')
		  AND b.start_time < $3
		  AND b.end_time   > $2
		  AND ($4 = '00000000-0000-0000-0000-000000000000'::uuid OR b.id <> $4)
		LIMIT 1`

	var found booking.Booking
	err := r.database.GetContext(ctx, &found, query, customerUserID, start, end, excludeBookingID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("postgres: find customer overlap failed: %w", err)
	}
	return &found, nil
}

// ExpireStalePending – cavabsiz qalmis pending bronlari legv edir.
//
// Muddet her biznesin schedule_settings-inden goturulur (staff uzre
// override varsa o, yoxsa biznes default-u, o da yoxdursa 1440 deqiqe).
// Randevu vaxti onsuz da kecmisse muddetden asili olmayaraq legv olunur:
// bele bron slotu bos yere tutur.
func (r *BookingRepository) ExpireStalePending(
	ctx context.Context,
	limit int,
) ([]*booking.Booking, error) {
	query := `
		WITH expired AS (
			SELECT b.id
			FROM bookings b
			LEFT JOIN schedule_settings ss_staff
			       ON ss_staff.business_id = b.business_id
			      AND ss_staff.staff_id    = b.staff_id
			LEFT JOIN schedule_settings ss_biz
			       ON ss_biz.business_id = b.business_id
			      AND ss_biz.staff_id IS NULL
			WHERE b.status = 'pending'
			  AND COALESCE(ss_staff.pending_expires_mins, ss_biz.pending_expires_mins, 1440) > 0
			  AND (
			        b.created_at + make_interval(
			            mins => COALESCE(ss_staff.pending_expires_mins, ss_biz.pending_expires_mins, 1440)
			        ) <= NOW()
			     OR b.start_time <= NOW()
			  )
			ORDER BY b.created_at
			LIMIT $1
			FOR UPDATE SKIP LOCKED
		)
		UPDATE bookings b
		SET status        = 'cancelled',
		    cancel_reason = 'Vaxtinda cavablandirilmadi',
		    updated_at    = NOW()
		FROM expired
		WHERE b.id = expired.id
		RETURNING ` + prefixColumns("b", bookingColumns)

	rows := make([]*booking.Booking, 0, limit)
	if err := r.database.SelectContext(ctx, &rows, query, limit); err != nil {
		return nil, fmt.Errorf("postgres: expire stale pending failed: %w", err)
	}
	return rows, nil
}
