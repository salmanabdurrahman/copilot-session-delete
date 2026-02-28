package platform_test

import (
	"os"
	"strings"
	"testing"

	"github.com/salmanabdurrahman/copilot-session-delete/internal/core/platform"
)

func TestSessionDir_WithOverride(t *testing.T) {
	want := "/tmp/my-custom-sessions"
	t.Setenv(platform.EnvOverride, want)

	got, err := platform.SessionDir()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestSessionDir_Default(t *testing.T) {
	// Remove override if set.
	t.Setenv(platform.EnvOverride, "")

	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot determine home dir:", err)
	}

	got, err := platform.SessionDir()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(got, home) {
		t.Errorf("expected path under %q, got %q", home, got)
	}
	if !strings.HasSuffix(got, "session-state") {
		t.Errorf("expected path to end with session-state, got %q", got)
	}
}
