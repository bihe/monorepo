package store

import (
	"fmt"
	"time"

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
	// this simplifies live-migrations by telling gorm to create missing tables
	con.W().AutoMigrate(&FileObject{}, &File{}, &Bookmark{})
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
	h := r.con.R().Joins("FileObject").Where(&File{ID: id}).First(&file)
	return file, h.Error
}

func (r *dbFileRepository) Save(item File) (File, error) {
	if item.FileObject == nil || item.FileObject.Payload == nil {
		return File{}, fmt.Errorf("missing payload for file")
	}

	if len(item.FileObject.Payload) != item.Size {
		return File{}, fmt.Errorf("the payload size is wrong")
	}
	if item.ID == "" {
		item.ID = uuid.New().String()
	}

	file, err := r.Get(item.ID)
	if err == nil {
		var fileObject FileObject
		if h := r.con.R().Where(&FileObject{ID: item.ID}).First(&fileObject); h.Error != nil {
			return File{}, fmt.Errorf("no FileObject available with given id: %v", h.Error)
		}

		// update the payload as a FileObject
		fileObject.Payload = file.FileObject.Payload
		if h := r.con.W().Save(&fileObject); h.Error != nil {
			return File{}, fmt.Errorf("could not save FileObject: %v", h.Error)
		}

		file.MimeType = item.MimeType
		file.Name = item.Name
		file.Size = item.Size
		file.FileObjectID = &item.ID
		file.Modified = time.Now()
		if h := r.con.W().Save(&file); h.Error != nil {
			return File{}, fmt.Errorf("could not save File: %v", h.Error)
		}
		return file, nil
	}

	// first create a new FileObject
	fileObject := FileObject{
		ID:      item.ID,
		Payload: item.FileObject.Payload,
	}
	if h := r.con.W().Create(&fileObject); h.Error != nil {
		return File{}, fmt.Errorf("could not save FileObject: %v", h.Error)
	}
	// create a new File which references the FileObject
	file = File{
		ID:           item.ID,
		MimeType:     item.MimeType,
		Name:         item.Name,
		Size:         item.Size,
		Modified:     time.Now(),
		FileObjectID: &fileObject.ID,
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

	fileObject := FileObject{
		ID: item.ID,
	}
	h := r.con.W().Delete(&fileObject)
	if h.Error != nil {
		return fmt.Errorf("cannot delete FileObject by id '%s': %v", item.ID, h.Error)
	}

	h = r.con.W().Delete(&item)
	if h.Error != nil {
		return fmt.Errorf("cannot delete File by id '%s': %v", item.ID, h.Error)
	}
	return nil
}

func (r *dbFileRepository) InUnitOfWork(handle func(repo FileRepository) error) error {
	return r.con.Begin(func(con persistence.Connection) error {
		repo := CreateFileRepo(con, r.logger)
		return handle(repo)
	})
}
