package store

import (
	"strings"
	"testing"
)

func TestGroupNoteRoundTrips(t *testing.T) {
	st := newTestStore(t)
	if err := st.AddGroup("backend", "/tmp", ""); err != nil {
		t.Fatalf("AddGroup: %v", err)
	}
	if err := st.SetGroupNote("backend", "ship the auth rewrite\n\n"); err != nil {
		t.Fatalf("SetGroupNote: %v", err)
	}
	note, err := st.GroupNote("backend")
	if err != nil {
		t.Fatalf("GroupNote: %v", err)
	}
	if note != "ship the auth rewrite" {
		t.Fatalf("note = %q, want the text without the editor's trailing blank lines", note)
	}
	groups, err := st.Groups()
	if err != nil {
		t.Fatalf("Groups: %v", err)
	}
	if len(groups) != 1 || groups[0].Note != "ship the auth rewrite" {
		t.Fatalf("the listed group did not carry its note: %+v", groups)
	}
}

// A group with no row of its own, the root included, has no note rather
// than an error: an agent may ask about a group name that was never
// registered.
func TestGroupNoteOfAnUnregisteredGroupIsEmpty(t *testing.T) {
	st := newTestStore(t)
	for _, name := range []string{"", "never-created"} {
		note, err := st.GroupNote(name)
		if err != nil || note != "" {
			t.Fatalf("GroupNote(%q) = %q, %v; want empty and no error", name, note, err)
		}
	}
	if err := st.SetGroupNote("", "anything"); err == nil {
		t.Fatal("the root took a note")
	}
	if err := st.SetGroupNote("never-created", "anything"); err == nil {
		t.Fatal("a group that does not exist took a note")
	}
}

// The note follows the group through a rename, which rewrites the row's
// name rather than making a new row.
func TestGroupNoteSurvivesARename(t *testing.T) {
	st := newTestStore(t)
	if err := st.AddGroup("api", "", ""); err != nil {
		t.Fatalf("AddGroup: %v", err)
	}
	if err := st.SetGroupNote("api", "keep the handlers thin"); err != nil {
		t.Fatalf("SetGroupNote: %v", err)
	}
	if err := st.RenameGroup("api", "backend"); err != nil {
		t.Fatalf("RenameGroup: %v", err)
	}
	note, err := st.GroupNote("backend")
	if err != nil {
		t.Fatalf("GroupNote: %v", err)
	}
	if note != "keep the handlers thin" {
		t.Fatalf("after the rename the note read %q", note)
	}
}

func TestAppendGroupNoteAddsAParagraphAndLeavesTheRest(t *testing.T) {
	st := newTestStore(t)
	if err := st.AddGroup("backend", "", ""); err != nil {
		t.Fatalf("AddGroup: %v", err)
	}
	// An append to an empty note opens it rather than leading with a gap.
	note, err := st.AppendGroupNote("backend", "\nthe queue is at-least-once\n")
	if err != nil {
		t.Fatalf("AppendGroupNote: %v", err)
	}
	if note != "the queue is at-least-once" {
		t.Fatalf("first append wrote %q", note)
	}
	note, err = st.AppendGroupNote("backend", "retries are idempotent by key")
	if err != nil {
		t.Fatalf("AppendGroupNote: %v", err)
	}
	if note != "the queue is at-least-once\n\nretries are idempotent by key" {
		t.Fatalf("second append wrote %q", note)
	}
	if _, err := st.AppendGroupNote("backend", "   \n "); err == nil {
		t.Fatal("an append of nothing but whitespace was written")
	}
	if _, err := st.AppendGroupNote("no-such-group", "text"); err == nil {
		t.Fatal("an append to a group that does not exist succeeded")
	}
}

// The cap refuses the write instead of storing a truncated note, so the
// caller knows the text did not land.
func TestGroupNoteRefusesAWriteOverTheCap(t *testing.T) {
	st := newTestStore(t)
	if err := st.AddGroup("backend", "", ""); err != nil {
		t.Fatalf("AddGroup: %v", err)
	}
	if err := st.SetGroupNote("backend", strings.Repeat("x", noteLimit+1)); err == nil {
		t.Fatal("a note past the cap was stored")
	}
	if err := st.SetGroupNote("backend", "short"); err != nil {
		t.Fatalf("SetGroupNote: %v", err)
	}
	if _, err := st.AppendGroupNote("backend", strings.Repeat("y", noteLimit)); err == nil {
		t.Fatal("an append past the cap was stored")
	}
	note, err := st.GroupNote("backend")
	if err != nil || note != "short" {
		t.Fatalf("the refused append changed the note: %q, %v", note, err)
	}
}
