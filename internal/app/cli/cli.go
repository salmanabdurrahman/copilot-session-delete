// Package cli implements the non-interactive list and delete commands.
package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/salmanabdurrahman/copilot-session-delete/internal/adapters/output"
	"github.com/salmanabdurrahman/copilot-session-delete/internal/core/deletion"
	"github.com/salmanabdurrahman/copilot-session-delete/internal/core/safety"
	"github.com/salmanabdurrahman/copilot-session-delete/internal/core/session"
)

// ListOptions holds parameters for the list command.
type ListOptions struct {
	SessionDir string
	JSON       bool
	Out        io.Writer
}

// RunList lists all sessions in the given directory.
func RunList(ctx context.Context, opts ListOptions) error {
	if opts.Out == nil {
		opts.Out = os.Stdout
	}
	sessions, err := loadSessions(ctx, opts.SessionDir)
	if err != nil {
		return err
	}

	format := output.FormatTable
	if opts.JSON {
		format = output.FormatJSON
	}
	return output.PrintSessions(opts.Out, sessions, format)
}

// DeleteOptions holds parameters for the delete command.
type DeleteOptions struct {
	SessionDir string
	IDs        []string // single or multiple IDs
	All        bool
	DryRun     bool
	Yes        bool
	Out        io.Writer
	ErrOut     io.Writer
}

// RunDelete deletes the specified sessions.
func RunDelete(ctx context.Context, opts DeleteOptions) error {
	if opts.Out == nil {
		opts.Out = os.Stdout
	}
	if opts.ErrOut == nil {
		opts.ErrOut = os.Stderr
	}

	// Validate flag combinations.
	if opts.All && len(opts.IDs) > 0 {
		return fmt.Errorf("--all and --id/--ids cannot be used together")
	}
	if opts.All && !opts.Yes {
		return fmt.Errorf("--all requires --yes to prevent accidental deletion")
	}

	// Validate all provided IDs up front.
	for _, id := range opts.IDs {
		if err := safety.ValidateUUID(strings.TrimSpace(id)); err != nil {
			return err
		}
	}

	sessions, err := loadSessions(ctx, opts.SessionDir)
	if err != nil {
		return err
	}

	// Resolve target sessions.
	targets, err := resolveTargets(sessions, opts)
	if err != nil {
		return err
	}
	if len(targets) == 0 {
		fmt.Fprintln(opts.Out, "No sessions found.")
		return nil
	}

	planner := deletion.NewPlanner(opts.SessionDir)
	plans, err := planner.Build(targets)
	if err != nil {
		return fmt.Errorf("build deletion plan: %w", err)
	}

	// Dry-run: print preview and return.
	if opts.DryRun {
		fmt.Fprintf(opts.Out, "[DRY-RUN] Would delete %d session(s):\n", len(plans))
		for _, p := range plans {
			if p.NotFound {
				fmt.Fprintf(opts.Out, "  • %s  (not found — would skip)\n", p.Session.ID)
			} else {
				fmt.Fprintf(opts.Out, "  • %s  (%s)\n", p.Session.ID, p.Session.RootPath)
			}
		}
		fmt.Fprintln(opts.Out, "\nRun with --dry-run=false --yes to delete for real.")
		return nil
	}

	// Real delete.
	executor := deletion.NewExecutor(false)
	results := executor.Execute(ctx, plans)

	var succeeded, failed int
	for r := range results {
		if r.Success {
			succeeded++
		} else {
			failed++
			fmt.Fprintf(opts.ErrOut, "Error: failed to delete %s: %v\n", r.SessionID, r.Err)
		}
	}

	if failed == 0 {
		fmt.Fprintf(opts.Out, "✓ %d session(s) deleted.\n", succeeded)
		return nil
	}
	if succeeded == 0 {
		return fmt.Errorf("all deletions failed (%d)", failed)
	}
	// Partial failure: exit code 2 will be set by the caller.
	fmt.Fprintf(opts.Out, "⚠ %d deleted, %d failed.\n", succeeded, failed)
	return fmt.Errorf("partial failure: %d session(s) could not be deleted", failed)
}

// loadSessions scans sessionDir and enriches each session with metadata.
func loadSessions(ctx context.Context, sessionDir string) ([]session.Session, error) {
	scanner := session.NewScanner(sessionDir)
	sessions, err := scanner.Scan(ctx)
	if err != nil {
		return nil, err
	}
	for i := range sessions {
		session.EnrichMetadata(&sessions[i])
	}
	return sessions, nil
}

// resolveTargets filters sessions to the ones the user wants to delete.
func resolveTargets(all []session.Session, opts DeleteOptions) ([]session.Session, error) {
	if opts.All {
		return all, nil
	}

	// Build lookup map.
	byID := make(map[string]session.Session, len(all))
	for _, s := range all {
		byID[strings.ToLower(s.ID)] = s
	}

	var targets []session.Session
	for _, id := range opts.IDs {
		id = strings.ToLower(strings.TrimSpace(id))
		s, ok := byID[id]
		if !ok {
			return nil, fmt.Errorf("session not found: %s", id)
		}
		targets = append(targets, s)
	}
	return targets, nil
}
