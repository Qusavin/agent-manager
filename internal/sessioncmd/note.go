package sessioncmd

import (
	"fmt"
	"strings"

	"github.com/YoanWai/agent-manager/internal/store"
)

// GroupNote is a group's standing context: what the work filed under it is
// for, and what has been settled about it. It is written for whoever comes
// next, the user reading it in Agent Manager and the agents spawned into
// the group alike.
type GroupNote struct {
	Group string `json:"group" jsonschema:"the group the note belongs to"`
	Note  string `json:"note" jsonschema:"the note as it now stands, empty when the group has none"`
}

// Note reads the note of the calling session's group. A session filed at
// the root has no group to carry one, which comes back as an empty note
// rather than an error: an agent asking is not doing anything wrong.
func (s *Sessions) Note(sessionID string) (GroupNote, error) {
	runtime, err := s.open()
	if err != nil {
		return GroupNote{}, err
	}
	defer runtime.store.Close()
	sess, err := runtime.caller(sessionID)
	if err != nil {
		return GroupNote{}, err
	}
	return runtime.groupNote(sess)
}

// AppendNote adds a paragraph to the end of the group's note and returns
// the note as it now stands. Appending rather than replacing is what keeps
// several sessions in one group, and the user typing in the editor, from
// writing over each other.
func (s *Sessions) AppendNote(sessionID, text string) (GroupNote, error) {
	if strings.TrimSpace(text) == "" {
		return GroupNote{}, fmt.Errorf("note is empty; pass the text to add")
	}
	runtime, err := s.open()
	if err != nil {
		return GroupNote{}, err
	}
	defer runtime.store.Close()
	sess, err := runtime.caller(sessionID)
	if err != nil {
		return GroupNote{}, err
	}
	if sess.Group == "" {
		return GroupNote{}, fmt.Errorf("this session is filed at the root, which holds no note; move it into a group first")
	}
	note, err := runtime.store.AppendGroupNote(sess.Group, text)
	if err != nil {
		return GroupNote{}, err
	}
	return GroupNote{Group: sess.Group, Note: note}, nil
}

func (r *runtime) groupNote(sess store.Session) (GroupNote, error) {
	if sess.Group == "" {
		return GroupNote{}, nil
	}
	note, err := r.store.GroupNote(sess.Group)
	if err != nil {
		return GroupNote{}, err
	}
	return GroupNote{Group: sess.Group, Note: note}, nil
}

// FormatGroupNote is what the agent reads back: the note under the group
// it belongs to, or a line saying there is none, which is an answer rather
// than an empty result to retry.
func FormatGroupNote(note GroupNote) string {
	if note.Group == "" {
		return "this session is filed at the root, which holds no note"
	}
	if note.Note == "" {
		return "group " + note.Group + " has no note yet"
	}
	return "group " + note.Group + ":\n" + note.Note
}
