package store

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

// noteLimit caps one group note. A note is a brief, not a log: an agent
// appending to it every turn would otherwise grow a row nothing reads to
// the end. Writes past the cap are refused rather than truncated, so the
// caller learns the note is full instead of silently losing its tail.
const noteLimit = 16000

// GroupNote reads a group's note. A group with no row of its own - the
// root, or one that only exists as a session's group name - has no note
// rather than an error, which is what an agent asking about its own group
// gets when that group was never registered.
func (s *Store) GroupNote(name string) (string, error) {
	if name == "" {
		return "", nil
	}
	var note string
	err := s.db.QueryRow(`SELECT note FROM groups WHERE name = ?`, name).Scan(&note)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return note, nil
}

// SetGroupNote replaces a group's note. The text is stored as written, so
// what the editor showed is what comes back, minus the trailing blank
// lines an editor leaves behind.
func (s *Store) SetGroupNote(name, note string) error {
	if name == "" {
		return errors.New("the root group cannot hold a note")
	}
	note = strings.TrimRight(note, "\n \t")
	if len(note) > noteLimit {
		return fmt.Errorf("the note is %d characters; %s holds %d", len(note), name, noteLimit)
	}
	res, err := s.db.Exec(`UPDATE groups SET note = ? WHERE name = ?`, note, name)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("no group named %s", name)
	}
	return nil
}

// AppendGroupNote adds a paragraph to the end of a group's note and
// returns the note as it now stands. An agent writes this way rather than
// with SetGroupNote: the append is one statement, so two sessions writing
// at once both land, and neither can drop what the user wrote.
func (s *Store) AppendGroupNote(name, text string) (string, error) {
	if name == "" {
		return "", errors.New("the root group cannot hold a note")
	}
	text = strings.TrimRight(strings.TrimLeft(text, "\n"), "\n \t")
	if text == "" {
		return "", errors.New("the text to append is empty")
	}
	tx, err := s.db.Begin()
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	var note string
	err = tx.QueryRow(`SELECT note FROM groups WHERE name = ?`, name).Scan(&note)
	if errors.Is(err, sql.ErrNoRows) {
		return "", fmt.Errorf("no group named %s", name)
	}
	if err != nil {
		return "", err
	}
	if note != "" {
		note += "\n\n"
	}
	note += text
	if len(note) > noteLimit {
		return "", fmt.Errorf("appending would take the note past %d characters; edit it in Agent Manager first", noteLimit)
	}
	if _, err := tx.Exec(`UPDATE groups SET note = ? WHERE name = ?`, note, name); err != nil {
		return "", err
	}
	if err := tx.Commit(); err != nil {
		return "", err
	}
	return note, nil
}
