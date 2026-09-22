package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
)

// raiseJump opens the overlay the way the key does and paints one frame,
// which is where the labels are handed out.
func raiseJump(t *testing.T, m *Model) *Model {
	t.Helper()
	updated, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(";")})
	m = updated.(*Model)
	if !m.jump.active {
		t.Fatal("the jump key did not raise the overlay")
	}
	m.View()
	return m
}

// The labels walk the alphabet down the painted rows, so the first row on
// screen answers to the first key under the fingers.
func TestJumpLabelsRunDownThePaintedRows(t *testing.T) {
	m := raiseJump(t, shotModel())
	if len(m.jump.rows) == 0 {
		t.Fatal("painting the frame handed out no labels")
	}
	for i, label := range jumpAlphabet[:3] {
		row, ok := m.jump.rows[label]
		if !ok {
			t.Fatalf("label %q was never handed out", string(label))
		}
		if row != m.railTop+i {
			t.Fatalf("label %q lands on row %d, want %d", string(label), row, m.railTop+i)
		}
	}
	// The label is drawn in place of the row's status dot, so it is on
	// screen for the key it asks for.
	if view := ansi.Strip(m.View()); !strings.Contains(view, string(jumpAlphabet[0])+" ") {
		t.Fatal("the first label was never painted")
	}
}

// A capital label parks the cursor rather than opening the row, which is
// what makes a jump usable as the step before any other key.
func TestJumpCapitalOnlyMovesTheCursor(t *testing.T) {
	m := raiseJump(t, shotModel())
	target, ok := m.jump.rows['d']
	if !ok {
		t.Fatal("no row wears the d label")
	}
	updated, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("D")})
	m = updated.(*Model)
	if m.jump.active {
		t.Fatal("the overlay stayed up after a label was read")
	}
	if m.cursor != target {
		t.Fatalf("D left the cursor at %d, want %d", m.cursor, target)
	}
	if m.mode != modeList {
		t.Fatalf("a capital label opened the row; mode = %v", m.mode)
	}
}

// Anything that is not a label drops the overlay and leaves the list as it
// was: a mistyped jump must not act.
func TestJumpCancelsOnAKeyThatIsNoLabel(t *testing.T) {
	m := raiseJump(t, shotModel())
	before := m.cursor
	updated, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(*Model)
	if m.jump.active {
		t.Fatal("esc left the overlay up")
	}
	if m.cursor != before {
		t.Fatalf("esc moved the cursor to %d, want %d", m.cursor, before)
	}

	m = raiseJump(t, shotModel())
	before = m.cursor
	updated, _ = m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("2")})
	m = updated.(*Model)
	if m.jump.active || m.cursor != before {
		t.Fatalf("an unlabelled key acted: active = %v, cursor = %d, want %d", m.jump.active, m.cursor, before)
	}
}

// The overlay reads a label typed on a Russian layout too: ф is the a key.
func TestJumpReadsALabelOnARussianLayout(t *testing.T) {
	m := raiseJump(t, shotModel())
	target, ok := m.jump.rows['a']
	if !ok {
		t.Fatal("no row wears the a label")
	}
	updated, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("Ф")})
	m = updated.(*Model)
	if m.cursor != target {
		t.Fatalf("Ф left the cursor at %d, want %d", m.cursor, target)
	}
}

// A lowercase label opens the row it lands on, the same door enter uses.
func TestJumpOpensTheRowItLandsOn(t *testing.T) {
	m := buildModel(t)
	dir := t.TempDir()
	createSessionOn(t, m, "jump-one", "quietchat", dir)
	createSessionOn(t, m, "jump-two", "quietchat", dir)
	m.focusOnEnter = true
	m.cursor = 0
	m = raiseJump(t, m)

	label, ok := m.jump.labels[len(m.rows)-1]
	if !ok {
		t.Fatal("the last row was never labelled")
	}
	updated, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{label}})
	m = updated.(*Model)
	if m.cursor != len(m.rows)-1 {
		t.Fatalf("the label left the cursor at %d, want %d", m.cursor, len(m.rows)-1)
	}
	if m.mode != modeFocus {
		t.Fatalf("the label did not open the row; mode = %v", m.mode)
	}
}
