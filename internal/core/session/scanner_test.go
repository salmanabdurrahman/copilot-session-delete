package session_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/salmanabdurrahman/copilot-session-delete/internal/core/session"
)

func TestScanner_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	scanner := session.NewScanner(dir)
	sessions, err := scanner.Scan(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sessions) != 0 {
		t.Errorf("expected 0 sessions, got %d", len(sessions))
	}
}

func TestScanner_RootNotExist(t *testing.T) {
	scanner := session.NewScanner("/nonexistent/path/abc")
	_, err := scanner.Scan(context.Background())
	if err == nil {
		t.Fatal("expected error for nonexistent root, got nil")
	}
}

func TestScanner_SkipsNonUUID(t *testing.T) {
	dir := t.TempDir()
	// Create valid UUID dirs and one non-UUID dir.
	validUUIDs := []string{
		"86334621-8152-4e67-b322-9f139d6c0a57",
		"c0c723f4-08d2-4257-9b30-5d2fd728dc45",
	}
	for _, id := range validUUIDs {
		if err := os.Mkdir(filepath.Join(dir, id), 0755); err != nil {
			t.Fatal(err)
		}
	}
	// Non-UUID entries that should be skipped.
	if err := os.Mkdir(filepath.Join(dir, "not-a-uuid"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "some-file.txt"), []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}

	scanner := session.NewScanner(dir)
	sessions, err := scanner.Scan(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sessions) != 2 {
		t.Errorf("expected 2 sessions, got %d", len(sessions))
	}
	for _, s := range sessions {
		if s.ID != validUUIDs[0] && s.ID != validUUIDs[1] {
			t.Errorf("unexpected session ID %q", s.ID)
		}
	}
}

func TestScanner_IntegrationWithFixtures(t *testing.T) {
	// Use the fixtures directory relative to module root.
	fixtureDir := filepath.Join("..", "..", "..", "test", "fixtures", "session-state-sample")
	if _, err := os.Stat(fixtureDir); os.IsNotExist(err) {
		t.Skip("fixtures directory not found:", fixtureDir)
	}

	scanner := session.NewScanner(fixtureDir)
	sessions, err := scanner.Scan(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sessions) < 2 {
		t.Errorf("expected at least 2 sessions from fixtures, got %d", len(sessions))
	}
}
