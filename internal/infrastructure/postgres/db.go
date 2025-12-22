// package postgres

// import (
// 	"context"
// 	"database/sql"
// 	"fmt"
// 	"time"

// 	"github.com/OrkhanNajaf1i/booking-service/internal/config"
// 	_ "github.com/jackc/pgx/v5/stdlib"
// )

// func New(cfg config.AppConfig) (*sql.DB, error) {
// 	dsn := fmt.Sprintf("postgress://%s:%s@%s:%s/%s?sslmode=disable", cfg.DBUser, cfg, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)
// 	db, err := sql.Open("pgx", dsn)
// 	if err != nil {
// 		return nil, err
// 	}
// 	db.SetMaxOpenConns(25)
// 	db.SetMaxIdleConns(10)
// 	db.SetConnMaxLifetime(5 * time.Minute)
// 	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
// 	defer cancel()
// 	if err := db.PingContext(ctx); err != nil {
// 		return nil, fmt.Errorf("Database ping failed: %w", err)
// 	}
// 	return db, nil
// }

// internal/infrastructure/postgres/db.go
package postgres

import (
	"database/sql"
	"fmt"

	"github.com/OrkhanNajaf1i/booking-service/internal/config"
	_ "github.com/jackc/pgx/v5/stdlib" // Import for side-effect (registration)
)

func New(cfg config.AppConfig) (*sql.DB, error) {
	// sql.Register silinib!

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName,
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
