// Package platform resolves OS-specific paths for Copilot CLI session storage.
package platform

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

const (
	// EnvOverride is the environment variable users can set to override the default session directory.
	EnvOverride = "COPILOT_SESSION_DIR"

	// defaultSubPath is the path relative to the user's home directory.
	defaultSubPath = ".copilot/session-state"
)

// SessionDir returns the resolved session-state directory path.
// It respects the COPILOT_SESSION_DIR environment variable as an override.
func SessionDir() (string, error) {
	if override := os.Getenv(EnvOverride); override != "" {
		abs, err := filepath.Abs(override)
		if err != nil {
			return "", fmt.Errorf("resolve override path %q: %w", override, err)
		}
		return abs, nil
	}
	return defaultSessionDir()
}

// defaultSessionDir returns the platform-default session directory path.
func defaultSessionDir() (string, error) {
	home, err := homeDir()
	if err != nil {
		return "", fmt.Errorf("determine home directory: %w", err)
	}
	return filepath.Join(home, defaultSubPath), nil
}

// homeDir returns the current user's home directory in a cross-platform way.
func homeDir() (string, error) {
	if runtime.GOOS == "windows" {
		if h := os.Getenv("USERPROFILE"); h != "" {
			return h, nil
		}
		// Fallback: HOMEDRIVE + HOMEPATH
		drive := os.Getenv("HOMEDRIVE")
		path := os.Getenv("HOMEPATH")
		if drive != "" && path != "" {
			return drive + path, nil
		}
		return "", fmt.Errorf("USERPROFILE, HOMEDRIVE, and HOMEPATH environment variables are not set")
	}
	// macOS / Linux
	if h := os.Getenv("HOME"); h != "" {
		return h, nil
	}
	return "", fmt.Errorf("HOME environment variable is not set")
}
