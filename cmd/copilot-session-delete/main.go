// copilot-session-delete is a TUI + CLI tool for browsing and deleting
// local GitHub Copilot CLI session data stored in ~/.copilot/session-state/.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/salmanabdurrahman/copilot-session-delete/internal/app/cli"
	"github.com/salmanabdurrahman/copilot-session-delete/internal/app/tui"
	"github.com/salmanabdurrahman/copilot-session-delete/internal/core/platform"
)

// Version is set at build time via -ldflags.
var Version = "dev"

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := newRootCmd().ExecuteContext(ctx); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	var sessionDir string

	root := &cobra.Command{
		Use:   "copilot-session-delete",
		Short: "Browse and delete local GitHub Copilot CLI sessions",
		Long: `copilot-session-delete lets you view and safely remove local GitHub Copilot CLI
chat sessions stored in ~/.copilot/session-state/ (or %USERPROFILE%\.copilot\session-state\ on Windows).

Running without a subcommand launches the interactive TUI.`,
		Version:       Version,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			dir, err := resolveSessionDir(cmd, sessionDir)
			if err != nil {
				return err
			}
			return tui.Run(dir)
		},
	}

	root.PersistentFlags().StringVar(&sessionDir, "session-dir", "",
		"Override the session-state directory path (default: ~/.copilot/session-state)")

	root.AddCommand(
		newListCmd(&sessionDir),
		newDeleteCmd(&sessionDir),
	)

	return root
}

func newListCmd(sessionDirFlag *string) *cobra.Command {
	var jsonOut bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all local Copilot CLI sessions",
		RunE: func(cmd *cobra.Command, _ []string) error {
			dir, err := resolveSessionDir(cmd, *sessionDirFlag)
			if err != nil {
				return err
			}
			return cli.RunList(cmd.Context(), cli.ListOptions{
				SessionDir: dir,
				JSON:       jsonOut,
				Out:        cmd.OutOrStdout(),
			})
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output sessions as JSON")
	return cmd
}

func newDeleteCmd(sessionDirFlag *string) *cobra.Command {
	var (
		id     string
		ids    string
		all    bool
		dryRun bool
		yes    bool
	)

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete one or more Copilot CLI sessions",
		Long: `Delete one or more local Copilot CLI sessions by ID.

By default this runs in dry-run mode and only shows what would be deleted.
Use --dry-run=false --yes to perform a real deletion.

Examples:
  # Preview what would be deleted
  copilot-session-delete delete --id 86334621-8152-4e67-b322-9f139d6c0a57

  # Delete a single session for real
  copilot-session-delete delete --id 86334621-8152-4e67-b322-9f139d6c0a57 --dry-run=false --yes

  # Delete multiple sessions
  copilot-session-delete delete --ids 86334621-...,c0c723f4-... --dry-run=false --yes

  # Delete all sessions
  copilot-session-delete delete --all --yes`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			dir, err := resolveSessionDir(cmd, *sessionDirFlag)
			if err != nil {
				return err
			}

			var targetIDs []string
			if id != "" {
				targetIDs = append(targetIDs, id)
			}
			if ids != "" {
				for _, part := range strings.Split(ids, ",") {
					part = strings.TrimSpace(part)
					if part != "" {
						targetIDs = append(targetIDs, part)
					}
				}
			}

			return cli.RunDelete(cmd.Context(), cli.DeleteOptions{
				SessionDir: dir,
				IDs:        targetIDs,
				All:        all,
				DryRun:     dryRun,
				Yes:        yes,
				Out:        cmd.OutOrStdout(),
				ErrOut:     cmd.ErrOrStderr(),
			})
		},
	}

	cmd.Flags().StringVar(&id, "id", "", "Session ID (UUID) to delete")
	cmd.Flags().StringVar(&ids, "ids", "", "Comma-separated list of session IDs to delete")
	cmd.Flags().BoolVar(&all, "all", false, "Delete all sessions (requires --yes)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", true, "Preview what would be deleted without doing it")
	cmd.Flags().BoolVar(&yes, "yes", false, "Confirm destructive operation")

	return cmd
}

// resolveSessionDir determines the session directory from the flag value or
// the platform default.
func resolveSessionDir(cmd *cobra.Command, flagValue string) (string, error) {
	if flagValue != "" {
		return flagValue, nil
	}
	dir, err := platform.SessionDir()
	if err != nil {
		return "", fmt.Errorf("could not determine session directory: %w\n\nSet --session-dir or the %s environment variable.", err, platform.EnvOverride)
	}
	return dir, nil
}
