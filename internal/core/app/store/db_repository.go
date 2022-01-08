package store

import (
	"database/sql"
	"fmt"
	"strings"

	"golang.binggl.net/monorepo/pkg/logging"
	"gorm.io/gorm"
)

var _ Repository = &dbRepository{} // compile time interface check

// NewDBStore creates a new DB repository intance
func NewDBStore(db *gorm.DB, logger logging.Logger) *dbRepository {
	return &dbRepository{
		db:     db,
		logger: logger,
	}
}

type dbRepository struct {
	db     *gorm.DB
	logger logging.Logger
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
