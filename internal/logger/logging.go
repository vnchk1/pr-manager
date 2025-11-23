package logger

import (
	"log/slog"
	"os"
	"strings"
)

func NewLogger(lvlStr string) (logger *slog.Logger) {
	lvl := ConvertLogLvl(lvlStr)
	logHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: lvl,
	})
	logger = slog.New(logHandler)

	return
}

func ConvertLogLvl(lvl string) slog.Level {
	switch strings.ToLower(lvl) {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
