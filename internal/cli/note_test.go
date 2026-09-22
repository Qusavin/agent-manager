package cli

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/YoanWai/agent-manager/internal/sessioncmd"
)

type fakeNotes struct {
	note     sessioncmd.GroupNote
	appended string
	err      error
}

func (f *fakeNotes) Note(string) (sessioncmd.GroupNote, error) { return f.note, f.err }

func (f *fakeNotes) AppendNote(_ string, text string) (sessioncmd.GroupNote, error) {
	f.appended = text
	if f.err != nil {
		return sessioncmd.GroupNote{}, f.err
	}
	f.note.Note = strings.TrimSpace(f.note.Note + "\n\n" + text)
	return f.note, nil
}

func TestNoteReadPrintsTheGroupsNote(t *testing.T) {
	out := &bytes.Buffer{}
	notes := &fakeNotes{note: sessioncmd.GroupNote{Group: "backend", Note: "ship the auth rewrite"}}
	if err := runNoteRead(out, notes, nil, "cafe0001"); err != nil {
		t.Fatalf("note read: %v", err)
	}
	if !strings.Contains(out.String(), "ship the auth rewrite") {
		t.Fatalf("note read printed %q", out.String())
	}
}

// A sentence typed without quotes arrives as several operands, and the
// note is the whole sentence rather than its first word.
func TestNoteAppendJoinsItsOperands(t *testing.T) {
	out := &bytes.Buffer{}
	notes := &fakeNotes{note: sessioncmd.GroupNote{Group: "backend"}}
	if err := runNoteAppend(out, notes, []string{"retries", "are", "idempotent"}, "cafe0001"); err != nil {
		t.Fatalf("note append: %v", err)
	}
	if notes.appended != "retries are idempotent" {
		t.Fatalf("appended %q", notes.appended)
	}
	if !strings.Contains(out.String(), "retries are idempotent") {
		t.Fatalf("note append printed %q", out.String())
	}
}

func TestNoteAppendNeedsText(t *testing.T) {
	err := runNoteAppend(&bytes.Buffer{}, &fakeNotes{}, nil, "cafe0001")
	if err == nil || errors.Is(err, ErrUsageShown) {
		t.Fatalf("an empty append returned %v", err)
	}
}
