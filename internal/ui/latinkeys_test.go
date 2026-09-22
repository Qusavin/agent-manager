package ui

import (
	"strings"
	"testing"

	"github.com/YoanWai/agent-manager/internal/keybind"
	"github.com/YoanWai/agent-manager/internal/tmux"
	tea "github.com/charmbracelet/bubbletea"
)

func russianRune(r rune) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
}

// The list keys answer to the physical button, so a keyboard left on the
// Russian layout still drives the manager: "о" is the j key.
func TestListKeysAnswerOnARussianLayout(t *testing.T) {
	m := buildModel(t)
	dir := t.TempDir()
	createSessionOn(t, m, "latin-one", "quietchat", dir)
	createSessionOn(t, m, "latin-two", "quietchat", dir)
	m.cursor = 0

	updated, _ := m.handleKey(russianRune('о'))
	m = updated.(*Model)
	if m.cursor != 1 {
		t.Fatalf("о (the j key) left the cursor at %d, want 1", m.cursor)
	}
	updated, _ = m.handleKey(russianRune('л'))
	m = updated.(*Model)
	if m.cursor != 0 {
		t.Fatalf("л (the k key) left the cursor at %d, want 0", m.cursor)
	}
}

// A session key bound to alt+h detaches when alt+р arrives, which is what
// the same physical chord reports on a Russian layout.
func TestFocusDetachesOnARussianAltChord(t *testing.T) {
	m := buildModel(t)
	createSessionOn(t, m, "latin-detach", "quietchat", t.TempDir())
	altH, err := keybind.Parse("alt+h")
	if err != nil {
		t.Fatalf("parsing alt+h: %v", err)
	}
	m.keys = m.keys.With(keybind.Detach, keybind.Keys(altH))

	updated, _ := m.focusSelected()
	m = updated.(*Model)
	if m.mode != modeFocus {
		t.Fatalf("focus left mode = %v", m.mode)
	}
	updated, _ = m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("р"), Alt: true})
	m = updated.(*Model)
	if m.mode != modeList {
		t.Fatalf("alt+р did not detach; mode = %v", m.mode)
	}
}

// Folding the layout is for bindings only: a Russian character typed at a
// focused agent reaches its pane as the character it is.
func TestFocusForwardsCyrillicTextUnchanged(t *testing.T) {
	command, ok := focusKeyCommand(tmux.PaneTarget("x"), russianRune('р'))
	if !ok {
		t.Fatal("a cyrillic rune was dropped instead of forwarded")
	}
	if !strings.HasSuffix(command, "-H d1 80") {
		t.Fatalf("р went to the pane as %q, want its own utf-8 bytes d1 80", command)
	}
}
