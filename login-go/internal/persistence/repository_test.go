package persistence

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"

	per "golang.binggl.net/commons/persistence"
)

const fatalErr = "an error '%s' was not expected when opening a stub database connection"
const expectations = "there were unfulfilled expectations: %s"
const errExpected = "error expected!"

var Err = fmt.Errorf("error")

func TestNewRepository(t *testing.T) {
	_, err := NewRepository(per.Connection{})
	if err == nil {
		t.Errorf("no reader without connection possible")
	}

	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf(fatalErr, err)
	}
	defer db.Close()

	dbx := sqlx.NewDb(db, "mysql")
	_, err = NewRepository(per.NewFromDB(dbx))
	if err != nil {
		t.Errorf("could not get a repository: %v", err)
	}
}

func TestAtomic(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf(fatalErr, err)
	}
	defer db.Close()

	dbx := sqlx.NewDb(db, "mysql")
	repo, err := NewRepository(per.NewFromDB(dbx))
	if err != nil {
		t.Errorf("could not get a repository: %v", err)
	}

	mock.ExpectBegin()

	_, err = repo.CreateAtomic()
	if err != nil {
		t.Errorf("could not ceate a new atomic object: %v", err)
	}

	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf(expectations, err)
	}
}

func TestGetSiteByUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf(fatalErr, err)
	}
	defer db.Close()

	dbx := sqlx.NewDb(db, "mysql")
	c := per.NewFromDB(dbx)
	repo := dbRepository{c}

	q := "SELECT name,user,url,permission_list,created FROM USERSITE"
	columns := []string{"name", "user", "url", "permission_list", "created"}

	now := time.Now().UTC()
	username := "USER"
	errCould := "could not get user-sites: %v"

	// success
	rows := sqlmock.NewRows(columns).
		AddRow("name1", "user1", "http://url1", username, now).
		AddRow("name2", "user2", "http://url2", username, now)
	mock.ExpectQuery(q).WithArgs(strings.ToLower(username)).WillReturnRows(rows)

	sites, err := repo.GetSitesByUser(username)
	if err != nil {
		t.Errorf(errCould, err)
	}
	if len(sites) != 2 {
		t.Errorf("expected 2 items, got %d", len(sites))
	}

	// no results
	rows = sqlmock.NewRows(columns)
	mock.ExpectQuery(q).WillReturnRows(rows)

	sites, err = repo.GetSitesByUser(username)
	if err != nil {
		t.Errorf(errCould, err)
	}
	if len(sites) != 0 {
		t.Errorf("expected 0 items, got %d", len(sites))
	}

	// error
	mock.ExpectQuery(q).WillReturnError(Err)

	sites, err = repo.GetSitesByUser(username)
	if err == nil {
		t.Errorf(errExpected)
	}

	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf(expectations, err)
	}
}

func TestStoreSiteForUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf(fatalErr, err)
	}
	defer db.Close()

	dbx := sqlx.NewDb(db, "mysql")
	c := per.NewFromDB(dbx)
	repo := dbRepository{c}

	errSites := "could not save user-sites: %v"
	del := "DELETE FROM USERSITE WHERE"
	stmt := "INSERT INTO USERSITE \\(name,user,url,permission_list, created\\)"
	username := "USER"
	now := time.Now().UTC()
	sites := []UserSite{
		UserSite{
			Name:     "site1",
			User:     username,
			URL:      "http://url1",
			PermList: "perm1;perm2",
			Created:  now,
		},
	}

	// success
	mock.ExpectBegin()
	mock.ExpectExec(del).WithArgs(username).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(stmt).WithArgs(sites[0].Name, sites[0].User, sites[0].URL, sites[0].PermList, sites[0].Created).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err = repo.StoreSiteForUser(username, sites, per.Atomic{})
	if err != nil {
		t.Errorf(errSites, err)
	}

	// external TX
	mock.ExpectBegin()
	mock.ExpectExec(del).WithArgs(username).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(stmt).WithArgs(sites[0].Name, sites[0].User, sites[0].URL, sites[0].PermList, sites[0].Created).WillReturnResult(sqlmock.NewResult(1, 1))

	a, err := c.CreateAtomic()
	err = repo.StoreSiteForUser(username, sites, a)
	if err != nil {
		t.Errorf(errSites, err)
	}

	// error cannot delete
	mock.ExpectBegin()
	mock.ExpectExec(del).WithArgs(username).WillReturnError(Err)
	mock.ExpectRollback()

	err = repo.StoreSiteForUser(username, sites, per.Atomic{})
	if err == nil {
		t.Errorf(errExpected)
	}

	// cannot insert
	mock.ExpectBegin()
	mock.ExpectExec(del).WithArgs(username).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(stmt).WithArgs(sites[0].Name, sites[0].User, sites[0].URL, sites[0].PermList, sites[0].Created).WillReturnError(Err)
	mock.ExpectRollback()

	err = repo.StoreSiteForUser(username, sites, per.Atomic{})
	if err == nil {
		t.Errorf(errExpected)
	}

	// rows affected
	mock.ExpectBegin()
	mock.ExpectExec(del).WithArgs(username).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(stmt).WithArgs(sites[0].Name, sites[0].User, sites[0].URL, sites[0].PermList, sites[0].Created).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectRollback()

	err = repo.StoreSiteForUser(username, sites, per.Atomic{})
	if err == nil {
		t.Errorf(errExpected)
	}

	// wrong number of rows
	mock.ExpectBegin()
	mock.ExpectExec(del).WithArgs(username).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(stmt).WithArgs(sites[0].Name, sites[0].User, sites[0].URL, sites[0].PermList, sites[0].Created).WillReturnResult(sqlmock.NewErrorResult(Err))
	mock.ExpectRollback()

	err = repo.StoreSiteForUser(username, sites, per.Atomic{})
	if err == nil {
		t.Errorf(errExpected)
	}

	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf(expectations, err)
	}
}

func TestStoreLogin(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf(fatalErr, err)
	}
	defer db.Close()

	dbx := sqlx.NewDb(db, "mysql")
	c := per.NewFromDB(dbx)
	repo := dbRepository{c}

	errLogin := "could not save login: %v"
	now := time.Now().UTC()
	stmt := "INSERT INTO LOGINS \\(user,created,type\\)"
	login := Login{User: "USER", Type: DIRECT, Created: now}

	// success
	mock.ExpectBegin()
	mock.ExpectExec(stmt).WithArgs(login.User, login.Created, login.Type).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err = repo.StoreLogin(login, per.Atomic{})
	if err != nil {
		t.Errorf(errLogin, err)
	}

	// error
	mock.ExpectBegin()
	mock.ExpectExec(stmt).WillReturnError(Err)
	mock.ExpectRollback()

	err = repo.StoreLogin(login, per.Atomic{})
	if err == nil {
		t.Errorf(errExpected)
	}

	// results err
	mock.ExpectBegin()
	mock.ExpectExec(stmt).WithArgs(login.User, login.Created, login.Type).WillReturnResult(sqlmock.NewErrorResult(Err))
	mock.ExpectRollback()

	err = repo.StoreLogin(login, per.Atomic{})
	if err == nil {
		t.Errorf(errExpected)
	}

	// no results
	mock.ExpectBegin()
	mock.ExpectExec(stmt).WithArgs(login.User, login.Created, login.Type).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectRollback()

	err = repo.StoreLogin(login, per.Atomic{})
	if err == nil {
		t.Errorf(errExpected)
	}

	// externally supplied tx
	mock.ExpectBegin()
	mock.ExpectExec(stmt).WithArgs(login.User, login.Created, login.Type).WillReturnResult(sqlmock.NewResult(1, 1))

	a, err := c.CreateAtomic()
	err = repo.StoreLogin(login, a)
	if err != nil {
		t.Errorf(errLogin, err)
	}

	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf(expectations, err)
	}
}

func TestGetUsersForSite(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf(fatalErr, err)
	}
	defer db.Close()

	dbx := sqlx.NewDb(db, "mysql")
	c := per.NewFromDB(dbx)
	repo := dbRepository{c}

	q := "SELECT user FROM USERSITE"
	columns := []string{"user"}
	errCould := "could not get users for site: %v"
	site := "site"

	// success
	rows := sqlmock.NewRows(columns).
		AddRow("user1").
		AddRow("user2")
	mock.ExpectQuery(q).WithArgs(site).WillReturnRows(rows)

	sites, err := repo.GetUsersForSite(site)
	if err != nil {
		t.Errorf(errCould, err)
	}
	if len(sites) != 2 {
		t.Errorf("expected 2 items, got %d", len(sites))
	}

	// no results
	rows = sqlmock.NewRows(columns)
	mock.ExpectQuery(q).WillReturnRows(rows)

	sites, err = repo.GetUsersForSite(site)
	if err != nil {
		t.Errorf(errCould, err)
	}
	if len(sites) != 0 {
		t.Errorf("expected 0 items, got %d", len(sites))
	}

	// error
	mock.ExpectQuery(q).WillReturnError(Err)

	sites, err = repo.GetUsersForSite(site)
	if err == nil {
		t.Errorf(errExpected)
	}

	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf(expectations, err)
	}
}
