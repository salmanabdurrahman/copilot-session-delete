// Package safety provides validation and path-guard functions for all destructive operations.
// Every check in this package must pass before any deletion is allowed.
package safety

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

// uuidV4RE matches a canonical lowercase UUID v4.
var uuidV4RE = regexp.MustCompile(
	`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`,
)

// ValidateUUID returns an error if s is not a valid UUID v4.
func ValidateUUID(s string) error {
	if s == "" {
		return fmt.Errorf("session ID must not be empty")
	}
	if !uuidV4RE.MatchString(strings.ToLower(s)) {
		return fmt.Errorf("invalid session ID %q: must be a UUID v4 (e.g. 86334621-8152-4e67-b322-9f139d6c0a57)", s)
	}
	return nil
}

// ValidateDescendant returns an error if target is not strictly under root.
// Both paths must be absolute and already resolved (no symlinks).
func ValidateDescendant(root, target string) error {
	if root == "" || target == "" {
		return fmt.Errorf("root and target path must not be empty")
	}
	// Ensure root ends with separator so prefix check is exact.
	rootWithSep := filepath.Clean(root) + string(filepath.Separator)
	cleanTarget := filepath.Clean(target)

	if cleanTarget == filepath.Clean(root) {
		return fmt.Errorf("target path must not be the session root itself: %q", target)
	}
	if !strings.HasPrefix(cleanTarget, rootWithSep) {
		return fmt.Errorf("target path %q is outside session root %q", target, root)
	}
	return nil
}
