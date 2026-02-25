package logger

import "log/slog"

type Logger interface {
	Info(msg string, args ...any)
	Error(msg string, args ...any)
	Warn(msg string, args ...any)
	With(args ...any) *slog.Logger
}
