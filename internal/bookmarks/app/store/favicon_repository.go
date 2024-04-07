package store

import (
	"fmt"
	"time"

	"golang.binggl.net/monorepo/pkg/logging"
	"golang.binggl.net/monorepo/pkg/persistence"
	"gorm.io/gorm"
)

// FaviconRepository offers CRUD functionality to store favicons
type FaviconRepository interface {
	persistence.DBRepository
	Get(id string) (Favicon, error)
	Save(item Favicon) (Favicon, error)
	Delete(item Favicon) error
}

// CreateFaviconRepoRW creates a new repository with read and write connection
func CreateFaviconRepoRW(read, write *gorm.DB, logger logging.Logger) FaviconRepository {
	return &dbFaviconRepository{
		GormRepository: persistence.GormRepository{
			Read:  read,
			Write: write,
		},
		logger: logger,
	}
}

// CreateFaviconRepoUnit creates a new repository with a transactional context
func CreateFaviconRepoUnit(u persistence.Unit, logger logging.Logger) FaviconRepository {
	return &dbFaviconRepository{
		GormRepository: persistence.GormRepository{
			Tx: u.Tx,
		},
		logger: logger,
	}
}

// --------------------------------------------------------------------------
// Implementation
// --------------------------------------------------------------------------

type dbFaviconRepository struct {
	persistence.GormRepository
	logger logging.Logger
}

// Get returns the given favicon by the provided id
func (r *dbFaviconRepository) Get(id string) (Favicon, error) {
	var favicon Favicon
	h := r.Con().Where(&Favicon{ID: id}).First(&favicon)
	return favicon, h.Error
}

// Save creates a new favicon or updates an existing one
func (r *dbFaviconRepository) Save(item Favicon) (Favicon, error) {
	if item.ID == "" {
		return Favicon{}, fmt.Errorf("missing id for favicon")
	}

	if len(item.Payload) == 0 {
		return Favicon{}, fmt.Errorf("missing payload for favicon")
	}

	r.logger.Debug(fmt.Sprintf("store favicon: %s", item.ID))
	fav, err := r.Get(item.ID)
	if err == nil {
		fav.Payload = item.Payload
		fav.LastModified = time.Now()
		if h := r.Con().Save(&fav); h.Error != nil {
			return Favicon{}, fmt.Errorf("could not save favicon: %v", h.Error)
		}
		return fav, nil
	}
	fav = Favicon{
		ID:           item.ID,
		Payload:      item.Payload,
		LastModified: time.Now(),
	}
	if h := r.WriteCon().Create(&fav); h.Error != nil {
		return Favicon{}, fmt.Errorf("could not save favicon: %v", h.Error)
	}
	return fav, nil
}

// Delete is used to remove a favicon
func (r *dbFaviconRepository) Delete(item Favicon) error {
	if item.ID == "" {
		return fmt.Errorf("missing id for favicon")
	}

	h := r.WriteCon().Delete(&item)
	if h.Error != nil {
		return fmt.Errorf("cannot delete favicon by id '%s': %v", item.ID, h.Error)
	}
	return nil
}
