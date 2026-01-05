package postgres

import (
	"database/sql"
	"fmt"

	"github.com/OrkhanNajaf1i/booking-service/internal/config"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func New(cfg config.AppConfig) (*sql.DB, error) {
	// dsn := fmt.Sprintf(
	// 	"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
	// 	cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName,
	// )
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName,
		// DBPort STRING olaraq qalır ✅ (bu düzgün)
	)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
