package slogx

import (
	"context"
	"log/slog"
	"testing"
	"time"
)

type stubHandler struct {
	records []slog.Record
}

func (s *stubHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return true
}

func (s *stubHandler) Handle(ctx context.Context, r slog.Record) error {
	// Clone to decouple from caller.
	s.records = append(s.records, r.Clone())
	return nil
}

func (s *stubHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return s
}

func (s *stubHandler) WithGroup(name string) slog.Handler {
	return s
}

func TestLevelBasedHandler_RoutesByLevel(t *testing.T) {
	low := &stubHandler{}
	errs := &stubHandler{}

	h := &LevelBasedHandler{
		LowLevelHandler: low,
		ErrorHandler:    errs,
	}

	ctx := context.Background()

	infoRec := slog.NewRecord(time.Now(), slog.LevelInfo, "info", 0)
	if err := h.Handle(ctx, infoRec); err != nil {
		t.Fatalf("Handle(info) error = %v", err)
	}

	if len(low.records) != 1 {
		t.Fatalf("expected lowLevelHandler to receive 1 record, got %d", len(low.records))
	}
	if len(errs.records) != 0 {
		t.Fatalf("expected errorHandler to receive 0 records for info, got %d", len(errs.records))
	}

	errorRec := slog.NewRecord(time.Now(), slog.LevelError, "error", 0)
	if err := h.Handle(ctx, errorRec); err != nil {
		t.Fatalf("Handle(error) error = %v", err)
	}

	if len(low.records) != 1 {
		t.Fatalf("expected lowLevelHandler to still have 1 record, got %d", len(low.records))
	}
	if len(errs.records) != 1 {
		t.Fatalf("expected errorHandler to receive 1 record for error, got %d", len(errs.records))
	}
}

func TestNew_NotNil(t *testing.T) {
	if got := New(); got == nil {
		t.Fatal("New() returned nil logger")
	}
}

func TestInitDefault_SetsGlobal(t *testing.T) {
	logger := InitDefault()
	if logger == nil {
		t.Fatal("InitDefault() returned nil logger")
	}

	// These calls should go through the level-based handler and
	// produce visible output when you run:
	//   go test -run TestInitDefault_SetsGlobal -v
	logger.Debug("debug message", "k", "v")
	logger.Info("info message", "k", "v")
	logger.Warn("warn message", "k", "v")
	logger.Error("error message", "k", "v")

	// Also via the global default.
	slog.Info("global info", "k", "v")
}
