package logger

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/OrkhanNajaf1i/booking-service/internal/config"
)

type Field struct {
	Key   string
	Value any
}
type Logger interface {
	Info(msg string, fields ...Field)
	Debug(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
}

type slogLogger struct {
	l *slog.Logger
}

func New(config *config.AppConfig) (Logger, error) {
	if config == nil {
		return nil, fmt.Errorf("Config is nil")
	}
	level, err := parseLevel(config.LogLevel)
	if err != nil {
		return nil, err
	}
	out := os.Stdout
	handler := slog.NewJSONHandler(out, &slog.HandlerOptions{
		Level: level,
	})
	base := slog.New(handler)
	var result Logger = &slogLogger{l: base}

	return result, nil

}

func (s *slogLogger) Info(msg string, fields ...Field) {
	s.l.Info(msg, toArgs(fields)...)
}
func (s *slogLogger) Debug(msg string, fields ...Field) {
	s.l.Debug(msg, toArgs(fields)...)
}
func (s *slogLogger) Error(msg string, fields ...Field) {
	s.l.Error(msg, toArgs(fields)...)
}
func (s *slogLogger) Warn(msg string, fields ...Field) {
	s.l.Warn(msg, toArgs(fields)...)
}

func (s *slogLogger) WithField(key string, value any) Logger {
	nl := s.l.With(key, value)
	return &slogLogger{l: nl}
}

func parseLevel(s string) (slog.Level, error) {
	lvl := strings.ToLower(strings.TrimSpace(s))

	switch lvl {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return 0, fmt.Errorf("invalid log level: %q", s)
	}
}
func toArgs(fields []Field) []any {
	args := make([]any, 0, len(fields)*2)
	for _, f := range fields {
		args = append(args, f.Key, f.Value)
	}
	return args
}
