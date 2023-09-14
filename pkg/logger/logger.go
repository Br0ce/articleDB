package logger

import (
	"log/slog"
	"os"
)

func New(devLogger bool) *slog.Logger {
	var level = new(slog.LevelVar)
	lg := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
	lg = lg.With("service", slog.StringValue("articleDB"))

	if devLogger {
		level.Set(slog.LevelDebug)
	}

	return lg
}

func NewTest(debug bool) *slog.Logger {
	var level = new(slog.LevelVar)
	lg := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level, AddSource: true}))

	if debug {
		level.Set(slog.LevelDebug)
	}

	return lg
}
