package deletion_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/salmanabdurrahman/copilot-session-delete/internal/core/deletion"
	"github.com/salmanabdurrahman/copilot-session-delete/internal/core/session"
)

// makeSessionDir creates a session directory with a dummy file and returns the path.
func makeSessionDir(t *testing.T, root, id string) string {
	t.Helper()
	dir := filepath.Join(root, id)
	if err := os.Mkdir(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "events.jsonl"), []byte(`{"type":"session.start"}`+"\n"), 0644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestExecutor_DryRun(t *testing.T) {
	root := t.TempDir()
	id := "86334621-8152-4e67-b322-9f139d6c0a57"
	dir := makeSessionDir(t, root, id)

	plans := []deletion.Plan{
		{Session: session.Session{ID: id, RootPath: dir}},
	}

	exec := deletion.NewExecutor(true) // dry-run
	results := collect(exec.Execute(context.Background(), plans))

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	r := results[0]
	if !r.Success {
		t.Errorf("dry-run result should be Success=true")
	}
	if !r.DryRun {
		t.Errorf("dry-run result should have DryRun=true")
	}
	// Dir should still exist after dry-run.
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Error("dry-run should not remove directory")
	}
}

func TestExecutor_RealDelete(t *testing.T) {
	root := t.TempDir()
	id := "c0c723f4-08d2-4257-9b30-5d2fd728dc45"
	dir := makeSessionDir(t, root, id)

	plans := []deletion.Plan{
		{Session: session.Session{ID: id, RootPath: dir}},
	}

	exec := deletion.NewExecutor(false)
	results := collect(exec.Execute(context.Background(), plans))

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if !results[0].Success {
		t.Errorf("expected Success=true, got error: %v", results[0].Err)
	}
	// Dir should be gone.
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Error("directory should have been removed")
	}
}

func TestExecutor_NotFound_IsIdempotent(t *testing.T) {
	plans := []deletion.Plan{
		{
			Session:  session.Session{ID: "86334621-8152-4e67-b322-9f139d6c0a57", RootPath: "/nonexistent/path"},
			NotFound: true,
		},
	}

	exec := deletion.NewExecutor(false)
	results := collect(exec.Execute(context.Background(), plans))

	if len(results) != 1 {
		t.Fatalf("expected 1 result")
	}
	if !results[0].Success {
		t.Error("NotFound plan should be treated as success")
	}
	if !results[0].NotFound {
		t.Error("NotFound flag should be set in result")
	}
}

// collect drains a Result channel into a slice.
func collect(ch <-chan deletion.Result) []deletion.Result {
	var out []deletion.Result
	for r := range ch {
		out = append(out, r)
	}
	return out
}
