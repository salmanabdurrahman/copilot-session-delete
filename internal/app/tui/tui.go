// Package tui implements the interactive Bubble Tea TUI for browsing and
// deleting Copilot CLI sessions.
package tui

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/salmanabdurrahman/copilot-session-delete/internal/adapters/output"
	"github.com/salmanabdurrahman/copilot-session-delete/internal/core/session"
)

// ─── View states ─────────────────────────────────────────────────────────────

// viewState tracks which screen is currently active.
type viewState int

const (
	viewList    viewState = iota // main session list (default)
	viewDetail                   // single-session detail panel
	viewConfirm                  // delete confirmation modal
)

// ─── Column layout modes ──────────────────────────────────────────────────────

type colMode int

const (
	colsFull     colMode = iota // all columns (width >= 80)
	colsNoEvents                // hide EVENTS column (width >= 60)
	colsMini                    // ID + time only (width >= 40)
)

// ─── Async messages ───────────────────────────────────────────────────────────

// sessionsLoadedMsg is delivered when the async session scan finishes.
type sessionsLoadedMsg struct {
	sessions []session.Session
	err      error
}

// ─── Model ───────────────────────────────────────────────────────────────────

// model is the Bubble Tea root model for the TUI.
type model struct {
	// config
	sessionDir string

	// data
	sessions []session.Session
	filtered []session.Session
	loadErr  error
	loading  bool

	// list navigation
	cursor   int
	offset   int
	listH    int // number of visible rows in the list area
	selected map[string]bool

	// search
	searchInput  textinput.Model
	searchActive bool

	// active view
	view viewState

	// detail panel — index into filtered
	detailIdx int

	// confirm modal — sessions targeted for deletion
	deleteTargets []session.Session

	// terminal dimensions
	width  int
	height int

	// status/notification message shown below the list
	statusMsg   string
	statusIsErr bool
}

func initialModel(sessionDir string) model {
	ti := textinput.New()
	ti.Placeholder = "filter sessions…"
	ti.CharLimit = 100
	ti.Prompt = "/ "
	return model{
		sessionDir:  sessionDir,
		selected:    make(map[string]bool),
		searchInput: ti,
		loading:     true,
	}
}

// ─── Entry point ──────────────────────────────────────────────────────────────

// Run starts the TUI application and blocks until the user exits.
func Run(sessionDir string) error {
	m := initialModel(sessionDir)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

// ─── Bubble Tea interface ─────────────────────────────────────────────────────

// Init triggers the async session load as soon as the program starts.
func (m model) Init() tea.Cmd {
	return loadSessionsCmd(m.sessionDir)
}

func loadSessionsCmd(dir string) tea.Cmd {
	return func() tea.Msg {
		sessions, err := session.ScanAndEnrich(context.Background(), dir)
		return sessionsLoadedMsg{sessions: sessions, err: err}
	}
}

// Update handles all incoming messages and key events.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.listH = max(0, m.height-7)
		return m, nil

	case sessionsLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.loadErr = msg.err
		} else {
			m.sessions = msg.sessions
			m.applyFilter()
		}
		return m, nil

	case tea.KeyMsg:
		// ctrl+c quits from any view.
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		switch m.view {
		case viewList:
			return m.updateList(msg)
		case viewDetail:
			return m.updateDetail(msg)
		case viewConfirm:
			return m.updateConfirm(msg)
		}
	}
	return m, nil
}

// ─── Key handlers ─────────────────────────────────────────────────────────────

func (m model) updateList(msg tea.KeyMsg) (model, tea.Cmd) {
	// Forward keystrokes to the textinput while search is active.
	if m.searchActive {
		switch msg.String() {
		case "esc":
			m.searchInput.SetValue("")
			m.searchActive = false
			m.searchInput.Blur()
			m.applyFilter()
			return m, nil
		case "enter":
			m.searchActive = false
			m.searchInput.Blur()
			return m, nil
		default:
			var cmd tea.Cmd
			m.searchInput, cmd = m.searchInput.Update(msg)
			m.applyFilter()
			return m, cmd
		}
	}

	switch msg.String() {
	case "q":
		return m, tea.Quit
	case "up", "k":
		m.moveCursor(-1)
	case "down", "j":
		m.moveCursor(1)
	case "pgup":
		m.moveCursor(-m.listH)
	case "pgdown":
		m.moveCursor(m.listH)
	case "g":
		m.cursor = 0
		m.offset = 0
	case "G":
		if len(m.filtered) > 0 {
			m.cursor = len(m.filtered) - 1
			m.scrollToCursor()
		}
	case "/":
		m.searchActive = true
		return m, m.searchInput.Focus()
	case " ":
		m.toggleSelect()
	case "a":
		m.toggleSelectAll()
	case "enter":
		if len(m.filtered) > 0 {
			m.detailIdx = m.cursor
			m.view = viewDetail
		}
	case "d":
		if targets := m.getDeleteTargets(); len(targets) > 0 {
			m.deleteTargets = targets
			m.statusMsg = ""
			m.view = viewConfirm
		}
	case "r":
		m.loading = true
		m.sessions = nil
		m.filtered = nil
		m.selected = make(map[string]bool)
		m.cursor = 0
		m.offset = 0
		m.loadErr = nil
		m.statusMsg = ""
		return m, loadSessionsCmd(m.sessionDir)
	}
	return m, nil
}

func (m model) updateDetail(msg tea.KeyMsg) (model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.view = viewList
	case "d":
		if m.detailIdx < len(m.filtered) {
			m.deleteTargets = []session.Session{m.filtered[m.detailIdx]}
			m.view = viewConfirm
		}
	}
	return m, nil
}

func (m model) updateConfirm(msg tea.KeyMsg) (model, tea.Cmd) {
	switch msg.String() {
	case "y":
		// Phase 3: confirm modal only — deletion is wired in Phase 4.
		n := len(m.deleteTargets)
		m.statusMsg = fmt.Sprintf("✓ %d session(s) queued for deletion (execution wired in Phase 4).", n)
		m.statusIsErr = false
		m.deleteTargets = nil
		m.view = viewList
	case "n", "esc":
		m.deleteTargets = nil
		m.view = viewList
	}
	return m, nil
}

// ─── View dispatcher ──────────────────────────────────────────────────────────

// View renders the full current UI state as a string.
func (m model) View() string {
	if m.width == 0 {
		return ""
	}
	if m.width < 40 {
		return "\n  Terminal too narrow. Please widen your terminal window.\n"
	}
	switch m.view {
	case viewDetail:
		return m.renderDetail()
	case viewConfirm:
		return m.renderConfirm()
	default:
		return m.renderList()
	}
}

// ─── List view ────────────────────────────────────────────────────────────────

func (m model) renderList() string {
	var b strings.Builder

	// 1. Header bar: title on left, selected/total counter on right.
	counter := fmt.Sprintf("%d/%d", m.selectedCount(), len(m.filtered))
	title := "copilot-session-delete  " + truncPath(m.sessionDir, m.width-len(counter)-6)
	b.WriteString(styleHeader.Width(m.width).Render(padRight(title, m.width-len(counter)-2) + counter))
	b.WriteByte('\n')

	// 2. Search bar.
	if m.searchActive {
		b.WriteString("  " + m.searchInput.View() + "\n")
	} else {
		q := m.searchInput.Value()
		hint := "/ to search"
		if q != "" {
			hint = "/ " + q
		}
		b.WriteString(styleDim.Render("  "+hint) + "\n")
	}

	// 3. Column headers + divider.
	b.WriteString(m.renderColHeaders() + "\n")
	b.WriteString(strings.Repeat("─", m.width) + "\n")

	// 4. Session rows.
	switch {
	case m.loading:
		b.WriteString("\n  Loading sessions…\n")
	case m.loadErr != nil:
		b.WriteString(styleError.Render(fmt.Sprintf("\n  ⚠ %v\n", m.loadErr)))
	case len(m.sessions) == 0:
		b.WriteString("\n  No sessions found.\n")
	case len(m.filtered) == 0:
		b.WriteString(fmt.Sprintf("\n  No sessions match %q.\n", m.searchInput.Value()))
	default:
		end := m.offset + m.listH
		if end > len(m.filtered) {
			end = len(m.filtered)
		}
		for i := m.offset; i < end; i++ {
			b.WriteString(m.renderRow(i))
			b.WriteByte('\n')
		}
	}

	// 5. Status message (success / error from last operation).
	if m.statusMsg != "" {
		b.WriteByte('\n')
		if m.statusIsErr {
			b.WriteString(styleError.Render("  " + m.statusMsg))
		} else {
			b.WriteString(styleSuccess.Render("  " + m.statusMsg))
		}
		b.WriteByte('\n')
	}

	// 6. Footer help bar.
	footer := " [↑/↓] navigate  [/] search  [space] select  [a] all  [d] delete  [enter] detail  [r] refresh  [q] quit"
	b.WriteString(styleFooter.Render(trunc(footer, m.width)))

	return b.String()
}

func (m model) renderColHeaders() string {
	switch m.currentColMode() {
	case colsMini:
		return fmt.Sprintf("  %-3s %-13s  %-16s", "", "SESSION ID", "UPDATED AT")
	case colsNoEvents:
		return fmt.Sprintf("  %-3s %-13s  %-16s  %-20s", "", "SESSION ID", "UPDATED AT", "CWD/REPO")
	default:
		return fmt.Sprintf("  %-3s %-13s  %-16s  %-20s  %6s", "", "SESSION ID", "UPDATED AT", "CWD/REPO", "EVENTS")
	}
}

func (m model) renderRow(idx int) string {
	s := m.filtered[idx]
	isCursor := idx == m.cursor
	isSelected := m.selected[s.ID]

	check := "[ ]"
	if isSelected {
		check = "[✓]"
	}

	id := trunc(s.ID, 13)

	ts := "—"
	if !s.UpdatedAt.IsZero() {
		ts = s.UpdatedAt.Format("2006-01-02 15:04")
	}

	lbl := trunc(s.Label(), 20)

	ev := "?"
	if s.EventCount >= 0 {
		ev = strconv.Itoa(s.EventCount)
	}

	var row string
	switch m.currentColMode() {
	case colsMini:
		row = fmt.Sprintf("  %s %-13s  %-16s", check, id, ts)
	case colsNoEvents:
		row = fmt.Sprintf("  %s %-13s  %-16s  %-20s", check, id, ts, lbl)
	default:
		row = fmt.Sprintf("  %s %-13s  %-16s  %-20s  %6s", check, id, ts, lbl, ev)
	}

	// Pad to full width so cursor/selected backgrounds fill the line.
	row = padRight(row, m.width)

	switch {
	case isCursor && isSelected:
		return styleCursorSelected.Render(row)
	case isCursor:
		return styleCursor.Render(row)
	case isSelected:
		return styleSelectedRow.Render(row)
	default:
		return row
	}
}

// ─── Detail view ──────────────────────────────────────────────────────────────

func (m model) renderDetail() string {
	if m.detailIdx >= len(m.filtered) {
		return m.renderList()
	}
	s := m.filtered[m.detailIdx]

	var content strings.Builder
	content.WriteString(fmt.Sprintf("ID      : %s\n", s.ID))
	if !s.CreatedAt.IsZero() {
		content.WriteString(fmt.Sprintf("Created : %s\n", s.CreatedAt.Format("2006-01-02 15:04:05")))
	}
	if !s.UpdatedAt.IsZero() {
		content.WriteString(fmt.Sprintf("Updated : %s\n", s.UpdatedAt.Format("2006-01-02 15:04:05")))
	}
	if s.CWD != "" {
		content.WriteString(fmt.Sprintf("CWD     : %s\n", s.CWD))
	}
	if s.Repository != "" {
		content.WriteString(fmt.Sprintf("Repo    : %s\n", s.Repository))
	}
	if s.Branch != "" {
		content.WriteString(fmt.Sprintf("Branch  : %s\n", s.Branch))
	}
	if s.Summary != "" {
		content.WriteString(fmt.Sprintf("Summary : %s\n", trunc(s.Summary, 50)))
	}

	ev := "?"
	if s.EventCount >= 0 {
		ev = strconv.Itoa(s.EventCount)
	}
	content.WriteString(fmt.Sprintf("Events  : %s\n", ev))

	if s.SizeBytes > 0 {
		content.WriteString(fmt.Sprintf("Size    : %s\n", output.FormatSize(s.SizeBytes)))
	}
	if s.MetadataErr != nil {
		content.WriteString(styleError.Render("⚠ Metadata: "+s.MetadataErr.Error()) + "\n")
	}

	panelWidth := min(m.width-4, 70)
	panel := styleDetailBorder.Width(panelWidth).Render(content.String())
	footer := "\n  " + styleDim.Render("[d] Delete this session    [esc] Back") + "\n"

	return "\n  " + styleBold.Render("Session Detail") + "\n\n" +
		indent(panel, 2) + footer
}

// ─── Confirm modal ────────────────────────────────────────────────────────────

func (m model) renderConfirm() string {
	n := len(m.deleteTargets)

	var body strings.Builder
	body.WriteString(styleWarning.Render(fmt.Sprintf("⚠  Delete %d session(s)?", n)) + "\n\n")

	const maxList = 5
	for i, s := range m.deleteTargets {
		if i >= maxList {
			body.WriteString(fmt.Sprintf("  … dan %d lainnya\n", n-maxList))
			break
		}
		ev := "?"
		if s.EventCount >= 0 {
			ev = strconv.Itoa(s.EventCount)
		}
		body.WriteString(fmt.Sprintf("  • %s (%s · %s events)\n",
			trunc(s.ID, 22), trunc(s.Label(), 15), ev))
	}

	var totalBytes int64
	for _, s := range m.deleteTargets {
		totalBytes += s.SizeBytes
	}
	if totalBytes > 0 {
		body.WriteString(fmt.Sprintf("\n  Total: %s will be removed.\n", output.FormatSize(totalBytes)))
	}

	body.WriteString(styleError.Render("\n  This action CANNOT be undone.") + "\n\n")
	body.WriteString("  [y] Delete       [n / esc] Cancel\n")

	modalWidth := min(m.width-8, 62)
	modal := styleModalBorder.Width(modalWidth).Render(body.String())
	return "\n" + indent(modal, 2)
}

// ─── Model helpers ─────────────────────────────────────────────────────────────

func (m model) currentColMode() colMode {
	switch {
	case m.width < 60:
		return colsMini
	case m.width < 80:
		return colsNoEvents
	default:
		return colsFull
	}
}

// applyFilter rebuilds m.filtered from m.sessions using the current search query.
func (m *model) applyFilter() {
	q := strings.ToLower(m.searchInput.Value())
	if q == "" {
		m.filtered = m.sessions
		m.clampCursor()
		return
	}
	out := make([]session.Session, 0, len(m.sessions))
	for _, s := range m.sessions {
		if matchSession(s, q) {
			out = append(out, s)
		}
	}
	m.filtered = out
	m.clampCursor()
}

func matchSession(s session.Session, q string) bool {
	return strings.Contains(strings.ToLower(s.ID), q) ||
		strings.Contains(strings.ToLower(s.CWD), q) ||
		strings.Contains(strings.ToLower(s.Repository), q) ||
		strings.Contains(strings.ToLower(s.Summary), q)
}

func (m *model) moveCursor(delta int) {
	if len(m.filtered) == 0 {
		return
	}
	m.cursor += delta
	m.clampCursor()
	m.scrollToCursor()
}

func (m *model) clampCursor() {
	if m.cursor < 0 {
		m.cursor = 0
	}
	if len(m.filtered) > 0 && m.cursor >= len(m.filtered) {
		m.cursor = len(m.filtered) - 1
	}
}

func (m *model) scrollToCursor() {
	if m.listH <= 0 {
		return
	}
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
	if m.cursor >= m.offset+m.listH {
		m.offset = m.cursor - m.listH + 1
	}
}

func (m *model) toggleSelect() {
	if len(m.filtered) == 0 {
		return
	}
	id := m.filtered[m.cursor].ID
	if m.selected[id] {
		delete(m.selected, id)
	} else {
		m.selected[id] = true
	}
}

func (m *model) toggleSelectAll() {
	if m.allVisibleSelected() {
		for _, s := range m.filtered {
			delete(m.selected, s.ID)
		}
	} else {
		for _, s := range m.filtered {
			m.selected[s.ID] = true
		}
	}
}

func (m model) allVisibleSelected() bool {
	if len(m.filtered) == 0 {
		return false
	}
	for _, s := range m.filtered {
		if !m.selected[s.ID] {
			return false
		}
	}
	return true
}

// getDeleteTargets returns selected sessions, or the cursor row if nothing selected.
func (m model) getDeleteTargets() []session.Session {
	var targets []session.Session
	for _, s := range m.filtered {
		if m.selected[s.ID] {
			targets = append(targets, s)
		}
	}
	if len(targets) == 0 && len(m.filtered) > 0 {
		targets = append(targets, m.filtered[m.cursor])
	}
	return targets
}

func (m model) selectedCount() int {
	return len(m.selected)
}

// ─── String helpers ────────────────────────────────────────────────────────────

// trunc truncates s to at most max runes, appending "…" when shortened.
func trunc(s string, max int) string {
	if max <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	if max == 1 {
		return "…"
	}
	return string(runes[:max-1]) + "…"
}

// padRight pads s with spaces on the right until it is at least width runes wide.
func padRight(s string, width int) string {
	runes := []rune(s)
	for len(runes) < width {
		runes = append(runes, ' ')
	}
	return string(runes)
}

// truncPath shortens a path from the left when it exceeds max characters.
func truncPath(p string, max int) string {
	if len(p) <= max {
		return p
	}
	return "…" + p[len(p)-max+1:]
}

// indent prepends n spaces to every non-empty line of s.
func indent(s string, n int) string {
	pad := strings.Repeat(" ", n)
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		if l != "" {
			lines[i] = pad + l
		}
	}
	return strings.Join(lines, "\n")
}
