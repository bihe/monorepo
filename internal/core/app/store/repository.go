package store

import (
	"database/sql"
	"fmt"
	"strings"

	"golang.binggl.net/monorepo/pkg/logging"
	"gorm.io/gorm"
)

// --------------------------------------------------------------------------
// Repository interface
// --------------------------------------------------------------------------

// Repository defines the methods interacting with the store
type Repository interface {
	InUnitOfWork(fn func(repo Repository) error) error
	GetSitesForUser(user string) ([]UserSiteEntity, error)
	GetUsersForSite(site string) ([]string, error)
	GetLoginsForUser(user string) (int64, error)
	StoreSiteForUser(sites []UserSiteEntity) (err error)
	StoreLogin(login LoginsEntity) (err error)
}

// Create a new repository
func Create(db *gorm.DB, logger logging.Logger) Repository {
	return &dbRepository{
		db:     db,
		logger: logger,
	}
}

var _ Repository = &dbRepository{} // compile time interface check

// --------------------------------------------------------------------------
// Implementation
// --------------------------------------------------------------------------

type dbRepository struct {
	db       *gorm.DB
	logger   logging.Logger
	activeTX bool
}

// InUnitOfWork uses a transaction to execute the supplied function
func (r *dbRepository) InUnitOfWork(fn func(repo Repository) error) error {
	// be sure the stop recursion here
	if r.activeTX {
		return fmt.Errorf("a shared transaction is already available, will not start a new one")
	}
	return r.con().Transaction(func(tx *gorm.DB) error {
		return fn(&dbRepository{
			db:       tx, // the transaction is used as the shared connection
			logger:   r.logger,
			activeTX: true,
		})
	})
}

func (r *dbRepository) GetSitesForUser(user string) ([]UserSiteEntity, error) {
	var usersites []UserSiteEntity
	h := r.con().Order("name").Where("lower(user) = @user", sql.Named("user", strings.ToLower(user))).Find(&usersites)
	return usersites, h.Error
}

func (r *dbRepository) GetUsersForSite(site string) ([]string, error) {
	var users []string
	h := r.con().Model(&UserSiteEntity{}).Where("lower(name) = @site", sql.Named("site", strings.ToLower(site))).Pluck("user", &users)
	return users, h.Error
}

func (r *dbRepository) GetLoginsForUser(user string) (int64, error) {
	var c int64
	h := r.con().Model(&LoginsEntity{}).Where("lower(user) = @user", sql.Named("user", strings.ToLower(user))).Count(&c)
	return c, h.Error
}

func (r *dbRepository) StoreSiteForUser(sites []UserSiteEntity) (err error) {
	// brutal approach, delete all sites of the given user, and insert the new ones!
	h := r.con().Where(&UserSiteEntity{
		User: sites[0].User,
	}).Delete(UserSiteEntity{})
	if h.Error != nil {
		return fmt.Errorf("could not delete the sites for user, %w", h.Error)
	}

	h = r.con().Create(&sites)
	if h.Error != nil {
		return h.Error
	}
	if h.RowsAffected != int64(len(sites)) {
		err = fmt.Errorf("invalid number of rows affected, got %d", h.RowsAffected)
		return
	}
	return nil
}

func (r *dbRepository) StoreLogin(login LoginsEntity) (err error) {
	h := r.con().Create(&login)
	if h.Error != nil {
		return h.Error
	}
	if h.RowsAffected != 1 {
		err = fmt.Errorf("invalid number of rows affected, got %d", h.RowsAffected)
		return
	}
	return nil
}

// --------------------------------------------------------------------------
// internal logic / helpers
// --------------------------------------------------------------------------

// Con returns a database transaction, based on the state of the Repository this is a shared or transient connection
func (r *dbRepository) con() *gorm.DB {
	if r.db == nil {
		panic("no database connection is available")
	}
	return r.db
}
