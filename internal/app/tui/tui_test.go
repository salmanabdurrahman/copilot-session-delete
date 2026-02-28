// Tests for the Bubble Tea TUI model.
// Uses internal package access so unexported messages and view states are reachable.
package tui

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/salmanabdurrahman/copilot-session-delete/internal/core/deletion"
	"github.com/salmanabdurrahman/copilot-session-delete/internal/core/session"
)

// TestInitialModel verifies the initial model has sensible default values.
func TestInitialModel(t *testing.T) {
	m := initialModel("/fake/dir", false)
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
	m := initialModel("/fake/dir", false)
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
	m := initialModel("/fake/dir", false)
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
	m := initialModel("/fake/dir", false)
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
	m := initialModel("/fake/dir", false)
	m.width = 30
	m.height = 20
	out := m.View()
	if !strings.Contains(out, "too narrow") {
		t.Error("expected narrow terminal message for width < 40")
	}
}

// TestView_EmptyState verifies the empty state message is shown when no sessions exist.
func TestView_EmptyState(t *testing.T) {
	m := initialModel("/fake/dir", false)
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
	m := initialModel("/fake/dir", false)
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

// ─── Deletion wiring tests ──────────────────────────────────────────

// TestUpdate_ConfirmYStartsDeletion verifies pressing 'y' in confirm switches to
// the list view, sets deleting=true, and returns a non-nil async command.
func TestUpdate_ConfirmYStartsDeletion(t *testing.T) {
	m := modelWithSessions()
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	got, cmd := m2.(model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	result := got.(model)

	if !result.deleting {
		t.Error("expected deleting=true after confirm 'y'")
	}
	if result.view != viewList {
		t.Errorf("expected view to return to viewList after confirm 'y', got %v", result.view)
	}
	if cmd == nil {
		t.Error("expected non-nil delete cmd after confirm 'y'")
	}
}

// TestUpdate_DeleteComplete_AllSuccess verifies that on full success all deleted sessions
// are removed from the list and a success status message is shown.
func TestUpdate_DeleteComplete_AllSuccess(t *testing.T) {
	m := modelWithSessions()
	sessions := m.sessions

	m2, _ := m.Update(deleteCompleteMsg{
		results: []deletion.Result{
			{SessionID: sessions[0].ID, Success: true},
			{SessionID: sessions[1].ID, Success: true},
		},
	})
	got := m2.(model)

	if got.deleting {
		t.Error("expected deleting=false after deleteCompleteMsg")
	}
	if len(got.sessions) != 0 {
		t.Errorf("expected 0 remaining sessions after full success, got %d", len(got.sessions))
	}
	if got.statusIsErr {
		t.Error("expected no error status after full success")
	}
	if !strings.Contains(got.statusMsg, "2 session(s) deleted") {
		t.Errorf("unexpected statusMsg: %q", got.statusMsg)
	}
}

// TestUpdate_DeleteComplete_AllFailed verifies that on full failure sessions remain
// in the list and an error status message is shown.
func TestUpdate_DeleteComplete_AllFailed(t *testing.T) {
	m := modelWithSessions()
	sessions := m.sessions

	m2, _ := m.Update(deleteCompleteMsg{
		results: []deletion.Result{
			{SessionID: sessions[0].ID, Success: false, Err: fmt.Errorf("permission denied")},
			{SessionID: sessions[1].ID, Success: false, Err: fmt.Errorf("permission denied")},
		},
	})
	got := m2.(model)

	if len(got.sessions) != 2 {
		t.Errorf("expected sessions to remain on full failure, got %d", len(got.sessions))
	}
	if !got.statusIsErr {
		t.Error("expected statusIsErr=true after full failure")
	}
	if !strings.Contains(got.statusMsg, "All 2 deletion(s) failed") {
		t.Errorf("unexpected statusMsg: %q", got.statusMsg)
	}
}

// TestUpdate_DeleteComplete_PartialFailed verifies that on partial failure only
// successfully deleted sessions are removed, with a warning status message.
func TestUpdate_DeleteComplete_PartialFailed(t *testing.T) {
	m := modelWithSessions()
	sessions := m.sessions

	m2, _ := m.Update(deleteCompleteMsg{
		results: []deletion.Result{
			{SessionID: sessions[0].ID, Success: true},
			{SessionID: sessions[1].ID, Success: false, Err: fmt.Errorf("permission denied")},
		},
	})
	got := m2.(model)

	if len(got.sessions) != 1 {
		t.Errorf("expected 1 remaining session after partial failure, got %d", len(got.sessions))
	}
	if got.sessions[0].ID != sessions[1].ID {
		t.Errorf("expected failing session to remain, got %s", got.sessions[0].ID)
	}
	if !got.statusIsErr {
		t.Error("expected statusIsErr=true after partial failure")
	}
	if !strings.Contains(got.statusMsg, "1 deleted") || !strings.Contains(got.statusMsg, "1 failed") {
		t.Errorf("unexpected statusMsg: %q", got.statusMsg)
	}
}

// TestUpdate_DeleteComplete_PlannerError verifies that a planner-level error in
// deleteCompleteMsg is surfaced as an error status message.
func TestUpdate_DeleteComplete_PlannerError(t *testing.T) {
	m := modelWithSessions()
	m.deleting = true

	m2, _ := m.Update(deleteCompleteMsg{err: fmt.Errorf("safety check failed")})
	got := m2.(model)

	if got.deleting {
		t.Error("expected deleting=false after error")
	}
	if !got.statusIsErr {
		t.Error("expected statusIsErr=true after planner error")
	}
	if !strings.Contains(got.statusMsg, "safety check failed") {
		t.Errorf("unexpected statusMsg: %q", got.statusMsg)
	}
}

// TestUpdate_DeleteComplete_DryRun verifies that dry-run results do not remove
// sessions from the list and show a [DRY-RUN] status message.
func TestUpdate_DeleteComplete_DryRun(t *testing.T) {
	m := modelWithSessions()
	sessions := m.sessions

	m2, _ := m.Update(deleteCompleteMsg{
		results: []deletion.Result{
			{SessionID: sessions[0].ID, Success: true, DryRun: true},
			{SessionID: sessions[1].ID, Success: true, DryRun: true},
		},
	})
	got := m2.(model)

	if len(got.sessions) != 2 {
		t.Errorf("expected sessions to remain after dry-run, got %d", len(got.sessions))
	}
	if got.statusIsErr {
		t.Error("expected no error status after dry-run")
	}
	if !strings.Contains(got.statusMsg, "[DRY-RUN]") || !strings.Contains(got.statusMsg, "Would delete 2") {
		t.Errorf("unexpected dry-run statusMsg: %q", got.statusMsg)
	}
}

// TestUpdate_DryRunConfirmY verifies pressing 'y' in dry-run mode returns a
// non-nil cmd (the executor is called with dryRun=true).
func TestUpdate_DryRunConfirmY(t *testing.T) {
	m := modelWithSessions()
	m.dryRun = true
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	got, cmd := m2.(model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	result := got.(model)

	if !result.deleting {
		t.Error("expected deleting=true after dry-run confirm 'y'")
	}
	if cmd == nil {
		t.Error("expected non-nil cmd for dry-run deletion")
	}
}

// TestUpdate_InputBlockedWhileDeleting verifies all key input is ignored when
// a deletion is in progress.
func TestUpdate_InputBlockedWhileDeleting(t *testing.T) {
	m := modelWithSessions()
	m.deleting = true
	m.cursor = 0

	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if m2.(model).cursor != 0 {
		t.Error("expected cursor to not move while deleting")
	}
}

// TestView_DeletingIndicator verifies the "Deleting…" indicator appears in the
// list view while a deletion is in progress.
func TestView_DeletingIndicator(t *testing.T) {
	m := modelWithSessions()
	m.deleting = true
	out := m.View()
	if !strings.Contains(out, "Deleting") {
		t.Errorf("expected deleting indicator in view, got:\n%s", out)
	}
}

// TestView_DryRunFooter verifies the footer shows [DRY-RUN] when dry-run is active.
func TestView_DryRunFooter(t *testing.T) {
	m := modelWithSessions()
	m.dryRun = true
	out := m.View()
	if !strings.Contains(out, "[DRY-RUN]") {
		t.Errorf("expected [DRY-RUN] in footer, got:\n%s", out)
	}
}

// TestView_ConfirmModal_DryRun verifies the confirm modal renders dry-run specific
// title and button text.
func TestView_ConfirmModal_DryRun(t *testing.T) {
	m := modelWithSessions()
	m.dryRun = true
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	got := m2.(model)
	out := got.View()
	if !strings.Contains(out, "[DRY-RUN]") {
		t.Errorf("expected [DRY-RUN] badge in confirm modal, got:\n%s", out)
	}
	if !strings.Contains(out, "Preview (no files removed)") {
		t.Errorf("expected dry-run button text in confirm modal, got:\n%s", out)
	}
}

// TestView_DetailView verifies the detail panel renders session fields correctly.
func TestView_DetailView(t *testing.T) {
	m := modelWithSessions()
	m.view = viewDetail
	m.detailIdx = 0
	out := m.View()
	if !strings.Contains(out, "Session Detail") {
		t.Errorf("expected 'Session Detail' header in detail view, got:\n%s", out)
	}
	if !strings.Contains(out, "86334621") {
		t.Errorf("expected first session ID in detail view, got:\n%s", out)
	}
}

// TestView_DetailView_OutOfBounds verifies an out-of-bounds detailIdx falls back
// to the list view.
func TestView_DetailView_OutOfBounds(t *testing.T) {
	m := modelWithSessions()
	m.view = viewDetail
	m.detailIdx = 999 // beyond filtered length
	out := m.View()
	// Should fall back to list view (header present, no "Session Detail").
	if strings.Contains(out, "Session Detail") {
		t.Error("expected fallback to list view for out-of-bounds detailIdx")
	}
}

// TestUpdate_VimKeys verifies 'k' and 'j' navigate the cursor up and down.
func TestUpdate_VimKeys(t *testing.T) {
	m := modelWithSessions()

	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if m2.(model).cursor != 1 {
		t.Errorf("expected cursor 1 after 'j', got %d", m2.(model).cursor)
	}

	m3, _ := m2.(model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if m3.(model).cursor != 0 {
		t.Errorf("expected cursor 0 after 'k', got %d", m3.(model).cursor)
	}
}

// TestUpdate_GotoFirst verifies 'g' jumps the cursor to the first session.
func TestUpdate_GotoFirst(t *testing.T) {
	m := modelWithSessions()
	m.cursor = 1

	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}})
	if m2.(model).cursor != 0 {
		t.Errorf("expected cursor 0 after 'g', got %d", m2.(model).cursor)
	}
}

// TestUpdate_GotoLast verifies 'G' jumps the cursor to the last session.
func TestUpdate_GotoLast(t *testing.T) {
	m := modelWithSessions()
	m.cursor = 0

	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'G'}})
	got := m2.(model)
	want := len(got.filtered) - 1
	if got.cursor != want {
		t.Errorf("expected cursor %d after 'G', got %d", want, got.cursor)
	}
}

// TestUpdate_RefreshKey verifies 'r' resets the model to loading state and
// returns a non-nil loadSessionsCmd.
func TestUpdate_RefreshKey(t *testing.T) {
	m := modelWithSessions()
	m.statusMsg = "some previous status"

	m2, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	got := m2.(model)

	if !got.loading {
		t.Error("expected loading=true after 'r'")
	}
	if got.sessions != nil {
		t.Error("expected sessions cleared after 'r'")
	}
	if got.statusMsg != "" {
		t.Errorf("expected statusMsg cleared after 'r', got %q", got.statusMsg)
	}
	if cmd == nil {
		t.Error("expected non-nil loadSessionsCmd after 'r'")
	}
}

// TestUpdate_SearchEnterDismisses verifies pressing 'enter' while search is active
// commits the query and deactivates the search input.
func TestUpdate_SearchEnterDismisses(t *testing.T) {
	m := modelWithSessions()
	// Activate search.
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	// Press enter to commit.
	m3, _ := m2.(model).Update(tea.KeyMsg{Type: tea.KeyEnter})
	got := m3.(model)
	if got.searchActive {
		t.Error("expected searchActive=false after enter in search mode")
	}
}

// TestView_LoadError verifies the error message is shown when a load error occurred.
func TestView_LoadError(t *testing.T) {
	m := initialModel("/fake/dir", false)
	m.width = 100
	m.height = 30
	m.loading = false
	m.loadErr = fmt.Errorf("permission denied")
	out := m.View()
	if !strings.Contains(out, "permission denied") {
		t.Errorf("expected load error message in view, got:\n%s", out)
	}
}

// TestView_FilteredEmpty verifies the "no match" message is shown when the filter
// matches no sessions.
func TestView_FilteredEmpty(t *testing.T) {
	m := modelWithSessions()
	m.filtered = nil // filter produced no results
	out := m.View()
	if !strings.Contains(out, "No sessions match") {
		t.Errorf("expected no-match message in view, got:\n%s", out)
	}
}

// TestView_ColHeaders_NoEvents verifies the EVENTS column is hidden at medium width.
func TestView_ColHeaders_NoEvents(t *testing.T) {
	m := modelWithSessions()
	m.width = 70 // colsNoEvents range
	out := m.View()
	if strings.Contains(out, "EVENTS") {
		t.Error("expected EVENTS column to be hidden at width 70")
	}
}

// TestView_ColHeaders_Mini verifies only ID and time columns appear at narrow width.
func TestView_ColHeaders_Mini(t *testing.T) {
	m := modelWithSessions()
	m.width = 50 // colsMini range
	out := m.View()
	if strings.Contains(out, "CWD/REPO") {
		t.Error("expected CWD/REPO column to be hidden at width 50")
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
	m := initialModel("/fake/dir", false)
	m.loading = false
	m.width = 120
	m.height = 40
	m.listH = 33
	sessions := fixtureSessions()
	m.sessions = sessions
	m.filtered = sessions
	return m
}
