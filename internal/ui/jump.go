package ui

import (
	"unicode"

	tea "github.com/charmbracelet/bubbletea"
)

// jumpAlphabet is the order labels are handed out in, written for ten
// fingers: the home row first, from the strongest finger outwards, then
// the row above and the row below it. A jump is meant to be one reflex,
// so the keys it can land on are the ones already under the hands.
//
// A label is a single key. Rows past the alphabet go unlabelled rather
// than take a two-key label: a second keystroke on every jump is a poor
// trade for reaching a row the screen is already scrolling towards.
var jumpAlphabet = []rune("fjdksla;ghrueiwoqpvmc,x.")

// jumpState is the label overlay. While it is up every row the rail paints
// wears a key where its status dot goes, and that key is the whole gesture.
type jumpState struct {
	active bool
	// rows maps a label to the m.rows index it lands on, and labels maps
	// that index back to the key drawn on it. Both are filled at paint
	// time, the way railHits is, so a label can only ever name a row the
	// frame actually showed.
	rows   map[rune]int
	labels map[int]rune
}

// openJump raises the overlay. The labels themselves are not handed out
// here: which rows are on screen is the rail's answer, given while it
// paints the frame this key asked for.
func (m *Model) openJump() {
	if len(m.rows) == 0 {
		return
	}
	m.jump = jumpState{active: true, rows: map[rune]int{}, labels: map[int]rune{}}
}

func (m *Model) closeJump() { m.jump = jumpState{} }

// labelJumpRows gives the entries between start and end a key each, in the
// order they are painted, so the labels read down the screen.
func (m *Model) labelJumpRows(start, end, offset int) {
	if !m.jump.active {
		return
	}
	m.jump.rows = make(map[rune]int, end-start)
	m.jump.labels = make(map[int]rune, end-start)
	for i, next := start, 0; i < end && next < len(jumpAlphabet); i, next = i+1, next+1 {
		label := jumpAlphabet[next]
		m.jump.rows[label] = offset + i
		m.jump.labels[offset+i] = label
	}
}

// jumpMark is the label drawn in place of a row's status dot, empty for a
// row the overlay is not on or could not reach.
func (m *Model) jumpMark(index int) string {
	if !m.jump.active {
		return ""
	}
	label, ok := m.jump.labels[index]
	if !ok {
		return ""
	}
	return jumpLabelStyle.Render(string(label))
}

// handleJumpKey reads the label and parks the cursor on its row. It only
// selects: opening the row is the next key's business, so a jump can be
// the step before any of them rather than a door that commits to one.
// A shifted label still reads, and anything that is not a label closes the
// overlay without acting: a mistyped jump should leave the list as it was.
func (m *Model) handleJumpKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := keyName(msg)
	if key == "ctrl+c" {
		return m, tea.Quit
	}
	typed := []rune(key)
	if len(typed) != 1 {
		m.closeJump()
		return m, nil
	}
	label := unicode.ToLower(typed[0])
	index, found := m.jump.rows[label]
	m.closeJump()
	if !found {
		return m, nil
	}
	return m, m.selectRow(index)
}
