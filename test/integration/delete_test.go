package integration_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/salmanabdurrahman/copilot-session-delete/internal/core/deletion"
	"github.com/salmanabdurrahman/copilot-session-delete/internal/core/session"
)

// TestIntegration_DeleteSingleSession verifies that a single session directory is
// removed from disk after a successful planner + executor run.
func TestIntegration_DeleteSingleSession(t *testing.T) {
	root := t.TempDir()
	id := "86334621-8152-4e67-b322-9f139d6c0a57"
	dir := filepath.Join(root, id)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}

	s := session.Session{ID: id, RootPath: dir}
	plans, err := deletion.NewPlanner(root).Build([]session.Session{s})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	for r := range deletion.NewExecutor(false).Execute(context.Background(), plans) {
		if !r.Success {
			t.Fatalf("Execute: session %s failed: %v", r.SessionID, r.Err)
		}
	}

	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Errorf("expected session directory to be removed, stat err: %v", err)
	}
}

// TestIntegration_DeleteMultipleSessions verifies that all selected session directories
// are removed when multiple sessions are deleted at once.
func TestIntegration_DeleteMultipleSessions(t *testing.T) {
	root := t.TempDir()
	ids := []string{
		"86334621-8152-4e67-b322-9f139d6c0a57",
		"c0c723f4-08d2-4257-9b30-5d2fd728dc45",
	}

	var sessions []session.Session
	for _, id := range ids {
		dir := filepath.Join(root, id)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("setup: %v", err)
		}
		sessions = append(sessions, session.Session{ID: id, RootPath: dir})
	}

	plans, err := deletion.NewPlanner(root).Build(sessions)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	for r := range deletion.NewExecutor(false).Execute(context.Background(), plans) {
		if !r.Success {
			t.Fatalf("Execute: session %s failed: %v", r.SessionID, r.Err)
		}
	}

	for _, id := range ids {
		dir := filepath.Join(root, id)
		if _, err := os.Stat(dir); !os.IsNotExist(err) {
			t.Errorf("expected %s to be removed, stat err: %v", id, err)
		}
	}
}

// TestIntegration_DryRunDoesNotDeleteSessions verifies that dry-run mode reports
// success without removing any files from disk.
func TestIntegration_DryRunDoesNotDeleteSessions(t *testing.T) {
	root := t.TempDir()
	id := "86334621-8152-4e67-b322-9f139d6c0a57"
	dir := filepath.Join(root, id)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}

	s := session.Session{ID: id, RootPath: dir}
	plans, err := deletion.NewPlanner(root).Build([]session.Session{s})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	for r := range deletion.NewExecutor(true).Execute(context.Background(), plans) {
		if !r.Success || !r.DryRun {
			t.Fatalf("Execute dry-run: unexpected result %+v", r)
		}
	}

	// Directory must still exist after dry-run.
	if _, err := os.Stat(dir); err != nil {
		t.Errorf("expected session directory to remain after dry-run: %v", err)
	}
}

// TestIntegration_SafetyGuard_NeverDeletesOutsideRoot verifies the planner rejects
// any session whose path resolves outside the declared session-state root.
func TestIntegration_SafetyGuard_NeverDeletesOutsideRoot(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir() // separate temp dir simulating an out-of-bounds path

	id := "86334621-8152-4e67-b322-9f139d6c0a57"
	s := session.Session{ID: id, RootPath: outside}

	_, err := deletion.NewPlanner(root).Build([]session.Session{s})
	if err == nil {
		t.Fatal("expected planner to reject path outside root, got nil error")
	}
}
