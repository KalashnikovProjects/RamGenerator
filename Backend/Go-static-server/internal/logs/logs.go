package logs

import (
	"log/slog"
	"os"
)

func InitStdoutLogs(logLevel slog.Level) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))
	slog.SetDefault(logger)
}
