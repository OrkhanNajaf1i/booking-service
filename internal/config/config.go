package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type AppConfig struct {
	Port      int
	Host      string
	LogLevel  string
	JWTSecret string
	DbDsn     string

	DBUser     string
	DBPassword string
	DBHost     string
	DBPort     string
	DBName     string

	EncryptionKey  string
	EnableDebug    bool
	MaxConcurrency int

	SMTPHost string
	SMTPPort int
	SMTPUser string
	SMTPPass string
	SMTPFrom string
}

func Load() (*AppConfig, error) {
	cfg := &AppConfig{}
	var err error

	if err = LoadServerConfig(cfg); err != nil {
		return nil, fmt.Errorf("server config error: %w", err)
	}
	if err = LoadSecurityConfig(cfg); err != nil {
		return nil, fmt.Errorf("security config error: %w", err)
	}
	if err = LoadDatabaseConfig(cfg); err != nil {
		return nil, fmt.Errorf("database config error: %w", err)
	}
	if err = LoadEmailConfig(cfg); err != nil {
		return nil, fmt.Errorf("email config error: %w", err)
	}
	return cfg, nil
}

func LoadServerConfig(cfg *AppConfig) error {
	portStr := strings.TrimSpace(os.Getenv("PORT"))
	if portStr == "" {
		portStr = strings.TrimSpace(os.Getenv("APP_PORT"))
	}
	if portStr == "" {
		cfg.Port = 8080
	} else {
		port, err := strconv.Atoi(portStr)
		if err != nil {
			return fmt.Errorf("PORT/APP_PORT must be a number: %w", err)
		}
		cfg.Port = port
	}

	cfg.Host = strings.TrimSpace(os.Getenv("APP_HOST"))
	if cfg.Host == "" {
		cfg.Host = "0.0.0.0"
	}

	cfg.DbDsn = strings.TrimSpace(os.Getenv("APP_DB_DSN"))

	cfg.LogLevel = strings.TrimSpace(os.Getenv("APP_LOG_LEVEL"))
	if cfg.LogLevel == "" {
		cfg.LogLevel = "info"
	}
	return nil
}

func LoadSecurityConfig(cfg *AppConfig) error {
	cfg.JWTSecret = os.Getenv("APP_JWT_SECRET")
	cfg.EncryptionKey = os.Getenv("APP_ENCRYPTION_KEY")

	if cfg.JWTSecret == "" {
		return errors.New("JWT_SECRET is required but not set")
	}
	if cfg.EncryptionKey == "" {
		return errors.New("Encryption is required but not set")
	}
	return nil
}

func LoadDatabaseConfig(cfg *AppConfig) error {
	if strings.TrimSpace(cfg.DbDsn) != "" {
		return nil
	}

	cfg.DBUser = os.Getenv("APP_DB_USER")
	cfg.DBPassword = os.Getenv("APP_DB_PASSWORD")
	cfg.DBHost = os.Getenv("APP_DB_HOST")
	cfg.DBName = os.Getenv("APP_DB_NAME")
	cfg.DBPort = os.Getenv("APP_DB_PORT")

	if cfg.DBPort == "" {
		cfg.DBPort = "5432"
	}

	if cfg.DBUser == "" || cfg.DBPassword == "" || cfg.DBHost == "" || cfg.DBName == "" || cfg.DBPort == "" {
		return errors.New("missing required database environment variables.")
	}

	return nil
}

func LoadEmailConfig(cfg *AppConfig) error {
	cfg.SMTPHost = os.Getenv("SMTP_HOST")
	cfg.SMTPUser = os.Getenv("SMTP_USER")
	cfg.SMTPPass = os.Getenv("SMTP_PASS")
	cfg.SMTPFrom = os.Getenv("SMTP_FROM")

	portStr := os.Getenv("SMTP_PORT")
	if portStr == "" {
		cfg.SMTPPort = 587
	} else {
		port, err := strconv.Atoi(portStr)
		if err != nil {
			return fmt.Errorf("SMTP_PORT must be a number: %w", err)
		}
		cfg.SMTPPort = port
	}

	if cfg.SMTPHost == "" || cfg.SMTPUser == "" || cfg.SMTPPass == "" {
		return errors.New("missing required SMTP environment variables")
	}

	return nil
}
