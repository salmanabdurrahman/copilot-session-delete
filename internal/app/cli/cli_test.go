package cli_test

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

const (
	fixID1 = "86334621-8152-4e67-b322-9f139d6c0a57"
	fixID2 = "c0c723f4-08d2-4257-9b30-5d2fd728dc45"
)

// makeSession creates a minimal session directory with workspace.yaml in root.
func makeSession(t *testing.T, root, id, updatedAt string) {
	t.Helper()
	dir := filepath.Join(root, id)
	if err := os.Mkdir(dir, 0o755); err != nil {
		t.Fatalf("makeSession: mkdir: %v", err)
	}
	yaml := "id: " + id + "\nupdated_at: \"" + updatedAt + "\"\n"
	if err := os.WriteFile(filepath.Join(dir, "workspace.yaml"), []byte(yaml), 0o644); err != nil {
		t.Fatalf("makeSession: write workspace.yaml: %v", err)
	}
}

// ─── RunList ──────────────────────────────────────────────────────────────────

// TestRunList_TableFormat verifies the list command produces a table with session IDs.
func TestRunList_TableFormat(t *testing.T) {
	root := t.TempDir()
	makeSession(t, root, fixID1, "2026-02-28T09:47:00Z")
	makeSession(t, root, fixID2, "2026-02-28T10:18:00Z")

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
		t.Error("expected SESSION ID header in table output")
	}
	if !strings.Contains(out, fixID1[:8]) {
		t.Error("expected first session ID in table output")
	}
	if !strings.Contains(out, fixID2[:8]) {
		t.Error("expected second session ID in table output")
	}
}

// TestRunList_JSONFormat verifies the list command produces valid JSON with required fields.
func TestRunList_JSONFormat(t *testing.T) {
	root := t.TempDir()
	makeSession(t, root, fixID1, "2026-02-28T09:47:00Z")

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
	if len(decoded) != 1 {
		t.Fatalf("expected 1 session in JSON, got %d", len(decoded))
	}

	obj := decoded[0]
	for _, key := range []string{"id", "event_count", "size_bytes"} {
		if _, ok := obj[key]; !ok {
			t.Errorf("expected key %q in JSON output", key)
		}
	}
	if obj["id"] != fixID1 {
		t.Errorf("expected id %q, got %v", fixID1, obj["id"])
	}
}

// TestRunList_JSONFormat_StableSchema verifies all expected top-level keys are present
// and no unexpected keys appear in the JSON output (schema stability / golden check).
func TestRunList_JSONFormat_StableSchema(t *testing.T) {
	root := t.TempDir()
	// Use a rich workspace.yaml to populate optional fields.
	dir := filepath.Join(root, fixID1)
	if err := os.Mkdir(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	richYAML := "id: " + fixID1 + "\n" +
		"cwd: /home/user/project\n" +
		"repository: github/copilot-cli\n" +
		"branch: main\n" +
		"summary: Test session\n" +
		"created_at: \"2026-02-28T09:00:00Z\"\n" +
		"updated_at: \"2026-02-28T10:00:00Z\"\n"
	if err := os.WriteFile(filepath.Join(dir, "workspace.yaml"), []byte(richYAML), 0o644); err != nil {
		t.Fatal(err)
	}

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
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if len(decoded) != 1 {
		t.Fatalf("expected 1 session, got %d", len(decoded))
	}

	obj := decoded[0]
	// All fields from the JSON schema must be present.
	want := []string{"id", "created_at", "updated_at", "cwd", "repository", "branch", "summary", "event_count", "size_bytes"}
	for _, key := range want {
		if _, ok := obj[key]; !ok {
			t.Errorf("expected key %q in JSON output", key)
		}
	}
	// No unexpected keys.
	wantSet := make(map[string]bool, len(want))
	for _, k := range want {
		wantSet[k] = true
	}
	for key := range obj {
		if !wantSet[key] {
			t.Errorf("unexpected key %q in JSON output", key)
		}
	}
}

// TestRunList_EmptyDir verifies the list command handles an empty directory gracefully.
func TestRunList_EmptyDir(t *testing.T) {
	root := t.TempDir()

	// Table format: header must still appear.
	var buf bytes.Buffer
	if err := cli.RunList(context.Background(), cli.ListOptions{
		SessionDir: root,
		JSON:       false,
		Out:        &buf,
	}); err != nil {
		t.Fatalf("RunList table empty: %v", err)
	}
	if !strings.Contains(buf.String(), "SESSION ID") {
		t.Error("expected header in table output even for empty session list")
	}

	// JSON format: must produce an empty array.
	buf.Reset()
	if err := cli.RunList(context.Background(), cli.ListOptions{
		SessionDir: root,
		JSON:       true,
		Out:        &buf,
	}); err != nil {
		t.Fatalf("RunList JSON empty: %v", err)
	}
	var decoded []any
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("expected valid JSON array: %v", err)
	}
	if len(decoded) != 0 {
		t.Errorf("expected empty JSON array, got %d elements", len(decoded))
	}
}

// TestRunList_ScanError verifies the list command returns an error for a non-existent directory.
func TestRunList_ScanError(t *testing.T) {
	var buf bytes.Buffer
	err := cli.RunList(context.Background(), cli.ListOptions{
		SessionDir: "/does/not/exist/at/all",
		Out:        &buf,
	})
	if err == nil {
		t.Error("expected error for non-existent session directory, got nil")
	}
}

// ─── RunDelete ────────────────────────────────────────────────────────────────

// TestRunDelete_DryRun_Preview verifies dry-run prints what would be deleted
// without touching the session directory on disk.
func TestRunDelete_DryRun_Preview(t *testing.T) {
	root := t.TempDir()
	makeSession(t, root, fixID1, "2026-02-28T09:47:00Z")

	var out bytes.Buffer
	if err := cli.RunDelete(context.Background(), cli.DeleteOptions{
		SessionDir: root,
		IDs:        []string{fixID1},
		DryRun:     true,
		Out:        &out,
	}); err != nil {
		t.Fatalf("RunDelete dry-run: %v", err)
	}

	if !strings.Contains(out.String(), "[DRY-RUN]") {
		t.Errorf("expected [DRY-RUN] in preview output, got:\n%s", out.String())
	}
	// Session directory must still exist.
	if _, err := os.Stat(filepath.Join(root, fixID1)); err != nil {
		t.Errorf("expected session to remain after dry-run: %v", err)
	}
}

// TestRunDelete_SingleID_Real verifies delete --id removes the session directory.
func TestRunDelete_SingleID_Real(t *testing.T) {
	root := t.TempDir()
	makeSession(t, root, fixID1, "2026-02-28T09:47:00Z")

	var out bytes.Buffer
	if err := cli.RunDelete(context.Background(), cli.DeleteOptions{
		SessionDir: root,
		IDs:        []string{fixID1},
		DryRun:     false,
		Out:        &out,
	}); err != nil {
		t.Fatalf("RunDelete real: %v", err)
	}

	if _, err := os.Stat(filepath.Join(root, fixID1)); !os.IsNotExist(err) {
		t.Errorf("expected session directory to be removed after delete, stat err: %v", err)
	}
	if !strings.Contains(out.String(), "deleted") {
		t.Errorf("expected success message in output, got:\n%s", out.String())
	}
}

// TestRunDelete_All_Yes verifies delete --all --yes removes all sessions.
func TestRunDelete_All_Yes(t *testing.T) {
	root := t.TempDir()
	makeSession(t, root, fixID1, "2026-02-28T09:47:00Z")
	makeSession(t, root, fixID2, "2026-02-28T10:18:00Z")

	var out bytes.Buffer
	if err := cli.RunDelete(context.Background(), cli.DeleteOptions{
		SessionDir: root,
		All:        true,
		DryRun:     false,
		Yes:        true,
		Out:        &out,
	}); err != nil {
		t.Fatalf("RunDelete --all --yes: %v", err)
	}

	for _, id := range []string{fixID1, fixID2} {
		if _, err := os.Stat(filepath.Join(root, id)); !os.IsNotExist(err) {
			t.Errorf("expected session %s to be removed", id)
		}
	}
}

// TestRunDelete_All_WithoutYes verifies delete --all without --yes returns an error.
func TestRunDelete_All_WithoutYes(t *testing.T) {
	root := t.TempDir()
	makeSession(t, root, fixID1, "2026-02-28T09:47:00Z")

	err := cli.RunDelete(context.Background(), cli.DeleteOptions{
		SessionDir: root,
		All:        true,
		DryRun:     false,
		Yes:        false,
		Out:        &bytes.Buffer{},
	})
	if err == nil {
		t.Error("expected error for --all without --yes, got nil")
	}
	if !strings.Contains(err.Error(), "--yes") {
		t.Errorf("expected error mentioning --yes, got: %v", err)
	}
}

// TestRunDelete_IDAndAllConflict verifies delete returns an error when
// both --id and --all are specified.
func TestRunDelete_IDAndAllConflict(t *testing.T) {
	root := t.TempDir()
	err := cli.RunDelete(context.Background(), cli.DeleteOptions{
		SessionDir: root,
		IDs:        []string{fixID1},
		All:        true,
		DryRun:     true,
		Out:        &bytes.Buffer{},
	})
	if err == nil {
		t.Error("expected error for --id + --all, got nil")
	}
}

// TestRunDelete_InvalidUUID verifies delete returns an error for a malformed session ID.
func TestRunDelete_InvalidUUID(t *testing.T) {
	root := t.TempDir()
	err := cli.RunDelete(context.Background(), cli.DeleteOptions{
		SessionDir: root,
		IDs:        []string{"not-a-uuid"},
		DryRun:     true,
		Out:        &bytes.Buffer{},
	})
	if err == nil {
		t.Error("expected error for invalid UUID, got nil")
	}
}

// TestRunDelete_SessionNotFound verifies delete returns an error when the given
// session ID does not exist in the scanned directory.
func TestRunDelete_SessionNotFound(t *testing.T) {
	root := t.TempDir()
	makeSession(t, root, fixID1, "2026-02-28T09:47:00Z")

	err := cli.RunDelete(context.Background(), cli.DeleteOptions{
		SessionDir: root,
		IDs:        []string{fixID2}, // fixID2 not present on disk
		DryRun:     true,
		Out:        &bytes.Buffer{},
	})
	if err == nil {
		t.Error("expected error for session not found, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' in error, got: %v", err)
	}
}

// TestRunDelete_NoSessions verifies delete prints "No sessions found" when
// the directory is empty (no targets resolved).
func TestRunDelete_NoSessions(t *testing.T) {
	root := t.TempDir()

	var out bytes.Buffer
	if err := cli.RunDelete(context.Background(), cli.DeleteOptions{
		SessionDir: root,
		All:        true,
		DryRun:     false,
		Yes:        true,
		Out:        &out,
	}); err != nil {
		t.Fatalf("RunDelete on empty dir: %v", err)
	}
	if !strings.Contains(out.String(), "No sessions found") {
		t.Errorf("expected 'No sessions found' message, got:\n%s", out.String())
	}
}
