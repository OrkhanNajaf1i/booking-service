// File: internal/infrastructure/postgres/migrations.go
package postgres

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/OrkhanNajaf1i/booking-service/internal/config"
	"github.com/OrkhanNajaf1i/booking-service/internal/logger"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func RunMigrations(cfg config.AppConfig, appLogger logger.Logger) error {
	_, currentFile, _, _ := runtime.Caller(0)
	baseDir := filepath.Dir(currentFile)
	projectRoot := filepath.Join(baseDir, "..", "..", "..")
	migrationsDir := filepath.Join(projectRoot, "migrations")

	appLogger.Info("Migration path check", logger.Field{Key: "dir", Value: migrationsDir})

	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		return fmt.Errorf("migrations directory not found: %s (create it or move files)", migrationsDir)
	}

	migrationsPath := "file://" + filepath.ToSlash(migrationsDir)

	// ✅ DÜZƏLİŞ: x-schema=public silindi
	// search_path istifadə edə bilərsən, amma default public schema-dır
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName,
	)

	appLogger.Info("Migration DSN", logger.Field{Key: "dsn", Value: dsn})

	m, err := migrate.New(migrationsPath, dsn)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer m.Close()

	version, dirty, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		appLogger.Warn("Could not get migration version", logger.Field{Key: "error", Value: err.Error()})
	} else if err == nil {
		appLogger.Info("Current migration version",
			logger.Field{Key: "version", Value: version},
			logger.Field{Key: "dirty", Value: dirty},
		)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migration failed: %w", err)
	}

	appLogger.Info("Migrations completed successfully")
	return nil
}

// package postgres

// import (
// 	"fmt"

// 	"github.com/OrkhanNajaf1i/booking-service/internal/config"
// 	// "github.com/golang-migrate/migrate"
// 	"github.com/golang-migrate/migrate/v4"
// 	_ "github.com/golang-migrate/migrate/v4/database/postgres"
// 	_ "github.com/golang-migrate/migrate/v4/source/file"
// )

// func RunMigrations(cfg config.AppConfig) error {
// 	m, err := migrate.New("file://migrations", fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
// 		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName,
// 	))
// 	if err != nil {
// 		return err
// 	}
// 	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
// 		return err
// 	}
// 	return nil
// }

// File: internal/infrastructure/postgres/migrations.go
// File: internal/infrastructure/postgres/migrations.go
