package logger

import (
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
)

func New() *slog.Logger {
	return NewWithLevel(slog.LevelInfo)
}

func NewWithLevel(level slog.Level) *slog.Logger {
	return slog.New(tint.NewTextHandler(os.Stderr, &tint.Options{
		Level:      level,
		TimeFormat: time.RFC3339,
	}))
}

var Default = New()
