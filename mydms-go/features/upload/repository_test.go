package upload

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/bihe/mydms/internal/persistence"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

const fatalErr = "an error '%s' was not expected when opening a stub database connection"
const expectations = "there were unfulfilled expectations: %s"
const deleteExpErr = "error was not expected while delete item: %v"

var uploadItem = Upload{
	ID:       "id",
	FileName: "filename",
	MimeType: "mimetype",
	Created:  time.Now().UTC(),
}

func TestNewReaderWriter(t *testing.T) {
	_, err := NewRepository(persistence.Connection{})
	if err == nil {
		t.Errorf("no reader without connection possible")
	}

	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf(fatalErr, err)
	}
	defer db.Close()

	dbx := sqlx.NewDb(db, "mysql")
	_, err = NewRepository(persistence.NewFromDB(dbx))
	if err != nil {
		t.Errorf("could not get a reader: %v", err)
	}
}

func TestWrite(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf(fatalErr, err)
	}
	defer db.Close()

	dbx := sqlx.NewDb(db, "mysql")
	c := persistence.NewFromDB(dbx)
	rw := dbRepository{c}
	stmt := "INSERT INTO UPLOADS"

	item := uploadItem

	mock.ExpectBegin()
	mock.ExpectExec(stmt).WithArgs(item.ID, item.FileName, item.MimeType, item.Created).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	// now we execute our method
	if err = rw.Write(item, persistence.Atomic{}); err != nil {
		t.Errorf("error was not expected while insert item: %v", err)
	}

	// externally supplied tx
	mock.ExpectBegin()
	mock.ExpectExec(stmt).WithArgs(item.ID, item.FileName, item.MimeType, item.Created).WillReturnResult(sqlmock.NewResult(1, 1))

	a, err := c.CreateAtomic()
	if err = rw.Write(item, a); err != nil {
		t.Errorf("error was not expected for insert item: %v", err)
	}

	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf(expectations, err)
	}
}

func TestWriteError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf(fatalErr, err)
	}
	defer db.Close()

	dbx := sqlx.NewDb(db, "mysql")
	c := persistence.NewFromDB(dbx)
	rw := dbRepository{c}

	item := uploadItem

	mock.ExpectBegin()
	mock.ExpectRollback()

	// now we execute our method
	if err = rw.Write(item, persistence.Atomic{}); err == nil {
		t.Errorf("error was expected for insert item")
	}

	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf(expectations, err)
	}
}

func TestRead(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf(fatalErr, err)
	}
	defer db.Close()

	dbx := sqlx.NewDb(db, "mysql")
	c := persistence.NewFromDB(dbx)
	rw := dbRepository{c}
	columns := []string{"id", "filename", "mimetype", "created"}
	q := "SELECT id, filename, mimetype, created FROM UPLOADS"
	id := "id"

	expected := uploadItem

	// success
	rows := sqlmock.NewRows(columns).
		AddRow(expected.ID, expected.FileName, expected.MimeType, expected.Created)
	mock.ExpectQuery(q).WithArgs(id).WillReturnRows(rows)

	item, err := rw.Read(id)
	if err != nil {
		t.Errorf("could not get item: %v", err)
	}

	assert.Equal(t, expected.ID, item.ID)
	assert.Equal(t, expected.FileName, item.FileName)
	assert.Equal(t, expected.MimeType, item.MimeType)
	assert.Equal(t, expected.Created, item.Created)

	// no result
	rows = sqlmock.NewRows(columns)
	mock.ExpectQuery(q).WithArgs(id).WillReturnRows(rows)

	item, err = rw.Read(id)
	if err == nil {
		t.Errorf("should have returned an error")
	}

	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf(expectations, err)
	}
}

func TestDelete(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf(fatalErr, err)
	}
	defer db.Close()

	dbx := sqlx.NewDb(db, "mysql")
	c := persistence.NewFromDB(dbx)
	rw := dbRepository{c}
	stmt := "DELETE FROM UPLOADS"

	item := uploadItem

	mock.ExpectBegin()
	mock.ExpectExec(stmt).WithArgs(item.ID).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	// now we execute our method
	if err = rw.Delete(item.ID, persistence.Atomic{}); err != nil {
		t.Errorf(deleteExpErr, err)
	}

	// externally supplied tx
	mock.ExpectBegin()
	mock.ExpectExec(stmt).WithArgs(item.ID).WillReturnResult(sqlmock.NewResult(1, 1))

	a, err := c.CreateAtomic()
	if err = rw.Delete(item.ID, a); err != nil {
		t.Errorf("error was not expected while delete item: %v", err)
	}

	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf(expectations, err)
	}
}

func TestDeleteError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf(fatalErr, err)
	}
	defer db.Close()

	dbx := sqlx.NewDb(db, "mysql")
	c := persistence.NewFromDB(dbx)
	rw := dbRepository{c}

	item := uploadItem

	mock.ExpectBegin()
	mock.ExpectRollback()

	// now we execute our method
	if err = rw.Delete(item.ID, persistence.Atomic{}); err == nil {
		t.Errorf("error was expected for insert item")
	}

	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf(expectations, err)
	}
}
