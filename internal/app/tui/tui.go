// Package tui implements the interactive Bubble Tea TUI for browsing and
// deleting Copilot CLI sessions. Phase 1 stub — full model in Phase 3.
package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// Run starts the TUI application. It is the main entrypoint for the
// interactive mode and blocks until the user exits.
func Run(sessionDir string) error {
	m := initialModel(sessionDir)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

// model is the Bubble Tea application model. Fields will be expanded in Phase 3.
type model struct {
	sessionDir string
	ready      bool
	quitting   bool
}

func initialModel(sessionDir string) model {
	return model{sessionDir: sessionDir}
}

// Init is the Bubble Tea init hook.
func (m model) Init() tea.Cmd {
	return nil
}

// Update handles messages and key events.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.ready = true
	}
	return m, nil
}

// View renders the current UI.
func (m model) View() string {
	if m.quitting {
		return ""
	}
	return fmt.Sprintf(
		"\n  copilot-session-delete\n\n  Session dir: %s\n\n  [Phase 3: TUI not yet implemented]\n  Press q to quit.\n",
		m.sessionDir,
	)
}
