package output_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/salmanabdurrahman/copilot-session-delete/internal/adapters/output"
	"github.com/salmanabdurrahman/copilot-session-delete/internal/core/session"
)

// ── FormatSize ────────────────────────────────────────────────────────────────

func TestFormatSize(t *testing.T) {
	tests := []struct {
		bytes int64
		want  string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1023, "1023 B"},
		{1024, "~1.0 KB"},
		{1536, "~1.5 KB"},
		{1024 * 1024, "~1.0 MB"},
		{2_202_009, "~2.1 MB"}, // ≈ 2.1 MB
		{1024 * 1024 * 1024, "~1.0 GB"},
	}
	for _, tc := range tests {
		got := output.FormatSize(tc.bytes)
		if got != tc.want {
			t.Errorf("FormatSize(%d) = %q, want %q", tc.bytes, got, tc.want)
		}
	}
}

// ── PrintSessions (Table) ─────────────────────────────────────────────────────

func TestPrintSessions_TableFormat(t *testing.T) {
	sessions := makeSessions()
	var buf bytes.Buffer
	if err := output.PrintSessions(&buf, sessions, output.FormatTable); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()

	// Header must be present.
	if !strings.Contains(out, "SESSION ID") {
		t.Error("expected SESSION ID header in table output")
	}
	if !strings.Contains(out, "UPDATED AT") {
		t.Error("expected UPDATED AT header in table output")
	}
	// Session IDs must appear.
	if !strings.Contains(out, "86334621") {
		t.Error("expected first session ID in table output")
	}
	if !strings.Contains(out, "c0c723f4") {
		t.Error("expected second session ID in table output")
	}
	// Label (repo) must appear.
	if !strings.Contains(out, "github/copilot-cli") {
		t.Error("expected repository label in table output")
	}
}

func TestPrintSessions_TableEmpty(t *testing.T) {
	var buf bytes.Buffer
	if err := output.PrintSessions(&buf, nil, output.FormatTable); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Must still print the header, not panic.
	if !strings.Contains(buf.String(), "SESSION ID") {
		t.Error("expected header even for empty session list")
	}
}

// ── PrintSessions (JSON) ──────────────────────────────────────────────────────

func TestPrintSessions_JSONFormat(t *testing.T) {
	sessions := makeSessions()
	var buf bytes.Buffer
	if err := output.PrintSessions(&buf, sessions, output.FormatJSON); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var decoded []map[string]any
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, buf.String())
	}
	if len(decoded) != 2 {
		t.Errorf("expected 2 JSON objects, got %d", len(decoded))
	}
	// Verify required fields.
	first := decoded[0]
	if first["id"] == nil {
		t.Error("expected 'id' field in JSON output")
	}
	if first["event_count"] == nil {
		t.Error("expected 'event_count' field in JSON output")
	}
}

func TestPrintSessions_JSONEmpty(t *testing.T) {
	var buf bytes.Buffer
	if err := output.PrintSessions(&buf, nil, output.FormatJSON); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var decoded []any
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("empty list should still produce valid JSON array: %v", err)
	}
	if len(decoded) != 0 {
		t.Errorf("expected empty JSON array, got %d elements", len(decoded))
	}
}

// TestPrintSessions_TableFormat_ZeroTime verifies sessions with zero UpdatedAt
// display "—" in the UPDATED AT column.
func TestPrintSessions_TableFormat_ZeroTime(t *testing.T) {
	sessions := []session.Session{
		{
			ID:         "86334621-8152-4e67-b322-9f139d6c0a57",
			EventCount: 5,
			// UpdatedAt intentionally zero.
		},
	}
	var buf bytes.Buffer
	if err := output.PrintSessions(&buf, sessions, output.FormatTable); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "—") {
		t.Error("expected '—' for zero UpdatedAt in table output")
	}
}

// TestPrintSessions_TableFormat_NegativeEventCount verifies sessions with
// EventCount < 0 display "?" in the EVENTS column.
func TestPrintSessions_TableFormat_NegativeEventCount(t *testing.T) {
	sessions := []session.Session{
		{
			ID:         "86334621-8152-4e67-b322-9f139d6c0a57",
			EventCount: -1, // unknown
		},
	}
	var buf bytes.Buffer
	if err := output.PrintSessions(&buf, sessions, output.FormatTable); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "?") {
		t.Error("expected '?' for negative EventCount in table output")
	}
}

// TestPrintSessions_TableFormat_LongID verifies a session ID longer than the
// column width is truncated in the table output.
func TestPrintSessions_TableFormat_LongID(t *testing.T) {
	sessions := []session.Session{
		{
			ID:         "86334621-8152-4e67-b322-9f139d6c0a57",
			EventCount: 0,
		},
	}
	var buf bytes.Buffer
	if err := output.PrintSessions(&buf, sessions, output.FormatTable); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// The full UUID (36 chars) should be present at full width since it fits.
	if !strings.Contains(buf.String(), "86334621-8152-4e67-b322-9f139d6c0a57") {
		t.Error("expected full UUID in table output when it fits within column width")
	}
}

func makeSessions() []session.Session {
	return []session.Session{
		{
			ID:         "86334621-8152-4e67-b322-9f139d6c0a57",
			RootPath:   "/fake/session-state/86334621-8152-4e67-b322-9f139d6c0a57",
			UpdatedAt:  time.Date(2026, 2, 28, 9, 47, 0, 0, time.UTC),
			CWD:        "/home/user/copilot-cli",
			Repository: "github/copilot-cli",
			Branch:     "main",
			EventCount: 150,
			SizeBytes:  2_100_000,
		},
		{
			ID:         "c0c723f4-08d2-4257-9b30-5d2fd728dc45",
			RootPath:   "/fake/session-state/c0c723f4-08d2-4257-9b30-5d2fd728dc45",
			UpdatedAt:  time.Date(2026, 2, 28, 10, 18, 0, 0, time.UTC),
			CWD:        "/home/user/copilot-session-delete",
			EventCount: 42,
			SizeBytes:  512_000,
		},
	}
}
