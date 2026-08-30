// File: internal/infrastructure/postgres/dashboard_repo.go
package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// DashboardStats – admin panelinin yuxari kartlari.
//
// Butun deyerler biznesin oz timezone-unda deyil, serverin UTC-sinde
// hesablanir; gun serhedleri cagiran terefden gonderilir ki, "bu gun"
// istifadecinin gunu olsun.
type DashboardStats struct {
	// Bu gun
	TodayTotal     int `db:"today_total"     json:"today_total"`
	TodayPending   int `db:"today_pending"   json:"today_pending"`
	TodayConfirmed int `db:"today_confirmed" json:"today_confirmed"`

	// Butun vaxt uzre cavab gozleyenler
	PendingTotal int `db:"pending_total" json:"pending_total"`

	// Cari ay
	MonthCompleted int     `db:"month_completed" json:"month_completed"`
	MonthCancelled int     `db:"month_cancelled" json:"month_cancelled"`
	MonthRevenue   float64 `db:"month_revenue"   json:"month_revenue"`

	// Musteriler
	CustomersTotal int `db:"customers_total" json:"customers_total"`
	CustomersNew   int `db:"customers_new"   json:"customers_new"`
}

type DashboardRepository struct {
	database *sqlx.DB
}

func NewDashboardRepository(database *sqlx.DB) *DashboardRepository {
	return &DashboardRepository{database: database}
}

// GetStats – butun gostericileri bir sorguda hesablayir.
//
// Ayri-ayri COUNT sorgulari evezine FILTER islenir: baza cedveli bir
// defe oxuyur, panel ise 8 deyeri birden alir.
//
// Gelir yalniz "completed" bronlardan sayilir — tesdiqlenmis, lakin
// hele bas tutmamis randevu gelir deyil.
func (r *DashboardRepository) GetStats(
	ctx context.Context,
	businessID uuid.UUID,
	dayStart, dayEnd, monthStart time.Time,
) (*DashboardStats, error) {
	const query = `
		WITH booking_stats AS (
			SELECT
				COUNT(*) FILTER (
					WHERE start_time >= $2 AND start_time < $3
					  AND status <> 'cancelled'
				) AS today_total,

				COUNT(*) FILTER (
					WHERE start_time >= $2 AND start_time < $3
					  AND status = 'pending'
				) AS today_pending,

				COUNT(*) FILTER (
					WHERE start_time >= $2 AND start_time < $3
					  AND status = 'confirmed'
				) AS today_confirmed,

				COUNT(*) FILTER (WHERE status = 'pending') AS pending_total,

				COUNT(*) FILTER (
					WHERE status = 'completed' AND start_time >= $4
				) AS month_completed,

				COUNT(*) FILTER (
					WHERE status = 'cancelled' AND start_time >= $4
				) AS month_cancelled
			FROM bookings
			WHERE business_id = $1
		),
		revenue AS (
			SELECT COALESCE(SUM(s.price), 0) AS month_revenue
			FROM bookings b
			JOIN services s ON s.id = b.service_id
			WHERE b.business_id = $1
			  AND b.status = 'completed'
			  AND b.start_time >= $4
		),
		customer_stats AS (
			SELECT
				COUNT(*) AS customers_total,
				COUNT(*) FILTER (WHERE created_at >= $4) AS customers_new
			FROM customers
			WHERE business_id = $1
		)
		SELECT
			booking_stats.today_total,
			booking_stats.today_pending,
			booking_stats.today_confirmed,
			booking_stats.pending_total,
			booking_stats.month_completed,
			booking_stats.month_cancelled,
			revenue.month_revenue,
			customer_stats.customers_total,
			customer_stats.customers_new
		FROM booking_stats, revenue, customer_stats`

	var stats DashboardStats
	err := r.database.GetContext(ctx, &stats, query, businessID, dayStart, dayEnd, monthStart)
	if err != nil {
		return nil, fmt.Errorf("postgres: dashboard stats failed: %w", err)
	}
	return &stats, nil
}
