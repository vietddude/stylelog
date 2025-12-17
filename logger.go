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
	if level >= slog.LevelError {
		return h.ErrorHandler.Enabled(ctx, level)
	}
	return h.LowLevelHandler.Enabled(ctx, level)
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
// - optionally applies the provided tint.Options (if any) to both handlers
//
// If opts is provided, the first element is used as the base options for both
// handlers. The following fields are still controlled by slogx:
//   - low-level handler:  AddSource is forced to false
//   - error handler:      AddSource is forced to true and its ReplaceAttr is
//     wrapped to also color "err"/"error" attributes red.
func New(opts ...*tint.Options) *slog.Logger {
	// Start from zero-value options, or from the user-provided base options.
	var baseOpts tint.Options
	if len(opts) > 0 && opts[0] != nil {
		baseOpts = *opts[0]
	}

	// Low-level handler: same as base, but without source.
	lowOpts := baseOpts
	lowOpts.AddSource = false
	lowLevelHandler := tint.NewHandler(os.Stderr, &lowOpts)

	// Error handler: same as base, but with source and enhanced ReplaceAttr.
	errOpts := baseOpts
	errOpts.AddSource = true
	userReplace := errOpts.ReplaceAttr
	errOpts.ReplaceAttr = func(groups []string, a slog.Attr) slog.Attr {
		// Let the user-supplied ReplaceAttr run first, if present.
		if userReplace != nil {
			a = userReplace(groups, a)
		}
		// Then color errors red.
		if a.Key == "err" || a.Key == "error" {
			a = tint.Attr(9, a)
		}
		return a
	}
	errorHandler := tint.NewHandler(os.Stderr, &errOpts)

	return slog.New(&LevelBasedHandler{
		LowLevelHandler: lowLevelHandler,
		ErrorHandler:    errorHandler,
	})
}

// InitDefault creates the logger from New, sets it as slog's default,
// and returns it for direct use.
func InitDefault(opts ...*tint.Options) *slog.Logger {
	logger := New(opts...)
	slog.SetDefault(logger)
	return logger
}
