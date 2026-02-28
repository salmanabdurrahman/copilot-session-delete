package session_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/salmanabdurrahman/copilot-session-delete/internal/core/session"
)

// TestEnrichMetadata_ValidWorkspace verifies workspace.yaml with all fields present is parsed correctly.
func TestEnrichMetadata_ValidWorkspace(t *testing.T) {
	dir := t.TempDir()
	id := "86334621-8152-4e67-b322-9f139d6c0a57"
	sessionDir := filepath.Join(dir, id)
	if err := os.Mkdir(sessionDir, 0755); err != nil {
		t.Fatal(err)
	}

	yaml := `id: 86334621-8152-4e67-b322-9f139d6c0a57
cwd: /home/user/projects/foo
repository: github/copilot-cli
branch: main
summary: Test session
created_at: "2026-02-28T09:00:00Z"
updated_at: "2026-02-28T09:30:00Z"
`
	if err := os.WriteFile(filepath.Join(sessionDir, "workspace.yaml"), []byte(yaml), 0644); err != nil {
		t.Fatal(err)
	}
	// Write 3 lines of events.
	events := `{"type":"session.start"}
{"type":"user.message"}
{"type":"assistant.message"}
`
	if err := os.WriteFile(filepath.Join(sessionDir, "events.jsonl"), []byte(events), 0644); err != nil {
		t.Fatal(err)
	}

	s := session.Session{ID: id, RootPath: sessionDir}
	session.EnrichMetadata(&s)

	if s.MetadataErr != nil {
		t.Errorf("expected no metadata error, got: %v", s.MetadataErr)
	}
	if s.Repository != "github/copilot-cli" {
		t.Errorf("unexpected repository %q", s.Repository)
	}
	if s.Branch != "main" {
		t.Errorf("unexpected branch %q", s.Branch)
	}
	if s.EventCount != 3 {
		t.Errorf("expected 3 events, got %d", s.EventCount)
	}
	want := time.Date(2026, 2, 28, 9, 30, 0, 0, time.UTC)
	if !s.UpdatedAt.Equal(want) {
		t.Errorf("unexpected updated_at: got %v, want %v", s.UpdatedAt, want)
	}
}

// TestEnrichMetadata_MissingWorkspace verifies missing workspace.yaml sets MetadataErr and falls back to dir mtime.
func TestEnrichMetadata_MissingWorkspace(t *testing.T) {
	dir := t.TempDir()
	id := "c0c723f4-08d2-4257-9b30-5d2fd728dc45"
	sessionDir := filepath.Join(dir, id)
	if err := os.Mkdir(sessionDir, 0755); err != nil {
		t.Fatal(err)
	}
	// No workspace.yaml, no events.jsonl.

	s := session.Session{ID: id, RootPath: sessionDir}
	session.EnrichMetadata(&s)

	if s.MetadataErr == nil {
		t.Error("expected non-nil MetadataErr for missing workspace.yaml")
	}
	// UpdatedAt should fall back to directory mtime (non-zero).
	if s.UpdatedAt.IsZero() {
		t.Error("expected non-zero UpdatedAt fallback from directory mtime")
	}
	// EventCount should be -1 when events.jsonl is absent.
	if s.EventCount != -1 {
		t.Errorf("expected EventCount -1, got %d", s.EventCount)
	}
}

// TestEnrichMetadata_CorruptWorkspace verifies corrupt workspace.yaml sets MetadataErr and does not panic.
func TestEnrichMetadata_CorruptWorkspace(t *testing.T) {
	dir := t.TempDir()
	id := "86334621-8152-4e67-b322-9f139d6c0a57"
	sessionDir := filepath.Join(dir, id)
	if err := os.Mkdir(sessionDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sessionDir, "workspace.yaml"), []byte("{invalid yaml: [unclosed bracket"), 0644); err != nil {
		t.Fatal(err)
	}

	s := session.Session{ID: id, RootPath: sessionDir}
	session.EnrichMetadata(&s)

	if s.MetadataErr == nil {
		t.Error("expected MetadataErr for corrupt workspace.yaml")
	}
}

// TestEnrichMetadata_OptionalFieldsMissing verifies workspace.yaml with optional fields absent parses without error.
func TestEnrichMetadata_OptionalFieldsMissing(t *testing.T) {
	dir := t.TempDir()
	id := "86334621-8152-4e67-b322-9f139d6c0a57"
	sessionDir := filepath.Join(dir, id)
	if err := os.Mkdir(sessionDir, 0755); err != nil {
		t.Fatal(err)
	}
	// Only required id field; optional fields intentionally absent.
	yaml := "id: 86334621-8152-4e67-b322-9f139d6c0a57\nupdated_at: \"2026-02-28T10:00:00Z\"\n"
	if err := os.WriteFile(filepath.Join(sessionDir, "workspace.yaml"), []byte(yaml), 0644); err != nil {
		t.Fatal(err)
	}

	s := session.Session{ID: id, RootPath: sessionDir}
	session.EnrichMetadata(&s)

	if s.MetadataErr != nil {
		t.Errorf("expected no error for minimal workspace.yaml, got: %v", s.MetadataErr)
	}
	if s.Repository != "" {
		t.Errorf("expected empty repository, got %q", s.Repository)
	}
	if s.Branch != "" {
		t.Errorf("expected empty branch, got %q", s.Branch)
	}
}

// TestEnrichMetadata_PartialCorruptEvents verifies events.jsonl with mixed valid and invalid JSON lines counts only valid lines.
func TestEnrichMetadata_PartialCorruptEvents(t *testing.T) {
	dir := t.TempDir()
	id := "c0c723f4-08d2-4257-9b30-5d2fd728dc45"
	sessionDir := filepath.Join(dir, id)
	if err := os.Mkdir(sessionDir, 0755); err != nil {
		t.Fatal(err)
	}

	// 3 valid JSON lines + 2 invalid lines.
	events := `{"type":"session.start"}
{"type":"user.message"}
not valid json at all
{"type":"assistant.message"}
{broken
`
	if err := os.WriteFile(filepath.Join(sessionDir, "events.jsonl"), []byte(events), 0644); err != nil {
		t.Fatal(err)
	}

	s := session.Session{ID: id, RootPath: sessionDir}
	session.EnrichMetadata(&s)

	if s.EventCount != 3 {
		t.Errorf("expected 3 valid events (skipping 2 corrupt), got %d", s.EventCount)
	}
}

// TestEnrichMetadata_ZeroUpdatedAtFallback verifies workspace.yaml with missing updated_at falls back to dir mtime.
func TestEnrichMetadata_ZeroUpdatedAtFallback(t *testing.T) {
	dir := t.TempDir()
	id := "86334621-8152-4e67-b322-9f139d6c0a57"
	sessionDir := filepath.Join(dir, id)
	if err := os.Mkdir(sessionDir, 0755); err != nil {
		t.Fatal(err)
	}
	// workspace.yaml without updated_at field.
	yaml := "id: 86334621-8152-4e67-b322-9f139d6c0a57\ncwd: /some/path\n"
	if err := os.WriteFile(filepath.Join(sessionDir, "workspace.yaml"), []byte(yaml), 0644); err != nil {
		t.Fatal(err)
	}

	s := session.Session{ID: id, RootPath: sessionDir}
	session.EnrichMetadata(&s)

	if s.MetadataErr != nil {
		t.Errorf("unexpected metadata error: %v", s.MetadataErr)
	}
	if s.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should fall back to dir mtime when field is absent, but got zero")
	}
}

func TestSession_Label(t *testing.T) {
	tests := []struct {
		name string
		s    session.Session
		want string
	}{
		{
			name: "repository takes priority",
			s:    session.Session{ID: "86334621-8152-4e67-b322-9f139d6c0a57", Repository: "github/copilot-cli", CWD: "/home/user/foo"},
			want: "github/copilot-cli",
		},
		{
			name: "cwd basename when no repo",
			s:    session.Session{ID: "86334621-8152-4e67-b322-9f139d6c0a57", CWD: "/home/user/my-project"},
			want: "my-project",
		},
		{
			name: "id prefix as last resort",
			s:    session.Session{ID: "86334621-8152-4e67-b322-9f139d6c0a57"},
			want: "86334621…",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.s.Label()
			if got != tc.want {
				t.Errorf("Label() = %q, want %q", got, tc.want)
			}
		})
	}
}
