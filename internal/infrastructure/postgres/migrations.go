// File: internal/infrastructure/postgres/migrations.go
package postgres

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/OrkhanNajaf1i/booking-service/internal/config"
	"github.com/OrkhanNajaf1i/booking-service/internal/logger"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// buildMigrationDSN:
// - Prioritet: cfg.DbDsn (APP_DB_DSN)
// - Yoxdursa: cfg.DBHost.. ilə DSN yığır
// - Supabase üçün sslmode=require
// - Transaction pooler (6543) üçün prepared statements söndürür
func buildMigrationDSN(cfg config.AppConfig) string {
	dsn := strings.TrimSpace(cfg.DbDsn)

	if dsn == "" {
		// fallback: env parçaları ilə DSN
		dsn = fmt.Sprintf(
			"postgres://%s:%s@%s:%s/%s?sslmode=require",
			cfg.DBUser,
			cfg.DBPassword,
			cfg.DBHost,
			cfg.DBPort,
			cfg.DBName,
		)
	} else {
		// DSN-də sslmode yoxdursa əlavə et
		if !strings.Contains(dsn, "sslmode=") {
			sep := "?"
			if strings.Contains(dsn, "?") {
				sep = "&"
			}
			dsn += sep + "sslmode=require"
		}
	}

	// Supabase transaction pooler (6543) prepared statements dəstəkləmir [web:131]
	// pgx üçün “simple protocol” ilə işləməyə məcbur edirik.
	if strings.Contains(dsn, "pooler.supabase.com:6543") && !strings.Contains(dsn, "default_query_exec_mode=") {
		sep := "?"
		if strings.Contains(dsn, "?") {
			sep = "&"
		}
		dsn += sep + "default_query_exec_mode=simple_protocol"
	}

	return dsn
}

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

	// ✅ FIX: DSN-i düzgün seçirik
	dsn := buildMigrationDSN(cfg)
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
