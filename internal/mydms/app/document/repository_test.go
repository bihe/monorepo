package document

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"golang.binggl.net/monorepo/internal/mydms/app/shared"
)

const fatalErr = "an error '%s' was not expected when opening a stub database connection"
const expectations = "there were unfulfilled expectations: %s"
const deleteExpErr = "error was not expected while delete item: %v"
const existsErr = "error was not expected while checking for existence of item: %v"
const expectedErr = "error expected"

const stmtInsertDocs = "INSERT INTO DOCUMENTS"
const queryDocs = "SELECT id,title,filename,alternativeid,previewlink,amount,taglist,senderlist,created,modified,invoicenumber FROM DOCUMENTS"

var Err = fmt.Errorf("error")

func getDbx(t *testing.T) (dbx *sqlx.DB, db *sql.DB, mock sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf(fatalErr, err)
	}
	dbx = sqlx.NewDb(db, "sqlite3")
	return
}

func TestAtomic(t *testing.T) {
	dbx, db, mock := getDbx(t)
	defer db.Close()
	repo, err := NewRepository(shared.NewFromDB(dbx))
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

func TestNewRepository(t *testing.T) {
	_, err := NewRepository(shared.Connection{})
	if err == nil {
		t.Errorf("no reader without connection possible")
	}

	dbx, db, _ := getDbx(t)
	defer db.Close()
	_, err = NewRepository(shared.NewFromDB(dbx))
	if err != nil {
		t.Errorf("could not get a repository: %v", err)
	}
}

func TestSave(t *testing.T) {
	dbx, db, mock := getDbx(t)
	defer db.Close()

	var now = time.Now().UTC().Add(-(time.Second * 10))

	c := shared.NewFromDB(dbx)
	rw := dbRepository{c}

	item := DocEntity{
		Title:      "title",
		FileName:   "filename",
		Amount:     10,
		TagList:    "taglist",
		SenderList: "senderlist",
	}

	errInsert := "error was not expected while inserting item: %v"

	// INSERT
	mock.ExpectBegin()
	mock.ExpectExec(stmtInsertDocs).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	var (
		d   DocEntity
		err error
	)
	if d, err = rw.Save(item, shared.Atomic{}); err != nil {
		t.Errorf(errInsert, err)
	}
	assert.Equal(t, item.Title, d.Title)
	assert.Equal(t, item.FileName, d.FileName)
	assert.Equal(t, item.Amount, d.Amount)
	assert.Equal(t, item.TagList, d.TagList)
	assert.Equal(t, item.SenderList, d.SenderList)
	assert.True(t, d.ID != "")
	assert.True(t, d.AltID != "")

	// UPDATE
	mock.ExpectBegin()
	item.ID = uuid.New().String()
	item.AltID = d.AltID

	rows := sqlmock.NewRows([]string{"id", "title", "filename", "alternativeid", "previewlink", "amount", "taglist", "senderlist", "created", "modified", "invoicenumber"}).
		AddRow(item.ID, item.Title, item.FileName, item.AltID, item.PreviewLink, item.Amount, item.TagList, item.SenderList, d.Created, nil, item.InvoiceNumber)
	mock.ExpectQuery(queryDocs).WillReturnRows(rows)
	mock.ExpectExec("UPDATE DOCUMENTS").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	var up DocEntity
	if up, err = rw.Save(item, shared.Atomic{}); err != nil {
		t.Errorf(errInsert, err)
	}
	assert.Equal(t, item.ID, up.ID)
	assert.Equal(t, item.AltID, up.AltID)
	assert.Equal(t, item.Title, up.Title)
	assert.Equal(t, item.FileName, up.FileName)
	assert.Equal(t, item.Amount, up.Amount)
	assert.Equal(t, item.TagList, up.TagList)
	assert.Equal(t, item.SenderList, up.SenderList)
	assert.Equal(t, d.Created, up.Created)
	t.Logf("now: %+v, updated: %+v", now, up.Modified.Time)
	assert.True(t, up.Modified.Time.After(now))
	assert.Equal(t, item.InvoiceNumber, up.InvoiceNumber)

	// UPDATE with wrong ID
	mock.ExpectBegin()
	item.ID = uuid.New().String()
	item.AltID = d.AltID

	mock.ExpectQuery(queryDocs).WillReturnError(fmt.Errorf("no rows"))
	mock.ExpectExec(stmtInsertDocs).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	if up, err = rw.Save(item, shared.Atomic{}); err != nil {
		t.Errorf(errInsert, err)
	}

	assert.NotEqual(t, item.AltID, up.AltID)
	assert.Equal(t, item.Title, up.Title)
	assert.Equal(t, item.FileName, up.FileName)
	assert.Equal(t, item.Amount, up.Amount)
	assert.Equal(t, item.TagList, up.TagList)
	assert.Equal(t, item.SenderList, up.SenderList)
	assert.True(t, up.Created.After(now))

	// externally supplied tx
	item.ID = ""
	mock.ExpectBegin()
	mock.ExpectExec(stmtInsertDocs).WillReturnResult(sqlmock.NewResult(1, 1))
	a, _ := c.CreateAtomic()
	if d, err = rw.Save(item, a); err != nil {
		t.Errorf(errInsert, err)
	}
	assert.Equal(t, item.Title, d.Title)
	assert.Equal(t, item.FileName, d.FileName)
	assert.Equal(t, item.Amount, d.Amount)
	assert.Equal(t, item.TagList, d.TagList)
	assert.Equal(t, item.SenderList, d.SenderList)
	assert.True(t, d.ID != "")
	assert.True(t, d.AltID != "")

	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf(expectations, err)
	}
}

func TestSaveError(t *testing.T) {
	var err error
	dbx, db, mock := getDbx(t)
	defer db.Close()
	c := shared.NewFromDB(dbx)
	rw := dbRepository{c}

	item := DocEntity{
		Title:      "title",
		FileName:   "filename",
		Amount:     10,
		TagList:    "taglist",
		SenderList: "senderlist",
	}

	var errInsert = "error was expected while insert item"

	// INSERT Error
	mock.ExpectBegin()
	mock.ExpectExec(stmtInsertDocs).WillReturnError(fmt.Errorf("does not work"))
	mock.ExpectRollback()

	if _, err = rw.Save(item, shared.Atomic{}); err == nil {
		t.Error(errInsert)
	}

	// Rows affected Error
	mock.ExpectBegin()
	mock.ExpectExec(stmtInsertDocs).WillReturnResult(sqlmock.NewErrorResult(fmt.Errorf("result error")))
	mock.ExpectRollback()

	if _, err = rw.Save(item, shared.Atomic{}); err == nil {
		t.Error(errInsert)
	}

	// Rows affected number mismatch
	mock.ExpectBegin()
	mock.ExpectExec(stmtInsertDocs).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectRollback()

	if _, err = rw.Save(item, shared.Atomic{}); err == nil {
		t.Error(errInsert)
	}
}

func TestRead(t *testing.T) {
	var err error
	dbx, db, mock := getDbx(t)
	defer db.Close()
	c := shared.NewFromDB(dbx)
	rw := dbRepository{c}
	columns := []string{"id", "title", "filename", "alternativeid", "previewlink", "amount", "taglist", "senderlist", "created", "modified", "invoicenumber"}
	q := queryDocs
	id := "id"

	expected := DocEntity{
		ID:            "id",
		Title:         "title",
		FileName:      "filename",
		AltID:         "altid",
		PreviewLink:   sql.NullString{String: "previewlink", Valid: true},
		Amount:        1.0,
		Created:       time.Now().UTC(),
		Modified:      sql.NullTime{},
		TagList:       "tags",
		SenderList:    "senders",
		InvoiceNumber: sql.NullString{String: "invoicenumber", Valid: true},
	}

	// success
	rows := sqlmock.NewRows(columns).
		AddRow(expected.ID, expected.Title, expected.FileName, expected.AltID, expected.PreviewLink, expected.Amount, expected.TagList, expected.SenderList, expected.Created, expected.Modified, expected.InvoiceNumber)
	mock.ExpectQuery(q).WithArgs(id).WillReturnRows(rows)

	item, err := rw.Get(id)
	if err != nil {
		t.Errorf("could not get item: %v", err)
	}

	assert.Equal(t, expected.ID, item.ID)
	assert.Equal(t, expected.Title, item.Title)
	assert.Equal(t, expected.FileName, item.FileName)
	assert.Equal(t, expected.AltID, item.AltID)
	assert.Equal(t, expected.PreviewLink, item.PreviewLink)
	assert.Equal(t, expected.Amount, item.Amount)
	assert.Equal(t, expected.TagList, item.TagList)
	assert.Equal(t, expected.SenderList, item.SenderList)
	assert.Equal(t, expected.Created, item.Created)
	assert.Equal(t, expected.Modified, item.Modified)
	assert.Equal(t, expected.InvoiceNumber, item.InvoiceNumber)

	// no result
	rows = sqlmock.NewRows(columns)
	mock.ExpectQuery(q).WithArgs(id).WillReturnRows(rows)

	item, err = rw.Get(id)
	if err == nil {
		t.Errorf("should have returned an error")
	}

	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf(expectations, err)
	}
}

func TestDelete(t *testing.T) {
	var err error
	dbx, db, mock := getDbx(t)
	defer db.Close()
	c := shared.NewFromDB(dbx)
	rw := dbRepository{c}
	stmt := "DELETE FROM DOCUMENTS"

	item := DocEntity{
		ID: "id",
	}

	mock.ExpectBegin()
	mock.ExpectExec(stmt).WithArgs(item.ID).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	// now we execute our method
	if err = rw.Delete(item.ID, shared.Atomic{}); err != nil {
		t.Errorf(deleteExpErr, err)
	}

	// externally supplied tx
	mock.ExpectBegin()
	mock.ExpectExec(stmt).WithArgs(item.ID).WillReturnResult(sqlmock.NewResult(1, 1))

	a, _ := c.CreateAtomic()
	if err = rw.Delete(item.ID, a); err != nil {
		t.Errorf("error was not expected while delete item: %v", err)
	}

	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf(expectations, err)
	}
}

func TestExists(t *testing.T) {
	var err error
	dbx, db, mock := getDbx(t)
	defer db.Close()
	c := shared.NewFromDB(dbx)
	rw := dbRepository{c}
	q := "SELECT filename FROM DOCUMENTS"
	id := "id"
	fileName := "test.pdf"
	rows := []string{"filename"}

	mock.ExpectBegin()
	mock.ExpectQuery(q).WithArgs(id).WillReturnRows(sqlmock.NewRows(rows).AddRow(fileName))
	mock.ExpectCommit()

	// now we execute our method
	var f string
	if f, err = rw.Exists(id, shared.Atomic{}); err != nil {
		t.Errorf(existsErr, err)
	}
	if f != fileName {
		t.Errorf("filename does not match")
	}

	// externally supplied tx
	mock.ExpectBegin()
	mock.ExpectQuery(q).WithArgs(id).WillReturnRows(sqlmock.NewRows(rows).AddRow(fileName))
	a, _ := c.CreateAtomic()
	if _, err = rw.Exists(id, a); err != nil {
		t.Errorf(existsErr, err)
	}

	// error
	mock.ExpectBegin()
	mock.ExpectQuery(q).WithArgs(id).WillReturnError(fmt.Errorf("error"))
	mock.ExpectRollback()

	if _, err = rw.Exists(id, shared.Atomic{}); err == nil {
		t.Error(expectedErr)
	}

	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf(expectations, err)
	}
}

func TestDeleteError(t *testing.T) {
	var err error
	dbx, db, mock := getDbx(t)
	defer db.Close()
	c := shared.NewFromDB(dbx)
	rw := dbRepository{c}

	item := DocEntity{
		ID: "id",
	}

	mock.ExpectBegin()
	mock.ExpectRollback()

	// now we execute our method
	if err = rw.Delete(item.ID, shared.Atomic{}); err == nil {
		t.Errorf("error was expected for insert item")
	}

	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf(expectations, err)
	}
}

func TestSearch(t *testing.T) {
	var err error
	dbx, db, mock := getDbx(t)
	defer db.Close()
	c := shared.NewFromDB(dbx)
	rw := dbRepository{c}
	columns := []string{"id", "title", "filename", "alternativeid", "previewlink", "amount", "taglist", "senderlist", "created", "modified", "invoicenumber"}

	qc := "SELECT count\\(id\\) FROM DOCUMENTS"

	expected := DocEntity{
		ID:            "id",
		Title:         "title",
		FileName:      "filename",
		AltID:         "altid",
		PreviewLink:   sql.NullString{String: "previewlink", Valid: true},
		Amount:        1.0,
		Created:       time.Now().UTC(),
		Modified:      sql.NullTime{},
		TagList:       "tags",
		SenderList:    "senders",
		InvoiceNumber: sql.NullString{String: "invoicenumber", Valid: true},
	}

	// success
	cr := sqlmock.NewRows([]string{"count(id)"}).AddRow(1)
	mock.ExpectQuery(qc).WillReturnRows(cr)

	dr := sqlmock.NewRows(columns).
		AddRow(expected.ID, expected.Title, expected.FileName, expected.AltID, expected.PreviewLink, expected.Amount, expected.TagList, expected.SenderList, expected.Created, expected.Modified, expected.InvoiceNumber)
	mock.ExpectQuery(queryDocs).WillReturnRows(dr)

	ts := time.Now().UTC()
	from := ts.Add(-time.Hour)
	until := ts.Add(time.Hour)
	search := DocSearch{
		Skip:   1,
		Limit:  1,
		Title:  "title",
		Tag:    "tags",
		Sender: "senders",
		From:   from,
		Until:  until,
	}
	order := []OrderBy{
		{Order: DESC, Field: "modified"},
		{Order: ASC, Field: "title"},
	}

	doc, err := rw.Search(search, order)
	if err != nil {
		t.Errorf("could not query documents: %v", err)
	}
	if len(doc.Documents) == 0 {
		t.Errorf("document list empty, nothing returned")
	}

	item := doc.Documents[0]

	assert.Equal(t, expected.ID, item.ID)
	assert.Equal(t, expected.Title, item.Title)
	assert.Equal(t, expected.FileName, item.FileName)
	assert.Equal(t, expected.AltID, item.AltID)
	assert.Equal(t, expected.PreviewLink, item.PreviewLink)
	assert.Equal(t, expected.Amount, item.Amount)
	assert.Equal(t, expected.TagList, item.TagList)
	assert.Equal(t, expected.SenderList, item.SenderList)
	assert.Equal(t, expected.Created, item.Created)
	assert.Equal(t, expected.Modified, item.Modified)
	assert.Equal(t, expected.InvoiceNumber, item.InvoiceNumber)

	// failure1
	mock.ExpectQuery(qc).WillReturnError(fmt.Errorf("could not get count"))
	_, err = rw.Search(search, order)
	if err == nil {
		t.Error(expectedErr)
	}

	// failure2
	cr = sqlmock.NewRows([]string{"count(id)"}).AddRow(1)
	mock.ExpectQuery(qc).WillReturnRows(cr)
	mock.ExpectQuery(queryDocs).WillReturnError(fmt.Errorf("could not get documents"))
	_, err = rw.Search(search, order)
	if err == nil {
		t.Error(expectedErr)
	}

	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf(expectations, err)
	}
}

func TestSearchLists(t *testing.T) {
	var err error
	dbx, db, mock := getDbx(t)
	defer db.Close()
	c := shared.NewFromDB(dbx)
	rw := dbRepository{c}
	q := "SELECT distinct\\(taglist\\) as search FROM DOCUMENTS"
	columns := []string{"search"}
	searchErr := "error searching: %v"

	// multiple
	mock.ExpectQuery(q).WillReturnRows(sqlmock.NewRows(columns).AddRow("tag1").AddRow("tag2").AddRow("tag1;tag3"))
	tags, err := rw.SearchLists("tag", TAGS)
	if err != nil {
		t.Errorf(searchErr, err)
	}

	if len(tags) != 3 {
		t.Errorf("3 entries expected, got %d", len(tags))
	}

	if tags[0] != "tag1" || tags[1] != "tag2" || tags[2] != "tag3" {
		t.Errorf("wrong tags returned from search")
	}

	// single
	mock.ExpectQuery(q).WillReturnRows(sqlmock.NewRows(columns).AddRow("tag1").AddRow("tag2").AddRow("tag1;tag3"))
	tags, err = rw.SearchLists("tag2", TAGS)
	if err != nil {
		t.Errorf(searchErr, err)
	}

	if len(tags) != 1 {
		t.Errorf("1 entries expected, got %d", len(tags))
	}

	if tags[0] != "tag2" {
		t.Errorf("wrong tags returned from search")
	}

	// error1
	mock.ExpectQuery(q).WillReturnError(Err)
	_, err = rw.SearchLists("tag2", TAGS)
	if err == nil {
		t.Error(expectedErr)
	}

	// multiple
	mock.ExpectQuery("SELECT distinct\\(senderlist\\) as search FROM DOCUMENTS").WillReturnRows(sqlmock.NewRows(columns).AddRow("sender1").AddRow("sender2").AddRow("sender1;sender3"))
	senders, err := rw.SearchLists("sender", SENDERS)
	if err != nil {
		t.Errorf(searchErr, err)
	}

	if len(senders) != 3 {
		t.Errorf("3 entries expected, got %d", len(senders))
	}

	if senders[0] != "sender1" || senders[1] != "sender2" || senders[2] != "sender3" {
		t.Errorf("wrong senders returned from search")
	}

	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf(expectations, err)
	}

}
