package store

import (
	"fmt"

	"github.com/google/uuid"
	"golang.binggl.net/monorepo/pkg/logging"
	"golang.binggl.net/monorepo/pkg/persistence"
)

// FileRepository offers CRUD functionality to store files
type FileRepository interface {
	Get(id string) (File, error)
	Save(item File) (File, error)
	Delete(item File) error
	InUnitOfWork(handle func(repo FileRepository) error) error
}

// CreateFileRepo creates a new repository
func CreateFileRepo(con persistence.Connection, logger logging.Logger) FileRepository {
	return &dbFileRepository{
		con:    con,
		logger: logger,
	}
}

// --------------------------------------------------------------------------
// Implementation
// --------------------------------------------------------------------------

type dbFileRepository struct {
	con    persistence.Connection
	logger logging.Logger
}

func (r *dbFileRepository) Get(id string) (File, error) {
	var file File
	h := r.con.R().Where(&File{ID: id}).First(&file)
	return file, h.Error
}

func (r *dbFileRepository) Save(item File) (File, error) {
	if len(item.Payload) == 0 {
		return File{}, fmt.Errorf("missing payload for file")
	}

	if len(item.Payload) != item.Size {
		return File{}, fmt.Errorf("the payload size is wrong")
	}
	if item.ID == "" {
		item.ID = uuid.New().String()
	}

	r.logger.Debug(fmt.Sprintf("store file: %s", item.ID))
	file, err := r.Get(item.ID)
	if err == nil {
		file.MimeType = item.MimeType
		file.Name = item.Name
		file.Size = item.Size
		file.Payload = item.Payload
		if h := r.con.W().Save(&file); h.Error != nil {
			return File{}, fmt.Errorf("could not save file: %v", h.Error)
		}
		return file, nil
	}
	file = File{
		ID:       item.ID,
		MimeType: item.MimeType,
		Name:     item.Name,
		Size:     item.Size,
		Payload:  item.Payload,
	}
	if h := r.con.W().Create(&file); h.Error != nil {
		return File{}, fmt.Errorf("could not save File: %v", h.Error)
	}
	return file, nil
}
func (r *dbFileRepository) Delete(item File) error {
	if item.ID == "" {
		return fmt.Errorf("missing id for file")
	}

	h := r.con.W().Delete(&item)
	if h.Error != nil {
		return fmt.Errorf("cannot delete file by id '%s': %v", item.ID, h.Error)
	}
	return nil
}
func (r *dbFileRepository) InUnitOfWork(handle func(repo FileRepository) error) error {
	return r.con.Begin(func(con persistence.Connection) error {
		repo := CreateFileRepo(con, r.logger)
		return handle(repo)
	})
}
