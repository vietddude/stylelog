package slogx

import (
	"context"
	"log/slog"
	"os"

	"github.com/lmittmann/tint"
)

// LevelBasedHandler routes log records to different handlers based on level:
// - Info, Debug, Warn -> LowLevelHandler (no source, lighter output)
// - Error and above   -> ErrorHandler (with source, highlighted errors)
type LevelBasedHandler struct {
	LowLevelHandler slog.Handler
	ErrorHandler    slog.Handler
}

func (h *LevelBasedHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return true
}

func (h *LevelBasedHandler) Handle(ctx context.Context, r slog.Record) error {
	if r.Level >= slog.LevelError {
		return h.ErrorHandler.Handle(ctx, r)
	}
	return h.LowLevelHandler.Handle(ctx, r)
}

func (h *LevelBasedHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &LevelBasedHandler{
		LowLevelHandler: h.LowLevelHandler.WithAttrs(attrs),
		ErrorHandler:    h.ErrorHandler.WithAttrs(attrs),
	}
}

func (h *LevelBasedHandler) WithGroup(name string) slog.Handler {
	return &LevelBasedHandler{
		LowLevelHandler: h.LowLevelHandler.WithGroup(name),
		ErrorHandler:    h.ErrorHandler.WithGroup(name),
	}
}

// New returns a slog.Logger that:
// - uses a tint handler without source for Info/Debug/Warn
// - uses a tint handler with source (and red-colored "err"/"error" fields) for Error+
func New() *slog.Logger {
	lowLevelHandler := tint.NewHandler(os.Stderr, &tint.Options{
		AddSource: false,
	})

	errorHandler := tint.NewHandler(os.Stderr, &tint.Options{
		AddSource: true,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Color errors red
			if a.Key == "err" || a.Key == "error" {
				a = tint.Attr(9, a)
			}
			return a
		},
	})

	return slog.New(&LevelBasedHandler{
		LowLevelHandler: lowLevelHandler,
		ErrorHandler:    errorHandler,
	})
}

// InitDefault creates the logger from New, sets it as slog's default,
// and returns it for direct use.
func InitDefault() *slog.Logger {
	logger := New()
	slog.SetDefault(logger)
	return logger
}
