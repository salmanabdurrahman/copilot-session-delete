package main_test

import (
	"os/exec"
	"testing"
)

// TestBinaryCompiles verifies the binary can be built from source.
// It runs "go build ." from the package directory (i.e., cmd/copilot-session-delete).
func TestBinaryCompiles(t *testing.T) {
	out := t.TempDir() + "/copilot-session-delete"
	cmd := exec.Command("go", "build", "-o", out, ".")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build failed:\n%s", output)
	}
}
