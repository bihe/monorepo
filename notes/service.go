// Package notes implements a service to store notes. Typically used for saving small reminders or some credentials
package notes

import (
	log "github.com/sirupsen/logrus"
)

// Service defines the available features for notes
type Service interface {
	Save(note Note, user string) (Note, error)
	Get(id string, user string) (Note, error)
	Delete(id string, user string) error
	Find(search string, user string) ([]Note, error)
}

// NewService creates a new instance of the Service
func NewService(logger *log.Entry) Service {
	return &notesService{
		logger: logger,
	}
}

// --------------------------------------------------------------------------

type notesService struct {
	logger *log.Entry
}

func (s *notesService) Save(note Note, user string) (Note, error) {
	return Note{}, nil
}

func (s *notesService) Get(id string, user string) (Note, error) {
	return Note{}, nil
}

func (s *notesService) Find(search string, user string) ([]Note, error) {
	return nil, nil
}

func (s *notesService) Delete(id string, user string) error {
	return nil
}
