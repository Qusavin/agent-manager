package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
)

func typeInto(t *testing.T, m *Model, text string) {
	t.Helper()
	for _, r := range text {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
		if r == '\n' {
			msg = tea.KeyMsg{Type: tea.KeyEnter}
		}
		updated, _ := m.handleKey(msg)
		*m = *updated.(*Model)
	}
}

// N on a group opens its note, esc writes it, and the row wears the mark
// in the same frame rather than waiting for the poller.
func TestNoteWritesOnTheGroupUnderTheCursor(t *testing.T) {
	m := buildModel(t)
	if err := m.store.CreateGroup("backend", t.TempDir()); err != nil {
		t.Fatalf("create group: %v", err)
	}
	m.applyCmd(t, m.refreshCmd())
	m.selectGroupRow(t, "backend")

	updated, _ := m.handleKey(runeKey("N"))
	m = updated.(*Model)
	if m.mode != modeNote {
		t.Fatalf("N left mode = %v", m.mode)
	}
	typeInto(t, m, "the queue is at-least-once")
	updated, _ = m.handleKey(namedKey(tea.KeyEsc))
	m = updated.(*Model)
	if m.mode != modeList {
		t.Fatalf("esc left mode = %v", m.mode)
	}

	stored, err := m.store.GroupNote("backend")
	if err != nil || stored != "the queue is at-least-once" {
		t.Fatalf("stored note = %q, %v", stored, err)
	}
	if m.groupNote("backend") != stored {
		t.Fatalf("the row map still reads %q", m.groupNote("backend"))
	}
	if view := ansi.Strip(m.viewListFrame()); !strings.Contains(view, noteMark) {
		t.Fatalf("the group row does not wear the note mark:\n%s", view)
	}
}

// Reopening shows what was written, and the caret sits at the end so an
// addition is typed rather than pushed in front of it.
func TestNoteReopensOnItsTextWithTheCaretAtTheEnd(t *testing.T) {
	m := buildModel(t)
	if err := m.store.CreateGroup("backend", t.TempDir()); err != nil {
		t.Fatalf("create group: %v", err)
	}
	if err := m.store.SetGroupNote("backend", "first"); err != nil {
		t.Fatalf("SetGroupNote: %v", err)
	}
	m.applyCmd(t, m.refreshCmd())
	m.selectGroupRow(t, "backend")

	updated, _ := m.handleKey(runeKey("N"))
	m = updated.(*Model)
	if got := m.note.input.Value(); got != "first" {
		t.Fatalf("the editor opened on %q", got)
	}
	typeInto(t, m, " then second")
	updated, _ = m.handleKey(namedKey(tea.KeyEsc))
	m = updated.(*Model)
	stored, _ := m.store.GroupNote("backend")
	if stored != "first then second" {
		t.Fatalf("stored note = %q", stored)
	}
}

// A session row borrows the note of the group it is filed under: the note
// describes that work either way.
func TestNoteOnASessionRowOpensItsGroupsNote(t *testing.T) {
	m := buildModel(t)
	dir := t.TempDir()
	if err := m.store.CreateGroup("backend", dir); err != nil {
		t.Fatalf("create group: %v", err)
	}
	m.applyCmd(t, m.refreshCmd())
	createSession(t, m, "note-sess", dir, "backend")
	m.selectSessionRow(t, "note-sess")

	updated, _ := m.handleKey(runeKey("N"))
	m = updated.(*Model)
	if m.mode != modeNote || m.note.group != "backend" {
		t.Fatalf("N on a session opened %q in mode %v", m.note.group, m.mode)
	}
}

// Root has no row of its own to keep a note on, so it says so instead of
// opening an editor whose text would have nowhere to go.
func TestNoteRefusesTheRoot(t *testing.T) {
	m := buildModel(t)
	createSessionOn(t, m, "rootless", "quietchat", t.TempDir())
	m.cursor = 0
	if entry, ok := m.selectedRow(); !ok || !entry.isRoot() {
		t.Skip("the first row is not root in this layout")
	}
	updated, _ := m.handleKey(runeKey("N"))
	m = updated.(*Model)
	if m.mode == modeNote {
		t.Fatal("root opened a note editor")
	}
	if m.errBar.text == "" {
		t.Fatal("refusing root said nothing")
	}
}

// Emptying a note clears the mark as well as the text.
func TestClearingANoteDropsItsMark(t *testing.T) {
	m := buildModel(t)
	if err := m.store.CreateGroup("backend", t.TempDir()); err != nil {
		t.Fatalf("create group: %v", err)
	}
	if err := m.store.SetGroupNote("backend", "temporary"); err != nil {
		t.Fatalf("SetGroupNote: %v", err)
	}
	m.applyCmd(t, m.refreshCmd())
	m.selectGroupRow(t, "backend")

	updated, _ := m.handleKey(runeKey("N"))
	m = updated.(*Model)
	m.note.input.SetValue("")
	updated, _ = m.handleKey(namedKey(tea.KeyEsc))
	m = updated.(*Model)

	if note, _ := m.store.GroupNote("backend"); note != "" {
		t.Fatalf("the cleared note still reads %q", note)
	}
	if m.groupNote("backend") != "" {
		t.Fatal("the row map kept the cleared note")
	}
	if view := ansi.Strip(m.viewListFrame()); strings.Contains(view, noteMark) {
		t.Fatalf("the mark outlived the note:\n%s", view)
	}
}

// The note takes the top of the right-hand column when a group is
// selected, which is where the standing context belongs.
func TestNoteShowsBesideTheGroupsAgents(t *testing.T) {
	m := buildModel(t)
	if err := m.store.CreateGroup("backend", t.TempDir()); err != nil {
		t.Fatalf("create group: %v", err)
	}
	if err := m.store.SetGroupNote("backend", "ship the auth rewrite"); err != nil {
		t.Fatalf("SetGroupNote: %v", err)
	}
	m.applyCmd(t, m.refreshCmd())
	m.selectGroupRow(t, "backend")

	view := ansi.Strip(m.viewListFrame())
	if !strings.Contains(view, "ship the auth rewrite") {
		t.Fatalf("the note is not beside the group:\n%s", view)
	}
}
