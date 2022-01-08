package store_test

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"golang.binggl.net/monorepo/internal/core/app/store"
	"golang.binggl.net/monorepo/pkg/logging"

	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// const fatalErr = "an error '%s' was not expected when opening a stub database connection"
const expectations = "there were unfulfilled expectations: %s"
const errExpected = "error expected!"

var Err = fmt.Errorf("error")

var logger = logging.NewNop()

func repo(t *testing.T) (store.Repository, *gorm.DB) {
	var (
		DB  *gorm.DB
		err error
	)
	//if DB, err = gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{}); err != nil {
	if DB, err = gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{}); err != nil {
		t.Fatalf("cannot create database connection: %v", err)
	}
	// Migrate the schema
	DB.AutoMigrate(&store.UserSiteEntity{})
	return store.NewDBStore(DB), DB
}

func mockRepo(t *testing.T, useTx bool) (store.Repository, *gorm.DB, sqlmock.Sqlmock) {
	var (
		DB   *gorm.DB
		db   *sql.DB
		err  error
		mock sqlmock.Sqlmock
	)

	if db, mock, err = sqlmock.New(); err != nil {
		t.Fatalf("could not create a db-mock; %v", err)
	}

	if DB, err = gorm.Open(mysql.New(mysql.Config{Conn: db, SkipInitializeWithVersion: true}), &gorm.Config{
		SkipDefaultTransaction: !useTx,
	}); err != nil {
		t.Fatalf("cannot create database connection: %v", err)
	}
	return store.NewDBStore(DB), DB, mock
}

func closeRepo(db *gorm.DB, t *testing.T) {
	c, err := db.DB()
	if err != nil {
		t.Fatalf("could not close connection: %v", err)
	}
	c.Close()
}

func Test_New_Repository(t *testing.T) {
	repo, DB := repo(t)
	defer closeRepo(DB, t)
	if repo == nil {
		t.Errorf("cannot create a new repository")
	}
}

func Test_Create_Site(t *testing.T) {
	repo, DB := repo(t)
	defer closeRepo(DB, t)

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
	defer closeRepo(DB, t)

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

const skipTransactionUsage = false

func Test_Errors_Using_Mock(t *testing.T) {
	repo, DB, mock := mockRepo(t, skipTransactionUsage)
	defer closeRepo(DB, t)

	// -- wrong number of results after insert for USERSITES
	// ------------------------------------------------------------------------------------------------------------
	mock.ExpectExec("DELETE FROM `USERSITE`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO `USERSITE`").WillReturnResult(sqlmock.NewResult(1, 0))

	var sites []store.UserSiteEntity
	sites = append(sites, store.UserSiteEntity{
		Name:     "test",
		User:     "user",
		URL:      "url",
		PermList: "perm",
	})
	err := repo.StoreSiteForUser(sites)
	if err == nil {
		t.Errorf(errExpected)
	}

	// // -- error deleting USERSITES
	// // ------------------------------------------------------------------------------------------------------------
	mock.ExpectExec("DELETE FROM `USERSITE`").WillReturnError(Err)
	err = repo.StoreSiteForUser(sites)
	if err == nil {
		t.Errorf(errExpected)
	}

	// // -- error inserting USERSITES
	// // ------------------------------------------------------------------------------------------------------------
	mock.ExpectExec("DELETE FROM `USERSITE`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO `USERSITE`").WillReturnError(Err)
	err = repo.StoreSiteForUser(sites)
	if err == nil {
		t.Errorf(errExpected)
	}

	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf(expectations, err)
	}
}
