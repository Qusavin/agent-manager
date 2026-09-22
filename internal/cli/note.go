package cli

import (
	"io"
	"strings"

	"github.com/YoanWai/agent-manager/internal/sessioncmd"
)

const (
	usageNoteRead   = "note read [--json]"
	usageNoteAppend = "note append <text> [--json]"
)

type noteCommands interface {
	Note(sessionID string) (sessioncmd.GroupNote, error)
	AppendNote(sessionID, text string) (sessioncmd.GroupNote, error)
}

func newNotes(configDir string) noteCommands {
	return sessioncmd.NewSessions(configDir, sessioncmd.CLIVocabulary())
}

func noteVerbs() []command {
	return []command{
		{name: "read", usage: usageNoteRead, about: "read the standing note on the group this session is filed under: what the work is for, and what has been settled", run: bind(newNotes, runNoteRead)},
		{name: "append", usage: usageNoteAppend, about: "add a paragraph to that note, for whoever works in this group next; it never overwrites what is there", run: bind(newNotes, runNoteAppend)},
	}
}

func noteSection() section {
	return groupSection("Group note", "note", "the standing context on the group this session belongs to", noteVerbs())
}

func runNoteRead(out io.Writer, notes noteCommands, args []string, sessionID string) error {
	set := newFlagSet(usageNoteRead)
	asJSON := jsonFlag(set)
	if _, err := parseCommand(out, set, args, 0, 0); err != nil {
		return err
	}
	note, err := notes.Note(sessionID)
	if err != nil {
		return err
	}
	return emit(out, *asJSON, note, sessioncmd.FormatGroupNote(note))
}

// The text is taken as every operand joined, so a sentence typed without
// quotes lands whole rather than as its first word.
func runNoteAppend(out io.Writer, notes noteCommands, args []string, sessionID string) error {
	set := newFlagSet(usageNoteAppend)
	asJSON := jsonFlag(set)
	operands, err := parseCommand(out, set, args, 1, anyNumber)
	if err != nil {
		return err
	}
	note, err := notes.AppendNote(sessionID, strings.Join(operands, " "))
	if err != nil {
		return err
	}
	return emit(out, *asJSON, note, sessioncmd.FormatGroupNote(note))
}
