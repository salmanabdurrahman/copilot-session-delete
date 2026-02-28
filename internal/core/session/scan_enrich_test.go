package session_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/salmanabdurrahman/copilot-session-delete/internal/core/session"
)

// TestScanAndEnrich_Basic verifies that ScanAndEnrich returns enriched sessions
// sorted newest-first from a real temp directory.
func TestScanAndEnrich_Basic(t *testing.T) {
	root := t.TempDir()
	writeWorkspace(t, root, "86334621-8152-4e67-b322-9f139d6c0a57",
		"repository: github/copilot-cli\nupdated_at: \"2026-02-28T09:00:00Z\"\n")
	writeWorkspace(t, root, "c0c723f4-08d2-4257-9b30-5d2fd728dc45",
		"repository: my-project\nupdated_at: \"2026-02-28T10:00:00Z\"\n")

	sessions, err := session.ScanAndEnrich(context.Background(), root)
	if err != nil {
		t.Fatalf("ScanAndEnrich: %v", err)
	}
	if len(sessions) != 2 {
		t.Fatalf("expected 2 sessions, got %d", len(sessions))
	}
	// Newest first.
	if sessions[0].ID != "c0c723f4-08d2-4257-9b30-5d2fd728dc45" {
		t.Errorf("expected newest session first, got %s", sessions[0].ID)
	}
	// Metadata enriched.
	if sessions[0].Repository != "my-project" {
		t.Errorf("expected repository 'my-project', got %q", sessions[0].Repository)
	}
}

// TestScanAndEnrich_EmptyRoot verifies an empty root directory returns an empty
// slice without error.
func TestScanAndEnrich_EmptyRoot(t *testing.T) {
	sessions, err := session.ScanAndEnrich(context.Background(), t.TempDir())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sessions) != 0 {
		t.Errorf("expected 0 sessions, got %d", len(sessions))
	}
}

// TestScanAndEnrich_ScanError verifies a non-existent root returns an error.
func TestScanAndEnrich_ScanError(t *testing.T) {
	_, err := session.ScanAndEnrich(context.Background(), "/does/not/exist/scan-enrich")
	if err == nil {
		t.Error("expected error for non-existent root, got nil")
	}
}

// TestScanAndEnrich_ContextCancellation verifies that cancelling the context
// stops enrichment early and returns a context error.
func TestScanAndEnrich_ContextCancellation(t *testing.T) {
	root := t.TempDir()
	// Create several sessions so cancellation has a chance to fire.
	ids := []string{
		"86334621-8152-4e67-b322-9f139d6c0a57",
		"c0c723f4-08d2-4257-9b30-5d2fd728dc45",
		"d1e2f3a4-0000-4000-a000-000000000001",
		"d1e2f3a4-0000-4000-a000-000000000002",
	}
	for _, id := range ids {
		writeWorkspace(t, root, id, "updated_at: \"2026-02-28T10:00:00Z\"\n")
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately before scanning

	_, err := session.ScanAndEnrich(ctx, root)
	if err == nil {
		t.Error("expected context error from ScanAndEnrich when context is cancelled")
	}
}

// ─── helpers ─────────────────────────────────────────────────────────────────

// writeWorkspace creates a session directory with the given workspace.yaml content.
func writeWorkspace(t *testing.T, root, id, content string) {
	t.Helper()
	dir := filepath.Join(root, id)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("writeWorkspace mkdir: %v", err)
	}
	full := "id: " + id + "\n" + content
	if err := os.WriteFile(filepath.Join(dir, "workspace.yaml"), []byte(full), 0o644); err != nil {
		t.Fatalf("writeWorkspace write: %v", err)
	}
}
