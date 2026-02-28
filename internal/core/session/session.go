// Package session defines the Session entity and provides the Scanner for
// discovering Copilot CLI sessions from the local session-state directory.
package session

import (
	"path/filepath"
	"time"
)

// Session represents a single Copilot CLI chat session discovered on disk.
type Session struct {
	// ID is the UUID v4 that identifies this session (also the folder name).
	ID string

	// RootPath is the absolute canonical path to the session folder.
	RootPath string

	// CreatedAt is when the session was started.
	// Zero value if workspace.yaml is missing or unreadable.
	CreatedAt time.Time

	// UpdatedAt is when the session was last active.
	// Falls back to directory mtime if workspace.yaml is unavailable.
	UpdatedAt time.Time

	// CWD is the working directory active when the session started.
	CWD string

	// Repository is the git remote name (e.g. "github/copilot-cli"), if available.
	Repository string

	// Branch is the active git branch, if available.
	Branch string

	// Summary is a short human-readable title for the session.
	Summary string

	// EventCount is the number of valid JSON lines in events.jsonl.
	// -1 indicates the file was not found or not readable.
	EventCount int

	// SizeBytes is the total size of all files under RootPath.
	// 0 if not yet calculated.
	SizeBytes int64

	// MetadataErr holds any non-fatal error encountered while reading workspace.yaml.
	// A session with MetadataErr is still displayed but marked with a warning indicator.
	MetadataErr error
}

// Label returns a short human-readable identifier for the session suitable
// for display in table rows and confirmation dialogs.
// Priority: Repository > basename(CWD) > session ID prefix.
func (s Session) Label() string {
	if s.Repository != "" {
		return s.Repository
	}
	if s.CWD != "" {
		return filepath.Base(s.CWD)
	}
	if len(s.ID) >= 8 {
		return s.ID[:8] + "…"
	}
	return s.ID
}
