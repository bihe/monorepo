package store

import (
	"fmt"
	"sync"
	"time"

	"golang.binggl.net/monorepo/pkg/logging"
	"gorm.io/gorm"
)

// FaviconRepository offers CRUD functionality to store favicons
type FaviconRepository interface {
	InUnitOfWork(fn func(repo FaviconRepository) error) error
	Get(id string) (Favicon, error)
	Save(item Favicon) (Favicon, error)
	Delete(item Favicon) error
}

// CreateBookmarkRepo a new repository
func CreateFaviconRepo(db *gorm.DB, logger logging.Logger) FaviconRepository {
	return &dbFaviconRepository{
		transient: db,
		shared:    nil,
		logger:    logger,
	}
}

// --------------------------------------------------------------------------
// Implementation
// --------------------------------------------------------------------------

type dbFaviconRepository struct {
	transient *gorm.DB
	shared    *gorm.DB
	logger    logging.Logger
	sync.Mutex
}

// InUnitOfWork uses a transaction to execute the supplied function
func (r *dbFaviconRepository) InUnitOfWork(fn func(repo FaviconRepository) error) error {
	return r.con().Transaction(func(tx *gorm.DB) error {
		// be sure the stop recursion here
		if r.shared != nil {
			return fmt.Errorf("a shared connection/transaction is already available, will not start a new one")
		}

		// lock concurrent access for transactional tasks
		r.Lock()
		defer r.Unlock()

		return fn(&dbFaviconRepository{
			transient: r.transient,
			shared:    tx, // the transaction is used as the shared connection
			logger:    r.logger,
		})
	})
}

// Get returns the given favicon by the provided id
func (r *dbFaviconRepository) Get(id string) (Favicon, error) {
	var favicon Favicon
	h := r.con().Where(&Favicon{ID: id}).First(&favicon)
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
		if h := r.con().Save(&fav); h.Error != nil {
			return Favicon{}, fmt.Errorf("could not save favicon: %v", h.Error)
		}
		return fav, nil
	}
	fav = Favicon{
		ID:           item.ID,
		Payload:      item.Payload,
		LastModified: time.Now(),
	}
	if h := r.con().Create(&fav); h.Error != nil {
		return Favicon{}, fmt.Errorf("could not save favicon: %v", h.Error)
	}
	return fav, nil
}

// Delete is used to remove a favicon
func (r *dbFaviconRepository) Delete(item Favicon) error {
	if item.ID == "" {
		return fmt.Errorf("missing id for favicon")
	}

	h := r.con().Delete(&item)
	if h.Error != nil {
		return fmt.Errorf("cannot delete favicon by id '%s': %v", item.ID, h.Error)
	}
	return nil
}

// --------------------------------------------------------------------------
// internal logic / helpers
// --------------------------------------------------------------------------

func (r *dbFaviconRepository) con() *gorm.DB {
	if r.shared != nil {
		return r.shared
	}
	if r.transient == nil {
		panic("no database connection is available")
	}
	return r.transient
}
