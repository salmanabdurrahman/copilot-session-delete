// Package tui — lipgloss style definitions.
package tui

import "github.com/charmbracelet/lipgloss"

// Adaptive primary text colour: readable on both dark and light terminals.
var colorPrimary = lipgloss.AdaptiveColor{Light: "#24292e", Dark: "#f0f6fc"}

var (
	styleHeader = lipgloss.NewStyle().
			Bold(true).
			Background(lipgloss.Color("#24292e")).
			Foreground(lipgloss.Color("#ffffff")).
			PaddingLeft(1).
			PaddingRight(1)

	styleFooter = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6a737d"))

	styleColHeader = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.AdaptiveColor{Light: "#24292e", Dark: "#f0f6fc"})

	// styleCursor highlights the row the cursor is on.
	styleCursor = lipgloss.NewStyle().
			Background(lipgloss.Color("#0366d6")).
			Foreground(lipgloss.Color("#ffffff"))

	// styleSelectedRow tints selected (but non-cursor) rows yellow.
	styleSelectedRow = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#f9c513"))

	// styleCursorSelected combines cursor bg + selected fg.
	styleCursorSelected = lipgloss.NewStyle().
				Background(lipgloss.Color("#0366d6")).
				Foreground(lipgloss.Color("#f9c513")).
				Bold(true)

	styleError = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#d73a49"))

	styleSuccess = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#28a745"))

	styleWarning = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#f9c513")).
			Bold(true)

	styleDim = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6a737d"))

	styleBold = lipgloss.NewStyle().Bold(true)

	styleModalBorder = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#d73a49")).
				PaddingTop(1).
				PaddingBottom(1).
				PaddingLeft(2).
				PaddingRight(2)

	styleDetailBorder = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.AdaptiveColor{Light: "#24292e", Dark: "#f0f6fc"}).
				PaddingTop(0).
				PaddingBottom(0).
				PaddingLeft(1).
				PaddingRight(1)
)
