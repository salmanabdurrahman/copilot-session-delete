package session

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/salmanabdurrahman/copilot-session-delete/internal/core/safety"
)

// Scanner discovers Session entries from a session-state root directory.
type Scanner struct {
	root string
}

// NewScanner returns a Scanner that reads from the given root path.
func NewScanner(root string) *Scanner {
	return &Scanner{root: root}
}

// Scan reads all valid UUID v4 subdirectories under root and returns them as
// a slice of Session values. It skips non-directory entries and entries whose
// names are not valid UUID v4 values; these are not treated as errors.
//
// The returned Sessions have only ID and RootPath populated.
// Use MetadataExtractor to fill the remaining fields.
func (s *Scanner) Scan(ctx context.Context) ([]Session, error) {
	info, err := os.Stat(s.root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("session directory not found: %s", s.root)
		}
		return nil, fmt.Errorf("read session directory %s: %w", s.root, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("session path is not a directory: %s", s.root)
	}

	entries, err := os.ReadDir(s.root)
	if err != nil {
		return nil, fmt.Errorf("list session directory %s: %w", s.root, err)
	}

	var sessions []Session
	for _, e := range entries {
		if ctx.Err() != nil {
			return sessions, ctx.Err()
		}
		if !e.IsDir() {
			continue
		}
		if err := safety.ValidateUUID(e.Name()); err != nil {
			// Not a UUID — silently skip (could be other data).
			continue
		}
		sessions = append(sessions, Session{
			ID:       e.Name(),
			RootPath: filepath.Join(s.root, e.Name()),
		})
	}
	return sessions, nil
}
