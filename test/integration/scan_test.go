package integration_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

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
