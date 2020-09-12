package notes_test

import (
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.binggl.net/monorepo/notes"
)

var logger = log.New().WithField("mode", "test")

const user = "_test_User"

func TestServiceSave(t *testing.T) {
	s := notes.NewService(logger)

	_, err := s.Save(notes.Note{
		ID:      "ID",
		Created: time.Now().UTC(),
		Title:   "title",
		User:    user,
		Data:    "DATA",
		Tags:    []string{"a", "b"},
	}, user)

	if err != nil {
		t.Errorf("coul dnot save the notes, %v", err)
	}
}
