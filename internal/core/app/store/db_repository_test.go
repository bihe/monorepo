package store_test

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.binggl.net/monorepo/internal/core/app/store"
	"golang.binggl.net/monorepo/pkg/persistence"
)

var Err = fmt.Errorf("error")

func repo(t *testing.T) (store.Repository, *sql.DB) {
	var (
		err error
	)
	params := make([]persistence.SqliteParam, 0)
	con, err := persistence.CreateGormSqliteCon(":memory:", params)
	if err != nil {
		t.Fatalf("cannot create database connection: %v", err)
	}
	// Migrate the schema
	con.Write.AutoMigrate(&store.UserSiteEntity{})
	con.Read.AutoMigrate(&store.UserSiteEntity{})
	db, err := con.Write.DB()
	if err != nil {
		t.Fatalf("could not get DB handle; %v", err)
	}
	con.Read = con.Write
	repo := store.NewDBStore(con)
	return repo, db
}

func Test_New_Repository(t *testing.T) {
	repo, DB := repo(t)
	defer DB.Close()
	if repo == nil {
		t.Errorf("cannot create a new repository")
	}
}

func Test_Create_Site(t *testing.T) {
	repo, DB := repo(t)
	defer DB.Close()

	user := "test_" + time.Now().String()
	var sites []store.UserSiteEntity
	sites = append(sites, store.UserSiteEntity{
		Name:     "site1",
		User:     user,
		URL:      "http://www.site.com",
		PermList: "role1;role2",
	})
	err := repo.StoreSiteForUser(sites)
	if err != nil {
		t.Errorf("could not create sites for user: %v", err)
	}

	s, err := repo.GetSitesForUser(user)
	if err != nil {
		t.Errorf("could not fetch sites for user: %v", err)
	}
	assert.Equal(t, 1, len(s))
	assert.Equal(t, "site1", s[0].Name)
	assert.Equal(t, user, s[0].User)
	assert.Equal(t, "http://www.site.com", s[0].URL)
	assert.Equal(t, "role1;role2", s[0].PermList)

	assert.True(t, s[0].String() != "")
}

func Test_Get_Users_For_Site(t *testing.T) {
	repo, DB := repo(t)
	defer DB.Close()

	user1 := "user1_" + time.Now().String()
	var sites1 []store.UserSiteEntity
	sites1 = append(sites1, store.UserSiteEntity{
		Name:     "site1",
		User:     user1,
		URL:      "http://www.site.com",
		PermList: "role1;role2",
	})
	err := repo.StoreSiteForUser(sites1)
	if err != nil {
		t.Errorf("could not create sites for user: %v", err)
	}

	user2 := "user2_" + time.Now().String()
	var sites2 []store.UserSiteEntity
	sites2 = append(sites2, store.UserSiteEntity{
		Name:     "site1",
		User:     user2,
		URL:      "http://www.site.com",
		PermList: "role1;role2",
	})
	err = repo.StoreSiteForUser(sites2)
	if err != nil {
		t.Errorf("could not create sites for user: %v", err)
	}

	users, err := repo.GetUsersForSite("site1")
	if err != nil {
		t.Errorf("could not get users for site: %v", err)
	}
	assert.Equal(t, 2, len(users))
	var found bool
	for _, item := range users {
		if item == user1 || item == user2 {
			found = true
		}
	}
	assert.True(t, found)
}

func Test_Create_Site_UnitOfWork(t *testing.T) {
	repo, DB := repo(t)
	defer DB.Close()

	user := "test_" + time.Now().String()
	var sites []store.UserSiteEntity
	sites = append(sites, store.UserSiteEntity{
		Name:     "site1",
		User:     user,
		URL:      "http://www.site.com",
		PermList: "role1;role2",
	})

	err := repo.InUnitOfWork(func(r store.Repository) error {
		err := r.StoreSiteForUser(sites)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		t.Errorf("could not perform within transaction: %v", err)
	}
	sites, err = repo.GetSitesForUser(user)
	if err != nil {
		t.Errorf("could not get site: %v", err)
	}

	assert.Equal(t, 1, len(sites))
	assert.Equal(t, "site1", sites[0].Name)
	assert.Equal(t, user, sites[0].User)
	assert.Equal(t, "http://www.site.com", sites[0].URL)
	assert.Equal(t, "role1;role2", sites[0].PermList)
	assert.True(t, sites[0].String() != "")

	// perform with transaction rollback
	err = repo.InUnitOfWork(func(r store.Repository) error {
		sites, err = r.GetSitesForUser(user)
		if err != nil {
			t.Errorf("could not get site: %v", err)
		}

		sites[0].Name = "site_changed!"
		err := r.StoreSiteForUser(sites)
		if err != nil {
			return err
		}

		sites, err = r.GetSitesForUser(user)
		if err != nil {
			t.Errorf("could not get site: %v", err)
		}
		assert.Equal(t, "site_changed!", sites[0].Name)

		// forcefully roll back
		return fmt.Errorf("could not update within transaction")
	})

	// outsite of the rolled-back transaction we should get the old name
	sites, err = repo.GetSitesForUser(user)
	if err != nil {
		t.Errorf("could not get site: %v", err)
	}
	assert.Equal(t, "site1", sites[0].Name)

}
