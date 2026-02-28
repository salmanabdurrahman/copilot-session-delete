// Tests for the Bubble Tea TUI model (Phase 3).
// Uses internal package access so unexported messages and view states are reachable.
package tui

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/salmanabdurrahman/copilot-session-delete/internal/core/session"
)

// TestInitialModel verifies the initial model has sensible default values.
func TestInitialModel(t *testing.T) {
	m := initialModel("/fake/dir")
	if m.sessionDir != "/fake/dir" {
		t.Errorf("expected sessionDir %q, got %q", "/fake/dir", m.sessionDir)
	}
	if !m.loading {
		t.Error("expected loading=true on initial model")
	}
	if m.view != viewList {
		t.Errorf("expected initial view viewList, got %v", m.view)
	}
	if m.selected == nil {
		t.Error("expected non-nil selected map")
	}
	if m.cursor != 0 {
		t.Errorf("expected cursor 0, got %d", m.cursor)
	}
}

// TestUpdate_WindowSize verifies width, height, and listH are set on WindowSizeMsg.
func TestUpdate_WindowSize(t *testing.T) {
	m := initialModel("/fake/dir")
	m2, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	got := m2.(model)
	if got.width != 120 || got.height != 40 {
		t.Errorf("expected 120x40, got %dx%d", got.width, got.height)
	}
	if got.listH != 33 { // max(0, 40-7)
		t.Errorf("expected listH 33, got %d", got.listH)
	}
}

// TestUpdate_SessionsLoaded verifies sessions and filtered are populated on success.
func TestUpdate_SessionsLoaded(t *testing.T) {
	m := initialModel("/fake/dir")
	sessions := fixtureSessions()
	m2, _ := m.Update(sessionsLoadedMsg{sessions: sessions})
	got := m2.(model)

	if got.loading {
		t.Error("expected loading=false after sessions loaded")
	}
	if got.loadErr != nil {
		t.Errorf("unexpected loadErr: %v", got.loadErr)
	}
	if len(got.sessions) != 2 {
		t.Errorf("expected 2 sessions, got %d", len(got.sessions))
	}
	if len(got.filtered) != 2 {
		t.Errorf("expected 2 filtered sessions, got %d", len(got.filtered))
	}
}

// TestUpdate_SessionsLoadError verifies loadErr is set and loading cleared on error.
func TestUpdate_SessionsLoadError(t *testing.T) {
	m := initialModel("/fake/dir")
	m2, _ := m.Update(sessionsLoadedMsg{err: fmt.Errorf("permission denied")})
	got := m2.(model)

	if got.loading {
		t.Error("expected loading=false after error")
	}
	if got.loadErr == nil {
		t.Error("expected non-nil loadErr")
	}
	if len(got.sessions) != 0 {
		t.Errorf("expected 0 sessions on error, got %d", len(got.sessions))
	}
}

// TestUpdate_NavigateDown verifies cursor moves down on key "down".
func TestUpdate_NavigateDown(t *testing.T) {
	m := modelWithSessions()
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	got := m2.(model)
	if got.cursor != 1 {
		t.Errorf("expected cursor 1 after down, got %d", got.cursor)
	}
}

// TestUpdate_NavigateDown_Clamps verifies cursor does not exceed last index.
func TestUpdate_NavigateDown_Clamps(t *testing.T) {
	m := modelWithSessions()
	m.cursor = len(m.filtered) - 1 // already at last row
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	got := m2.(model)
	if got.cursor != len(got.filtered)-1 {
		t.Errorf("expected cursor to stay at last row, got %d", got.cursor)
	}
}

// TestUpdate_NavigateUp_AtTop verifies cursor stays at 0 when already at top.
func TestUpdate_NavigateUp_AtTop(t *testing.T) {
	m := modelWithSessions()
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyUp})
	got := m2.(model)
	if got.cursor != 0 {
		t.Errorf("expected cursor clamped at 0, got %d", got.cursor)
	}
}

// TestUpdate_ToggleSelect verifies space toggles a session's selected state.
func TestUpdate_ToggleSelect(t *testing.T) {
	m := modelWithSessions()
	id := m.filtered[0].ID

	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	got := m2.(model)
	if !got.selected[id] {
		t.Error("expected session to be selected after space")
	}

	m3, _ := got.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	got3 := m3.(model)
	if got3.selected[id] {
		t.Error("expected session to be deselected after second space")
	}
}

// TestUpdate_SelectAll verifies 'a' selects all visible sessions, and again deselects all.
func TestUpdate_SelectAll(t *testing.T) {
	m := modelWithSessions()

	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	got := m2.(model)
	for _, s := range got.filtered {
		if !got.selected[s.ID] {
			t.Errorf("expected session %s selected after 'a'", s.ID)
		}
	}

	m3, _ := got.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	got3 := m3.(model)
	if len(got3.selected) != 0 {
		t.Errorf("expected 0 selected after second 'a', got %d", len(got3.selected))
	}
}

// TestUpdate_SearchFilter verifies typing filters m.filtered in real-time.
func TestUpdate_SearchFilter(t *testing.T) {
	m := modelWithSessions()

	// Activate search.
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	got := m2.(model)
	if !got.searchActive {
		t.Fatal("expected searchActive after '/'")
	}

	// Type a query that matches only the first session ID prefix.
	for _, r := range "86334621" {
		got2, _ := got.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		got = got2.(model)
	}
	if len(got.filtered) != 1 {
		t.Errorf("expected 1 filtered session, got %d", len(got.filtered))
	}
	if got.filtered[0].ID != "86334621-8152-4e67-b322-9f139d6c0a57" {
		t.Errorf("unexpected filtered session ID: %s", got.filtered[0].ID)
	}
}

// TestUpdate_SearchEscClearsFilter verifies esc in search mode clears the filter.
func TestUpdate_SearchEscClearsFilter(t *testing.T) {
	m := modelWithSessions()
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	// Type to filter.
	for _, r := range "86334621" {
		m2, _ = m2.(model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	// Press esc to clear.
	m3, _ := m2.(model).Update(tea.KeyMsg{Type: tea.KeyEsc})
	got := m3.(model)
	if got.searchActive {
		t.Error("expected searchActive=false after esc")
	}
	if len(got.filtered) != 2 {
		t.Errorf("expected 2 sessions after clear, got %d", len(got.filtered))
	}
}

// TestUpdate_OpenDetail verifies enter opens the detail view for the cursor row.
func TestUpdate_OpenDetail(t *testing.T) {
	m := modelWithSessions()
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	got := m2.(model)
	if got.view != viewDetail {
		t.Errorf("expected viewDetail after enter, got %v", got.view)
	}
	if got.detailIdx != 0 {
		t.Errorf("expected detailIdx 0, got %d", got.detailIdx)
	}
}

// TestUpdate_DetailBackToList verifies esc from detail view returns to list.
func TestUpdate_DetailBackToList(t *testing.T) {
	m := modelWithSessions()
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m3, _ := m2.(model).Update(tea.KeyMsg{Type: tea.KeyEsc})
	if m3.(model).view != viewList {
		t.Errorf("expected viewList after esc from detail, got %v", m3.(model).view)
	}
}

// TestUpdate_OpenConfirmFromList verifies 'd' opens the confirm modal using cursor row.
func TestUpdate_OpenConfirmFromList(t *testing.T) {
	m := modelWithSessions()
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	got := m2.(model)
	if got.view != viewConfirm {
		t.Errorf("expected viewConfirm after 'd', got %v", got.view)
	}
	if len(got.deleteTargets) != 1 {
		t.Errorf("expected 1 target (cursor row), got %d", len(got.deleteTargets))
	}
	if got.deleteTargets[0].ID != m.filtered[0].ID {
		t.Errorf("unexpected target ID: %s", got.deleteTargets[0].ID)
	}
}

// TestUpdate_OpenConfirmFromList_MultiSelect verifies selected sessions become targets.
func TestUpdate_OpenConfirmFromList_MultiSelect(t *testing.T) {
	m := modelWithSessions()
	// Select all sessions.
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	// Open confirm.
	m3, _ := m2.(model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	got := m3.(model)
	if len(got.deleteTargets) != 2 {
		t.Errorf("expected 2 targets from multi-select, got %d", len(got.deleteTargets))
	}
}

// TestUpdate_OpenConfirmFromDetail verifies 'd' from detail opens confirm for that session.
func TestUpdate_OpenConfirmFromDetail(t *testing.T) {
	m := modelWithSessions()
	// Navigate to second session, then open detail.
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m3, _ := m2.(model).Update(tea.KeyMsg{Type: tea.KeyEnter})
	// Press 'd' from detail.
	m4, _ := m3.(model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	got := m4.(model)
	if got.view != viewConfirm {
		t.Errorf("expected viewConfirm, got %v", got.view)
	}
	if len(got.deleteTargets) != 1 {
		t.Errorf("expected 1 target, got %d", len(got.deleteTargets))
	}
	if got.deleteTargets[0].ID != m.filtered[1].ID {
		t.Errorf("unexpected target ID: %s", got.deleteTargets[0].ID)
	}
}

// TestUpdate_CancelConfirmWithN verifies 'n' closes confirm modal without side effects.
func TestUpdate_CancelConfirmWithN(t *testing.T) {
	m := modelWithSessions()
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	m3, _ := m2.(model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	got := m3.(model)
	if got.view != viewList {
		t.Errorf("expected viewList after 'n', got %v", got.view)
	}
	if len(got.deleteTargets) != 0 {
		t.Errorf("expected 0 targets after cancel, got %d", len(got.deleteTargets))
	}
}

// TestUpdate_CancelConfirmWithEsc verifies esc closes the confirm modal.
func TestUpdate_CancelConfirmWithEsc(t *testing.T) {
	m := modelWithSessions()
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	m3, _ := m2.(model).Update(tea.KeyMsg{Type: tea.KeyEsc})
	if m3.(model).view != viewList {
		t.Errorf("expected viewList after esc, got %v", m3.(model).view)
	}
}

// TestUpdate_Quit verifies 'q' returns a tea.Quit command.
func TestUpdate_Quit(t *testing.T) {
	m := modelWithSessions()
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Fatal("expected non-nil cmd after 'q'")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Errorf("expected tea.QuitMsg, got %T", cmd())
	}
}

// TestUpdate_CtrlCQuitsFromAnyView verifies ctrl+c always quits.
func TestUpdate_CtrlCQuitsFromAnyView(t *testing.T) {
	for _, view := range []viewState{viewList, viewDetail, viewConfirm} {
		m := modelWithSessions()
		m.view = view
		_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
		if cmd == nil {
			t.Fatalf("view %v: expected quit cmd on ctrl+c", view)
		}
		if _, ok := cmd().(tea.QuitMsg); !ok {
			t.Errorf("view %v: expected QuitMsg, got %T", view, cmd())
		}
	}
}

// TestView_NarrowTerminal verifies the narrow-terminal message is shown at width < 40.
func TestView_NarrowTerminal(t *testing.T) {
	m := initialModel("/fake/dir")
	m.width = 30
	m.height = 20
	out := m.View()
	if !strings.Contains(out, "too narrow") {
		t.Error("expected narrow terminal message for width < 40")
	}
}

// TestView_EmptyState verifies the empty state message is shown when no sessions exist.
func TestView_EmptyState(t *testing.T) {
	m := initialModel("/fake/dir")
	m.width = 100
	m.height = 30
	m.loading = false
	m.sessions = []session.Session{}
	m.filtered = []session.Session{}
	out := m.View()
	if !strings.Contains(out, "No sessions found") {
		t.Errorf("expected empty state message, got:\n%s", out)
	}
}

// TestView_LoadingState verifies the loading message is shown while loading.
func TestView_LoadingState(t *testing.T) {
	m := initialModel("/fake/dir")
	m.width = 100
	m.height = 30
	m.loading = true
	out := m.View()
	if !strings.Contains(out, "Loading") {
		t.Errorf("expected loading message, got:\n%s", out)
	}
}

// TestView_SessionList verifies session IDs appear in the list view.
func TestView_SessionList(t *testing.T) {
	m := modelWithSessions()
	m.listH = 10
	out := m.View()
	if !strings.Contains(out, "86334621") {
		t.Errorf("expected first session ID in list view, got:\n%s", out)
	}
	if !strings.Contains(out, "c0c723f4") {
		t.Errorf("expected second session ID in list view, got:\n%s", out)
	}
}

// TestTrunc verifies the trunc helper truncates and appends ellipsis correctly.
func TestTrunc(t *testing.T) {
	tests := []struct {
		s    string
		max  int
		want string
	}{
		{"hello", 10, "hello"},
		{"hello", 5, "hello"},
		{"hello", 4, "hel…"},
		{"hello", 1, "…"},
		{"hello", 0, ""},
		{"", 5, ""},
	}
	for _, tc := range tests {
		got := trunc(tc.s, tc.max)
		if got != tc.want {
			t.Errorf("trunc(%q, %d) = %q, want %q", tc.s, tc.max, got, tc.want)
		}
	}
}

// TestColMode verifies currentColMode returns the correct mode based on terminal width.
func TestColMode(t *testing.T) {
	tests := []struct {
		width int
		want  colMode
	}{
		{39, colsMini},
		{59, colsMini},
		{60, colsNoEvents},
		{79, colsNoEvents},
		{80, colsFull},
		{120, colsFull},
	}
	for _, tc := range tests {
		m := model{width: tc.width}
		got := m.currentColMode()
		if got != tc.want {
			t.Errorf("width %d: expected colMode %v, got %v", tc.width, tc.want, got)
		}
	}
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

// fixtureSessions returns two minimal sessions for testing.
func fixtureSessions() []session.Session {
	return []session.Session{
		{
			ID:         "86334621-8152-4e67-b322-9f139d6c0a57",
			Repository: "github/copilot-cli",
			EventCount: 150,
			SizeBytes:  2_100_000,
		},
		{
			ID:         "c0c723f4-08d2-4257-9b30-5d2fd728dc45",
			Repository: "my-project",
			EventCount: 42,
			SizeBytes:  512_000,
		},
	}
}

// modelWithSessions returns a ready-to-navigate model pre-loaded with two sessions.
func modelWithSessions() model {
	m := initialModel("/fake/dir")
	m.loading = false
	m.width = 120
	m.height = 40
	m.listH = 33
	sessions := fixtureSessions()
	m.sessions = sessions
	m.filtered = sessions
	return m
}
