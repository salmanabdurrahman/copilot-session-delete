package session_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/salmanabdurrahman/copilot-session-delete/internal/core/session"
)

// TestScanner_EmptyDir verifies empty root dir returns an empty list without error.
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

// TestScanner_RootNotExist verifies root path that does not exist returns an error.
func TestScanner_RootNotExist(t *testing.T) {
	scanner := session.NewScanner("/nonexistent/path/abc")
	_, err := scanner.Scan(context.Background())
	if err == nil {
		t.Fatal("expected error for nonexistent root, got nil")
	}
}

// TestScanner_SkipsNonUUID verifies non-UUID entries (folders and files) in root dir are skipped.
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

// TestScanner_IntegrationWithFixtures verifies all fixture sessions in the sample dir are discovered.
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

// TestScanner_SkipsFileWithUUIDName verifies a regular file whose name is a valid UUID is skipped.
func TestScanner_SkipsFileWithUUIDName(t *testing.T) {
	dir := t.TempDir()
	validID := "86334621-8152-4e67-b322-9f139d6c0a57"

	// Create a valid session directory.
	if err := os.Mkdir(filepath.Join(dir, validID), 0755); err != nil {
		t.Fatal(err)
	}
	// Create a FILE whose name is also a valid UUID — must be skipped.
	fileUUID := "c0c723f4-08d2-4257-9b30-5d2fd728dc45"
	if err := os.WriteFile(filepath.Join(dir, fileUUID), []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}

	scanner := session.NewScanner(dir)
	sessions, err := scanner.Scan(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sessions) != 1 {
		t.Errorf("expected 1 session (file UUID skipped), got %d", len(sessions))
	}
	if sessions[0].ID != validID {
		t.Errorf("expected session ID %q, got %q", validID, sessions[0].ID)
	}
}

// TestScanner_RootIsFile verifies root path that is a regular file returns an error.
func TestScanner_RootIsFile(t *testing.T) {
	f, err := os.CreateTemp("", "not-a-dir-*.txt")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	defer os.Remove(f.Name())

	scanner := session.NewScanner(f.Name())
	_, err = scanner.Scan(context.Background())
	if err == nil {
		t.Fatal("expected error when root is a file, got nil")
	}
}
