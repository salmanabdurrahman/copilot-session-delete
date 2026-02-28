package integration_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/salmanabdurrahman/copilot-session-delete/internal/core/session"
)

// TestScanFixtureDir verifies scanner + metadata enrichment against the
// bundled fixture sessions.
func TestScanFixtureDir(t *testing.T) {
	// Locate the fixture directory from the repo root.
	fixtureDir := filepath.Join("..", "fixtures", "session-state-sample")
	if _, err := os.Stat(fixtureDir); os.IsNotExist(err) {
		t.Skip("fixture dir not found:", fixtureDir)
	}

	ctx := context.Background()
	scanner := session.NewScanner(fixtureDir)
	sessions, err := scanner.Scan(ctx)
	if err != nil {
		t.Fatalf("scan failed: %v", err)
	}
	if len(sessions) < 2 {
		t.Fatalf("expected >=2 fixture sessions, got %d", len(sessions))
	}

	for i := range sessions {
		session.EnrichMetadata(&sessions[i])
		s := sessions[i]
		t.Logf("session %s: cwd=%q events=%d err=%v", s.ID, s.CWD, s.EventCount, s.MetadataErr)

		if s.ID == "" {
			t.Errorf("session[%d] has empty ID", i)
		}
		if s.UpdatedAt.IsZero() {
			t.Errorf("session[%d] has zero UpdatedAt", i)
		}
	}
}

// TestScanAndEnrich_SortedNewestFirst verifies the convenience function
// returns sessions sorted by UpdatedAt descending.
func TestScanAndEnrich_SortedNewestFirst(t *testing.T) {
	root := t.TempDir()

	// Create two sessions with distinct UpdatedAt values.
	older := makeFixtureSession(t, root,
		"86334621-8152-4e67-b322-9f139d6c0a57",
		"2026-02-28T09:00:00Z",
	)
	newer := makeFixtureSession(t, root,
		"c0c723f4-08d2-4257-9b30-5d2fd728dc45",
		"2026-02-28T10:00:00Z",
	)

	sessions, err := session.ScanAndEnrich(context.Background(), root)
	if err != nil {
		t.Fatalf("ScanAndEnrich failed: %v", err)
	}
	if len(sessions) != 2 {
		t.Fatalf("expected 2 sessions, got %d", len(sessions))
	}
	if sessions[0].ID != newer {
		t.Errorf("expected newest session first, got %q", sessions[0].ID)
	}
	if sessions[1].ID != older {
		t.Errorf("expected older session second, got %q", sessions[1].ID)
	}
}

// TestScanAndEnrich_EmptyDir returns empty slice without error.
func TestScanAndEnrich_EmptyDir(t *testing.T) {
	root := t.TempDir()
	sessions, err := session.ScanAndEnrich(context.Background(), root)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sessions) != 0 {
		t.Errorf("expected 0 sessions, got %d", len(sessions))
	}
}

// makeFixtureSession creates a minimal session directory with workspace.yaml
// and returns the session ID.
func makeFixtureSession(t *testing.T, root, id, updatedAt string) string {
	t.Helper()
	dir := filepath.Join(root, id)
	if err := os.Mkdir(dir, 0755); err != nil {
		t.Fatal(err)
	}
	content := "id: " + id + "\nupdated_at: \"" + updatedAt + "\"\n"
	if err := os.WriteFile(filepath.Join(dir, "workspace.yaml"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return id
}

// parseTime is a test helper to parse RFC3339 strings.
func parseTime(t *testing.T, s string) time.Time {
	t.Helper()
	ts, err := time.Parse(time.RFC3339, s)
	if err != nil {
		t.Fatalf("parseTime(%q): %v", s, err)
	}
	return ts
}
