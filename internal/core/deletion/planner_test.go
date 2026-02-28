package deletion_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/salmanabdurrahman/copilot-session-delete/internal/core/deletion"
	"github.com/salmanabdurrahman/copilot-session-delete/internal/core/session"
)

func TestPlanner_ValidSessions(t *testing.T) {
	root := t.TempDir()

	ids := []string{
		"86334621-8152-4e67-b322-9f139d6c0a57",
		"c0c723f4-08d2-4257-9b30-5d2fd728dc45",
	}
	var sessions []session.Session
	for _, id := range ids {
		dir := filepath.Join(root, id)
		if err := os.Mkdir(dir, 0755); err != nil {
			t.Fatal(err)
		}
		sessions = append(sessions, session.Session{ID: id, RootPath: dir})
	}

	planner := deletion.NewPlanner(root)
	plans, err := planner.Build(sessions)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(plans) != 2 {
		t.Errorf("expected 2 plans, got %d", len(plans))
	}
	for _, p := range plans {
		if p.NotFound {
			t.Errorf("session %s should be found", p.Session.ID)
		}
	}
}

func TestPlanner_NotFound(t *testing.T) {
	root := t.TempDir()
	id := "86334621-8152-4e67-b322-9f139d6c0a57"
	// Do NOT create the directory — simulate missing session.

	planner := deletion.NewPlanner(root)
	plans, err := planner.Build([]session.Session{{ID: id, RootPath: filepath.Join(root, id)}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(plans) != 1 {
		t.Fatalf("expected 1 plan, got %d", len(plans))
	}
	if !plans[0].NotFound {
		t.Error("expected NotFound=true for missing session directory")
	}
}

func TestPlanner_InvalidUUID(t *testing.T) {
	root := t.TempDir()
	planner := deletion.NewPlanner(root)
	_, err := planner.Build([]session.Session{{ID: "not-a-uuid", RootPath: filepath.Join(root, "not-a-uuid")}})
	if err == nil {
		t.Fatal("expected error for invalid UUID, got nil")
	}
}
