package sessioncmd

import (
	"strings"
	"testing"

	"github.com/YoanWai/agent-manager/internal/status"
	"github.com/YoanWai/agent-manager/internal/store"
	"github.com/google/uuid"
)

func TestNoteReadsTheCallersGroupNote(t *testing.T) {
	h := newSessionHarness(t)
	if err := h.store.SetGroupNote("backend", "ship the auth rewrite"); err != nil {
		t.Fatalf("SetGroupNote: %v", err)
	}
	note, err := h.sessions.Note(h.caller.ID)
	if err != nil {
		t.Fatalf("Note: %v", err)
	}
	if note.Group != "backend" || note.Note != "ship the auth rewrite" {
		t.Fatalf("Note = %+v", note)
	}
	if formatted := FormatGroupNote(note); !strings.Contains(formatted, "ship the auth rewrite") {
		t.Fatalf("FormatGroupNote = %q", formatted)
	}
}

// A group with nothing written yet answers rather than failing: an agent
// reading before there is a note has learned something either way.
func TestNoteOnAnEmptyGroupSaysSo(t *testing.T) {
	h := newSessionHarness(t)
	note, err := h.sessions.Note(h.caller.ID)
	if err != nil {
		t.Fatalf("Note: %v", err)
	}
	if note.Note != "" || note.Group != "backend" {
		t.Fatalf("Note = %+v", note)
	}
	if formatted := FormatGroupNote(note); !strings.Contains(formatted, "no note yet") {
		t.Fatalf("FormatGroupNote = %q", formatted)
	}
}

// Appending adds to what is there, so what the user wrote in Agent Manager
// and what another session added both survive.
func TestAppendNoteKeepsWhatIsAlreadyWritten(t *testing.T) {
	h := newSessionHarness(t)
	if err := h.store.SetGroupNote("backend", "the user's own line"); err != nil {
		t.Fatalf("SetGroupNote: %v", err)
	}
	note, err := h.sessions.AppendNote(h.caller.ID, "retries are idempotent by key")
	if err != nil {
		t.Fatalf("AppendNote: %v", err)
	}
	if note.Note != "the user's own line\n\nretries are idempotent by key" {
		t.Fatalf("note after the append = %q", note.Note)
	}
	if _, err := h.sessions.AppendNote(h.caller.ID, "   "); err == nil {
		t.Fatal("an append of whitespace was accepted")
	}
}

// A session at the root has no group to carry a note: reading says so,
// and appending refuses rather than dropping the text.
func TestNoteAtTheRootHasNowhereToGo(t *testing.T) {
	h := newSessionHarness(t)
	rootID := uuid.NewString()[:8]
	if err := h.store.CreateSession(store.Session{
		ID: rootID, Name: "rootless", Tool: h.caller.Tool, Cwd: h.caller.Cwd, Status: status.Idle,
	}); err != nil {
		t.Fatalf("create session row: %v", err)
	}
	note, err := h.sessions.Note(rootID)
	if err != nil {
		t.Fatalf("Note: %v", err)
	}
	if note.Group != "" || note.Note != "" {
		t.Fatalf("a root session read %+v", note)
	}
	if formatted := FormatGroupNote(note); !strings.Contains(formatted, "root") {
		t.Fatalf("FormatGroupNote = %q", formatted)
	}
	if _, err := h.sessions.AppendNote(rootID, "somewhere"); err == nil {
		t.Fatal("a root session appended a note")
	}
}
