// Package output provides formatters for session data, supporting both
// human-readable table output and machine-readable JSON output.
package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/salmanabdurrahman/copilot-session-delete/internal/core/session"
)

// Format controls the output format.
type Format string

const (
	FormatTable Format = "table"
	FormatJSON  Format = "json"
)

// PrintSessions writes a list of sessions to w using the given format.
func PrintSessions(w io.Writer, sessions []session.Session, format Format) error {
	switch format {
	case FormatJSON:
		return printJSON(w, sessions)
	default:
		return printTable(w, sessions)
	}
}

func printTable(w io.Writer, sessions []session.Session) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "SESSION ID\tUPDATED AT\tCWD / REPO\tEVENTS")
	fmt.Fprintln(tw, strings.Repeat("-", 10)+"\t"+strings.Repeat("-", 16)+"\t"+strings.Repeat("-", 20)+"\t"+strings.Repeat("-", 6))
	for _, s := range sessions {
		updated := formatTime(s.UpdatedAt)
		cwdRepo := truncate(sessionLabel(s), 30)
		events := formatEventCount(s.EventCount)
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", truncate(s.ID, 36), updated, cwdRepo, events)
	}
	return tw.Flush()
}

// jsonSession is the JSON-serialisable representation of a session.
type jsonSession struct {
	ID         string `json:"id"`
	CreatedAt  string `json:"created_at,omitempty"`
	UpdatedAt  string `json:"updated_at,omitempty"`
	CWD        string `json:"cwd,omitempty"`
	Repository string `json:"repository,omitempty"`
	Branch     string `json:"branch,omitempty"`
	Summary    string `json:"summary,omitempty"`
	EventCount int    `json:"event_count"`
	SizeBytes  int64  `json:"size_bytes"`
}

func printJSON(w io.Writer, sessions []session.Session) error {
	out := make([]jsonSession, len(sessions))
	for i, s := range sessions {
		out[i] = jsonSession{
			ID:         s.ID,
			CreatedAt:  formatTimeISO(s.CreatedAt),
			UpdatedAt:  formatTimeISO(s.UpdatedAt),
			CWD:        s.CWD,
			Repository: s.Repository,
			Branch:     s.Branch,
			Summary:    s.Summary,
			EventCount: s.EventCount,
			SizeBytes:  s.SizeBytes,
		}
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}

func sessionLabel(s session.Session) string {
	return s.Label()
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return "—"
	}
	return t.Local().Format("2006-01-02 15:04")
}

func formatTimeISO(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}

func formatEventCount(n int) string {
	if n < 0 {
		return "?"
	}
	return fmt.Sprintf("%d", n)
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}

// FormatSize returns a human-readable approximation of a byte count.
// Examples: "~2.1 MB", "~512 KB", "~900 B".
func FormatSize(bytes int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)
	switch {
	case bytes >= GB:
		return fmt.Sprintf("~%.1f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("~%.1f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("~%.1f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
