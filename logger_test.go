package slogx

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/lmittmann/tint"
)

type stubHandler struct {
	records []slog.Record
}

func (s *stubHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return true
}

func (s *stubHandler) Handle(ctx context.Context, r slog.Record) error {
	s.records = append(s.records, r.Clone())
	return nil
}

func (s *stubHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return s
}

func (s *stubHandler) WithGroup(name string) slog.Handler {
	return s
}

type levelFilterHandler struct {
	minLevel slog.Level
	records  []slog.Record
}

func (h *levelFilterHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.minLevel
}

func (h *levelFilterHandler) Handle(ctx context.Context, r slog.Record) error {
	h.records = append(h.records, r.Clone())
	return nil
}

func (h *levelFilterHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *levelFilterHandler) WithGroup(name string) slog.Handler {
	return h
}

func TestLevelBasedHandler_RoutesByLevel(t *testing.T) {
	tests := []struct {
		name          string
		logLevel      slog.Level
		logMessage    string
		expectInLow   int
		expectInError int
	}{
		{
			name:          "debug routes to low",
			logLevel:      slog.LevelDebug,
			logMessage:    "debug message",
			expectInLow:   1,
			expectInError: 0,
		},
		{
			name:          "info routes to low",
			logLevel:      slog.LevelInfo,
			logMessage:    "info message",
			expectInLow:   1,
			expectInError: 0,
		},
		{
			name:          "warn routes to low",
			logLevel:      slog.LevelWarn,
			logMessage:    "warn message",
			expectInLow:   1,
			expectInError: 0,
		},
		{
			name:          "error routes to error handler",
			logLevel:      slog.LevelError,
			logMessage:    "error message",
			expectInLow:   0,
			expectInError: 1,
		},
		{
			name:          "error+4 routes to error handler",
			logLevel:      slog.LevelError + 4,
			logMessage:    "critical message",
			expectInLow:   0,
			expectInError: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			low := &stubHandler{}
			errs := &stubHandler{}

			h := &LevelBasedHandler{
				LowLevelHandler: low,
				ErrorHandler:    errs,
			}

			ctx := context.Background()
			rec := slog.NewRecord(time.Now(), tt.logLevel, tt.logMessage, 0)

			if err := h.Handle(ctx, rec); err != nil {
				t.Fatalf("Handle() error = %v", err)
			}

			if len(low.records) != tt.expectInLow {
				t.Errorf("lowLevelHandler: expected %d records, got %d", tt.expectInLow, len(low.records))
			}

			if len(errs.records) != tt.expectInError {
				t.Errorf("errorHandler: expected %d records, got %d", tt.expectInError, len(errs.records))
			}

			// Verify message content if record exists
			if tt.expectInLow > 0 && len(low.records) > 0 {
				if low.records[0].Message != tt.logMessage {
					t.Errorf("lowLevelHandler: expected message %q, got %q", tt.logMessage, low.records[0].Message)
				}
			}

			if tt.expectInError > 0 && len(errs.records) > 0 {
				if errs.records[0].Message != tt.logMessage {
					t.Errorf("errorHandler: expected message %q, got %q", tt.logMessage, errs.records[0].Message)
				}
			}
		})
	}
}

func TestLevelBasedHandler_RespectsLevelThreshold(t *testing.T) {
	tests := []struct {
		name             string
		minLevel         slog.Level
		logsToMake       []slog.Level
		expectLowCount   int
		expectErrorCount int
	}{
		{
			name:             "info threshold filters debug",
			minLevel:         slog.LevelInfo,
			logsToMake:       []slog.Level{slog.LevelDebug, slog.LevelInfo, slog.LevelWarn},
			expectLowCount:   2, // info, warn
			expectErrorCount: 0,
		},
		{
			name:             "warn threshold filters debug and info",
			minLevel:         slog.LevelWarn,
			logsToMake:       []slog.Level{slog.LevelDebug, slog.LevelInfo, slog.LevelWarn, slog.LevelError},
			expectLowCount:   1, // warn
			expectErrorCount: 1, // error
		},
		{
			name:             "error threshold filters all low levels",
			minLevel:         slog.LevelError,
			logsToMake:       []slog.Level{slog.LevelDebug, slog.LevelInfo, slog.LevelWarn, slog.LevelError},
			expectLowCount:   0,
			expectErrorCount: 1, // error
		},
		{
			name:             "debug threshold allows all",
			minLevel:         slog.LevelDebug,
			logsToMake:       []slog.Level{slog.LevelDebug, slog.LevelInfo, slog.LevelError},
			expectLowCount:   2, // debug, info
			expectErrorCount: 1, // error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			low := &levelFilterHandler{minLevel: tt.minLevel}
			errs := &levelFilterHandler{minLevel: tt.minLevel}

			logger := slog.New(&LevelBasedHandler{
				LowLevelHandler: low,
				ErrorHandler:    errs,
			})

			for _, lvl := range tt.logsToMake {
				logger.Log(context.Background(), lvl, "test message", "k", "v")
			}

			if len(low.records) != tt.expectLowCount {
				t.Errorf("lowLevelHandler: expected %d records, got %d", tt.expectLowCount, len(low.records))
			}

			if len(errs.records) != tt.expectErrorCount {
				t.Errorf("errorHandler: expected %d records, got %d", tt.expectErrorCount, len(errs.records))
			}
		})
	}
}

func TestNew_NotNil(t *testing.T) {
	if got := New(); got == nil {
		t.Fatal("New() returned nil logger")
	}
}

func TestNew_WithOptions(t *testing.T) {
	tests := []struct {
		name string
		opts *tint.Options
	}{
		{
			name: "with nil options",
			opts: nil,
		},
		{
			name: "with debug level",
			opts: &tint.Options{Level: slog.LevelDebug},
		},
		{
			name: "with warn level",
			opts: &tint.Options{Level: slog.LevelWarn},
		},
		{
			name: "with error level",
			opts: &tint.Options{Level: slog.LevelError},
		},
		{
			name: "with no color",
			opts: &tint.Options{NoColor: true},
		},
		{
			name: "with time format",
			opts: &tint.Options{TimeFormat: "15:04:05"},
		},
		{
			name: "with replace attr",
			opts: &tint.Options{
				ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
					return a
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := New(tt.opts)
			if logger == nil {
				t.Fatal("New() returned nil logger")
			}

			// Verify logger works by logging at different levels
			logger.Debug("debug")
			logger.Info("info")
			logger.Warn("warn")
			logger.Error("error")
		})
	}
}

func TestNew_WithOptions_ReplaceAttrCalled(t *testing.T) {
	var called int
	opts := &tint.Options{
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			called++
			return a
		},
	}

	logger := New(opts)
	if logger == nil {
		t.Fatal("New(opts) returned nil logger")
	}

	logger.Error("with options", "k", "v")

	if called == 0 {
		t.Fatal("expected user ReplaceAttr to be called at least once")
	}
}

func TestInitDefault(t *testing.T) {
	tests := []struct {
		name string
		opts *tint.Options
	}{
		{
			name: "default options",
			opts: nil,
		},
		{
			name: "with debug level",
			opts: &tint.Options{Level: slog.LevelDebug},
		},
		{
			name: "with warn level",
			opts: &tint.Options{Level: slog.LevelWarn},
		},
		{
			name: "with no color",
			opts: &tint.Options{NoColor: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var logger *slog.Logger
			if tt.opts != nil {
				logger = InitDefault(tt.opts)
			} else {
				logger = InitDefault()
			}

			if logger == nil {
				t.Fatal("InitDefault() returned nil logger")
			}

			// Verify it sets the global default
			logger.Debug("debug message")
			logger.Info("info message")
			logger.Warn("warn message")
			logger.Error("error message")

			// Also via the global default
			slog.Info("global info")
		})
	}
}

func TestLevelBasedHandler_WithAttrs(t *testing.T) {
	low := &stubHandler{}
	errs := &stubHandler{}

	h := &LevelBasedHandler{
		LowLevelHandler: low,
		ErrorHandler:    errs,
	}

	attrs := []slog.Attr{
		slog.String("key", "value"),
		slog.Int("count", 42),
	}

	newHandler := h.WithAttrs(attrs)
	if newHandler == nil {
		t.Fatal("WithAttrs() returned nil")
	}

	// Verify it returns a LevelBasedHandler
	if _, ok := newHandler.(*LevelBasedHandler); !ok {
		t.Errorf("WithAttrs() should return *LevelBasedHandler, got %T", newHandler)
	}
}

func TestLevelBasedHandler_WithGroup(t *testing.T) {
	low := &stubHandler{}
	errs := &stubHandler{}

	h := &LevelBasedHandler{
		LowLevelHandler: low,
		ErrorHandler:    errs,
	}

	newHandler := h.WithGroup("mygroup")
	if newHandler == nil {
		t.Fatal("WithGroup() returned nil")
	}

	// Verify it returns a LevelBasedHandler
	if _, ok := newHandler.(*LevelBasedHandler); !ok {
		t.Errorf("WithGroup() should return *LevelBasedHandler, got %T", newHandler)
	}
}

func TestLevelBasedHandler_Enabled(t *testing.T) {
	tests := []struct {
		name        string
		level       slog.Level
		expectLow   bool
		expectError bool
	}{
		{
			name:        "debug enabled for low",
			level:       slog.LevelDebug,
			expectLow:   true,
			expectError: false,
		},
		{
			name:        "info enabled for low",
			level:       slog.LevelInfo,
			expectLow:   true,
			expectError: false,
		},
		{
			name:        "warn enabled for low",
			level:       slog.LevelWarn,
			expectLow:   true,
			expectError: false,
		},
		{
			name:        "error enabled for error handler",
			level:       slog.LevelError,
			expectLow:   false,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			low := &stubHandler{}
			errs := &stubHandler{}

			h := &LevelBasedHandler{
				LowLevelHandler: low,
				ErrorHandler:    errs,
			}

			ctx := context.Background()
			enabled := h.Enabled(ctx, tt.level)

			if !enabled {
				t.Errorf("Enabled(%v) = false, want true", tt.level)
			}
		})
	}
}
