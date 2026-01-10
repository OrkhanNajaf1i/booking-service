// package postgres

// import (
// 	"database/sql"
// 	"fmt"

// 	"github.com/OrkhanNajaf1i/booking-service/internal/config"
// 	_ "github.com/jackc/pgx/v5/stdlib"
// )

// func New(cfg config.AppConfig) (*sql.DB, error) {
// 	// dsn := fmt.Sprintf(
// 	// 	"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
// 	// 	cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName,
// 	// )
// 	dsn := fmt.Sprintf(
// 		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
// 		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName,
// 		// DBPort STRING olaraq qalır ✅ (bu düzgün)
// 	)

// 	db, err := sql.Open("pgx", dsn)
// 	if err != nil {
// 		return nil, err
// 	}

// 	if err := db.Ping(); err != nil {
// 		return nil, err
// 	}

//		return db, nil
//	}
package postgres

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/OrkhanNajaf1i/booking-service/internal/config"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func New(cfg config.AppConfig) (*sql.DB, error) {
	// 1) Əgər APP_DB_DSN verilibsə, ona prioritet ver
	dsn := strings.TrimSpace(cfg.DbDsn)

	if dsn == "" {
		// 2) Yoxdursa env-lərdən DSN yığ
		// Supabase üçün sslmode=disable olmaz; require istifadə edirik.
		dsn = fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=require",
			cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName,
		)
	} else {
		// 3) URL DSN-də sslmode yoxdursa, əlavə et (Supabase üçün)
		if !strings.Contains(dsn, "sslmode=") {
			sep := "?"
			if strings.Contains(dsn, "?") {
				sep = "&"
			}
			dsn += sep + "sslmode=require"
		}

		// 4) Transaction pooler (6543) üçün prepared statements söndür (pgx)
		// Supabase transaction pooler prepared statements dəstəkləmir.
		if strings.Contains(dsn, "pooler.supabase.com:6543") && !strings.Contains(dsn, "default_query_exec_mode=") {
			sep := "?"
			if strings.Contains(dsn, "?") {
				sep = "&"
			}
			dsn += sep + "default_query_exec_mode=simple_protocol"
		}
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	// Render-də ilk request-dən əvvəl problem çıxmasın deyə ping
	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
