package logger

import (
	"log/slog"
	"os"
)

func InitLogs() {
	logLevel := slog.LevelInfo
	if os.Getenv("DEBUG") == "1" {
		logLevel = slog.LevelDebug
	}
	h := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel})
	slog.SetDefault(slog.New(h))
}
