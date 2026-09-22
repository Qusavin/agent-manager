package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// noteMark rides after a group's name once it has a note, so the tree
// says where the standing context lives without opening anything.
const noteMark = "✎"

// noteEditorHeight is how many rows the editor gets: enough for a brief
// read at a glance, and short enough that the card still floats on a
// laptop terminal rather than filling it.
const noteEditorHeight = 14

type noteEditor struct {
	group string
	input textarea.Model
}

// openNote opens the note of the group under the cursor. A session row
// borrows its group's note rather than refusing: the note describes the
// work the session is filed under, which is what the cursor is on either
// way. The root holds no note, since it has no row of its own to keep one.
func (m *Model) openNote() {
	entry, ok := m.selectedRow()
	if !ok {
		return
	}
	group := entry.group
	if !entry.isGroup {
		group = entry.sess.Group
	}
	if group == "" {
		m.errBar.text = "notes live on a group; root is not one"
		return
	}
	note, err := m.store.GroupNote(group)
	if err != nil {
		m.errBar.text = err.Error()
		return
	}
	input := textarea.New()
	input.Placeholder = "what this group is for, and what has been decided"
	input.ShowLineNumbers = false
	input.CharLimit = 0
	input.SetHeight(noteEditorHeight)
	input.SetWidth(cardInnerWidth(m.noteCardWidth()))
	input.SetPromptFunc(2, func(int) string { return "  " })
	input.FocusedStyle.CursorLine = lipgloss.NewStyle()
	input.SetValue(note)
	// The caret lands at the end so an addition is typed rather than
	// pushed in front of what is already written.
	input.CursorEnd()
	input.Focus()
	m.note = noteEditor{group: group, input: input}
	m.mode = modeNote
	m.errBar.text = ""
}

// noteCardWidth gives the note a wider card than a dialog gets: a note is
// prose, and a 64-column box wraps it into a column too narrow to read.
func (m *Model) noteCardWidth() int {
	width := 92
	if m.width >= 28 && width > m.width-4 {
		width = m.width - 4
	}
	if fallback := m.cardWidth(); width < fallback {
		width = fallback
	}
	return width
}

// handleNoteKey types into the note. esc saves and closes, the way
// Settings does: a note is the user's own text, and there is nothing
// here an accidental keystroke could destroy that saving would not keep.
func (m *Model) handleNoteKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch keyName(msg) {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		return m.saveNote()
	}
	var cmd tea.Cmd
	m.note.input, cmd = m.note.input.Update(msg)
	return m, cmd
}

// saveNote writes the edited note and returns to the list. A write that
// fails keeps the editor open with the reason, so the text is never lost
// to a closing card.
func (m *Model) saveNote() (tea.Model, tea.Cmd) {
	note := strings.TrimRight(m.note.input.Value(), "\n \t")
	if err := m.store.SetGroupNote(m.note.group, note); err != nil {
		m.errBar.text = err.Error()
		return m, nil
	}
	// The map the rows read is updated here rather than waited for: the
	// mark belongs on the row in the same frame the card leaves.
	if m.groupNotes == nil {
		m.groupNotes = map[string]string{}
	}
	if note == "" {
		delete(m.groupNotes, m.note.group)
	} else {
		m.groupNotes[m.note.group] = note
	}
	m.note = noteEditor{}
	m.mode = modeList
	m.errBar.text = ""
	return m, nil
}

func (m *Model) viewNote() string {
	body := m.note.input.View()
	return m.cardSized(m.noteCardWidth(), "note · "+m.note.group, body, [][2]string{
		{"esc", "save and close"},
	})
}

// groupNote is the stored note for a group, empty for one that has none.
func (m *Model) groupNote(group string) string { return m.groupNotes[group] }

// noteLines renders a group's note for the right-hand column: a heading
// and as much of the note as the rows allow, with the rest counted rather
// than cut off silently.
func (m *Model) noteLines(group string, width, height int) []string {
	note := m.groupNote(group)
	if note == "" || height < 2 {
		return nil
	}
	lines := []string{subtleStyle.Render(noteMark + " note")}
	var wrapped []string
	for _, paragraph := range strings.Split(note, "\n") {
		if paragraph == "" {
			wrapped = append(wrapped, "")
			continue
		}
		wrapped = append(wrapped, strings.Split(ansi.Wordwrap(paragraph, max(width, 1), "-"), "\n")...)
	}
	room := height - 1
	if len(wrapped) > room {
		// The last row goes to the count, so a long note says how much of
		// it is out of sight instead of ending mid-sentence.
		shown := max(room-1, 0)
		lines = append(lines, renderNoteBody(wrapped[:shown])...)
		lines = append(lines, mutedStyle.Render(fmt.Sprintf("+%d more · N", len(wrapped)-shown)))
		return lines
	}
	return append(lines, renderNoteBody(wrapped)...)
}

func renderNoteBody(lines []string) []string {
	out := make([]string, len(lines))
	for i, line := range lines {
		out[i] = valueStyle.Render(line)
	}
	return out
}
