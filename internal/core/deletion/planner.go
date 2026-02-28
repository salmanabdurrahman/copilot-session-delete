// Package deletion provides the Planner and Executor for safely removing
// Copilot CLI session directories.
package deletion

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/salmanabdurrahman/copilot-session-delete/internal/core/safety"
	"github.com/salmanabdurrahman/copilot-session-delete/internal/core/session"
)

// Plan describes what will be deleted for a single session.
type Plan struct {
	// Session is the target to delete.
	Session session.Session

	// NotFound is true when the session directory no longer exists on disk.
	// Such plans are skipped by the Executor (treated as success).
	NotFound bool
}

// Result reports the outcome of a single session deletion.
type Result struct {
	SessionID string
	Success   bool
	NotFound  bool // session was already gone — treated as success
	DryRun    bool
	Err       error
}

// Planner builds deletion plans for a set of session IDs.
type Planner struct {
	root string
}

// NewPlanner creates a Planner for the given session-state root.
func NewPlanner(root string) *Planner {
	return &Planner{root: root}
}

// Build creates Plans for the given sessions, applying safety checks.
// It returns an error only for unrecoverable configuration issues.
// Per-session issues (not found, safety violation) are embedded in the Plan.
func (p *Planner) Build(sessions []session.Session) ([]Plan, error) {
	rootCanonical, err := filepath.EvalSymlinks(p.root)
	if err != nil {
		return nil, fmt.Errorf("resolve session root %q: %w", p.root, err)
	}

	plans := make([]Plan, 0, len(sessions))
	for _, s := range sessions {
		if err := safety.ValidateUUID(s.ID); err != nil {
			return nil, fmt.Errorf("invalid session ID in plan: %w", err)
		}

		canonical, err := safeCanonical(s.RootPath)
		if os.IsNotExist(err) {
			plans = append(plans, Plan{Session: s, NotFound: true})
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("resolve path for session %s: %w", s.ID, err)
		}

		if err := safety.ValidateDescendant(rootCanonical, canonical); err != nil {
			return nil, fmt.Errorf("safety check failed for session %s: %w", s.ID, err)
		}

		updated := s
		updated.RootPath = canonical
		plans = append(plans, Plan{Session: updated})
	}
	return plans, nil
}

// safeCanonical resolves p to its canonical path, handling the case where
// the path does not yet exist (returns os.ErrNotExist).
func safeCanonical(p string) (string, error) {
	abs, err := filepath.Abs(p)
	if err != nil {
		return "", fmt.Errorf("abs path: %w", err)
	}
	canon, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return "", err // caller checks os.IsNotExist
	}
	return canon, nil
}
