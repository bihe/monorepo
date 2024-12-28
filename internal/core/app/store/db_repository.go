package store

import (
	"database/sql"
	"fmt"
	"strings"

	"golang.binggl.net/monorepo/pkg/persistence"
)

var _ Repository = &dbRepository{} // compile time interface check

// NewDBStore creates a new DB repository instance
func NewDBStore(con persistence.Connection) *dbRepository {
	return &dbRepository{
		con: con,
	}
}

type dbRepository struct {
	con persistence.Connection
}

func (r *dbRepository) GetSitesForUser(user string) ([]UserSiteEntity, error) {
	var usersites []UserSiteEntity
	h := r.con.R().Order("name").Where("lower(user) = @user", sql.Named("user", strings.ToLower(user))).Find(&usersites)
	return usersites, h.Error
}

func (r *dbRepository) GetUsersForSite(site string) ([]string, error) {
	var users []string
	h := r.con.R().Model(&UserSiteEntity{}).Where("lower(name) = @site", sql.Named("site", strings.ToLower(site))).Pluck("user", &users)
	return users, h.Error
}

func (r *dbRepository) StoreSiteForUser(sites []UserSiteEntity) (err error) {
	// brutal approach, delete all sites of the given user, and insert the new ones!
	h := r.con.R().Where(&UserSiteEntity{
		User: sites[0].User,
	}).Delete(UserSiteEntity{})
	if h.Error != nil {
		return fmt.Errorf("could not delete the sites for user, %w", h.Error)
	}

	h = r.con.W().Create(&sites)
	if h.Error != nil {
		return h.Error
	}
	if h.RowsAffected != int64(len(sites)) {
		err = fmt.Errorf("invalid number of rows affected, got %d", h.RowsAffected)
		return
	}
	return nil
}

func (r *dbRepository) InUnitOfWork(handle func(repo Repository) error) error {
	return r.con.Begin(func(con persistence.Connection) error {
		repo := NewDBStore(con)
		return handle(repo)
	})
}
