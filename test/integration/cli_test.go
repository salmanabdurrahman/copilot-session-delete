package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/salmanabdurrahman/copilot-session-delete/internal/app/cli"
)

// ─── list command ─────────────────────────────────────────────────────────────

// TestIntegration_ListCommand_Table verifies the list command produces a
// human-readable table using a real temp directory.
func TestIntegration_ListCommand_Table(t *testing.T) {
	root := t.TempDir()
	makeCliSession(t, root, "86334621-8152-4e67-b322-9f139d6c0a57", "github/copilot-cli", "2026-02-28T09:47:00Z")
	makeCliSession(t, root, "c0c723f4-08d2-4257-9b30-5d2fd728dc45", "my-project", "2026-02-28T10:18:00Z")

	var buf bytes.Buffer
	if err := cli.RunList(context.Background(), cli.ListOptions{
		SessionDir: root,
		JSON:       false,
		Out:        &buf,
	}); err != nil {
		t.Fatalf("RunList: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "SESSION ID") {
		t.Error("expected header in table output")
	}
	if !strings.Contains(out, "86334621") {
		t.Error("expected first session ID in table output")
	}
	if !strings.Contains(out, "github/copilot-cli") {
		t.Error("expected repository label in table output")
	}
}

// TestIntegration_ListCommand_JSON verifies the list command produces valid JSON
// with the correct number of sessions and required fields.
func TestIntegration_ListCommand_JSON(t *testing.T) {
	root := t.TempDir()
	makeCliSession(t, root, "86334621-8152-4e67-b322-9f139d6c0a57", "github/copilot-cli", "2026-02-28T09:47:00Z")
	makeCliSession(t, root, "c0c723f4-08d2-4257-9b30-5d2fd728dc45", "my-project", "2026-02-28T10:18:00Z")

	var buf bytes.Buffer
	if err := cli.RunList(context.Background(), cli.ListOptions{
		SessionDir: root,
		JSON:       true,
		Out:        &buf,
	}); err != nil {
		t.Fatalf("RunList JSON: %v", err)
	}

	var decoded []map[string]any
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, buf.String())
	}
	if len(decoded) != 2 {
		t.Fatalf("expected 2 sessions in JSON, got %d", len(decoded))
	}

	// Both sessions must have stable required fields.
	for i, obj := range decoded {
		for _, key := range []string{"id", "event_count", "size_bytes"} {
			if _, ok := obj[key]; !ok {
				t.Errorf("session[%d]: expected key %q in JSON output", i, key)
			}
		}
	}
}

// TestIntegration_ListCommand_Empty verifies the list command outputs an empty
// JSON array and a valid table header when the directory has no sessions.
func TestIntegration_ListCommand_Empty(t *testing.T) {
	root := t.TempDir()

	// Table.
	var buf bytes.Buffer
	if err := cli.RunList(context.Background(), cli.ListOptions{SessionDir: root, Out: &buf}); err != nil {
		t.Fatalf("RunList empty table: %v", err)
	}
	if !strings.Contains(buf.String(), "SESSION ID") {
		t.Error("expected header even for empty session list")
	}

	// JSON.
	buf.Reset()
	if err := cli.RunList(context.Background(), cli.ListOptions{
		SessionDir: root,
		JSON:       true,
		Out:        &buf,
	}); err != nil {
		t.Fatalf("RunList empty JSON: %v", err)
	}
	var arr []any
	if err := json.Unmarshal(buf.Bytes(), &arr); err != nil {
		t.Fatalf("expected valid JSON array for empty result: %v", err)
	}
	if len(arr) != 0 {
		t.Errorf("expected empty JSON array, got %d elements", len(arr))
	}
}

// ─── delete command ───────────────────────────────────────────────────────────

// TestIntegration_DeleteCommand_DryRun verifies the delete command prints a
// dry-run preview and does not remove any files.
func TestIntegration_DeleteCommand_DryRun(t *testing.T) {
	root := t.TempDir()
	id := "86334621-8152-4e67-b322-9f139d6c0a57"
	makeCliSession(t, root, id, "github/copilot-cli", "2026-02-28T09:47:00Z")

	var out bytes.Buffer
	if err := cli.RunDelete(context.Background(), cli.DeleteOptions{
		SessionDir: root,
		IDs:        []string{id},
		DryRun:     true,
		Out:        &out,
	}); err != nil {
		t.Fatalf("RunDelete dry-run: %v", err)
	}

	if !strings.Contains(out.String(), "[DRY-RUN]") {
		t.Errorf("expected [DRY-RUN] in preview output, got:\n%s", out.String())
	}
	if _, err := os.Stat(filepath.Join(root, id)); err != nil {
		t.Errorf("session must remain after dry-run: %v", err)
	}
}

// TestIntegration_DeleteCommand_SingleID verifies the delete command removes
// the target session directory from disk.
func TestIntegration_DeleteCommand_SingleID(t *testing.T) {
	root := t.TempDir()
	id := "86334621-8152-4e67-b322-9f139d6c0a57"
	makeCliSession(t, root, id, "github/copilot-cli", "2026-02-28T09:47:00Z")

	var out bytes.Buffer
	if err := cli.RunDelete(context.Background(), cli.DeleteOptions{
		SessionDir: root,
		IDs:        []string{id},
		DryRun:     false,
		Out:        &out,
	}); err != nil {
		t.Fatalf("RunDelete single: %v", err)
	}

	if _, err := os.Stat(filepath.Join(root, id)); !os.IsNotExist(err) {
		t.Errorf("expected session directory to be removed, stat err: %v", err)
	}
}

// TestIntegration_DeleteCommand_AllYes verifies --all --yes removes every session.
func TestIntegration_DeleteCommand_AllYes(t *testing.T) {
	root := t.TempDir()
	ids := []string{
		"86334621-8152-4e67-b322-9f139d6c0a57",
		"c0c723f4-08d2-4257-9b30-5d2fd728dc45",
	}
	for _, id := range ids {
		makeCliSession(t, root, id, "repo", "2026-02-28T10:00:00Z")
	}

	if err := cli.RunDelete(context.Background(), cli.DeleteOptions{
		SessionDir: root,
		All:        true,
		DryRun:     false,
		Yes:        true,
		Out:        &bytes.Buffer{},
	}); err != nil {
		t.Fatalf("RunDelete --all --yes: %v", err)
	}

	for _, id := range ids {
		if _, err := os.Stat(filepath.Join(root, id)); !os.IsNotExist(err) {
			t.Errorf("expected session %s to be removed, stat err: %v", id, err)
		}
	}
}

// TestIntegration_DeleteCommand_AllWithoutYes verifies --all without --yes returns
// an error and does not modify the disk.
func TestIntegration_DeleteCommand_AllWithoutYes(t *testing.T) {
	root := t.TempDir()
	id := "86334621-8152-4e67-b322-9f139d6c0a57"
	makeCliSession(t, root, id, "repo", "2026-02-28T10:00:00Z")

	err := cli.RunDelete(context.Background(), cli.DeleteOptions{
		SessionDir: root,
		All:        true,
		DryRun:     false,
		Yes:        false,
		Out:        &bytes.Buffer{},
	})
	if err == nil {
		t.Fatal("expected error for --all without --yes, got nil")
	}

	// Session must remain untouched.
	if _, statErr := os.Stat(filepath.Join(root, id)); statErr != nil {
		t.Errorf("expected session to remain after rejected --all: %v", statErr)
	}
}

// ─── helpers ─────────────────────────────────────────────────────────────────

// makeCliSession creates a session directory with a workspace.yaml that includes
// repository and updatedAt fields.
func makeCliSession(t *testing.T, root, id, repo, updatedAt string) {
	t.Helper()
	dir := filepath.Join(root, id)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("makeCliSession mkdir: %v", err)
	}
	yaml := "id: " + id + "\n" +
		"repository: " + repo + "\n" +
		"updated_at: \"" + updatedAt + "\"\n"
	if err := os.WriteFile(filepath.Join(dir, "workspace.yaml"), []byte(yaml), 0o644); err != nil {
		t.Fatalf("makeCliSession write: %v", err)
	}
}
