package store

import (
	"fmt"
	"time"

	"golang.binggl.net/monorepo/pkg/logging"
	"golang.binggl.net/monorepo/pkg/persistence"
)

// FaviconRepository offers CRUD functionality to store favicons
type FaviconRepository interface {
	Get(id string) (Favicon, error)
	Save(item Favicon) (Favicon, error)
	Delete(item Favicon) error
	InUnitOfWork(handle func(repo FaviconRepository) error) error
}

// CreateFaviconRepo creates a new repository
func CreateFaviconRepo(con persistence.Connection, logger logging.Logger) FaviconRepository {
	return &dbFaviconRepository{
		con:    con,
		logger: logger,
	}
}

// --------------------------------------------------------------------------
// Implementation
// --------------------------------------------------------------------------

type dbFaviconRepository struct {
	con    persistence.Connection
	logger logging.Logger
}

// Get returns the given favicon by the provided id
func (r *dbFaviconRepository) Get(id string) (Favicon, error) {
	var favicon Favicon
	h := r.con.R().Where(&Favicon{ID: id}).First(&favicon)
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
		if h := r.con.W().Save(&fav); h.Error != nil {
			return Favicon{}, fmt.Errorf("could not save favicon: %v", h.Error)
		}
		return fav, nil
	}
	fav = Favicon{
		ID:           item.ID,
		Payload:      item.Payload,
		LastModified: time.Now(),
	}
	if h := r.con.W().Create(&fav); h.Error != nil {
		return Favicon{}, fmt.Errorf("could not save favicon: %v", h.Error)
	}
	return fav, nil
}

// Delete is used to remove a favicon
func (r *dbFaviconRepository) Delete(item Favicon) error {
	if item.ID == "" {
		return fmt.Errorf("missing id for favicon")
	}

	h := r.con.W().Delete(&item)
	if h.Error != nil {
		return fmt.Errorf("cannot delete favicon by id '%s': %v", item.ID, h.Error)
	}
	return nil
}

// InUnitOfWork is used to perform logic in a transactional context
func (r *dbFaviconRepository) InUnitOfWork(handle func(repo FaviconRepository) error) error {
	return r.con.Begin(func(con persistence.Connection) error {
		repo := CreateFaviconRepo(con, r.logger)
		return handle(repo)
	})
}
