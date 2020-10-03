package document

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"golang.binggl.net/monorepo/internal/mydms/app/shared"
	"golang.binggl.net/monorepo/pkg/persistence"
)

// DocEntity represents a record in the persistence store
type DocEntity struct {
	ID            string         `db:"id"`
	Title         string         `db:"title"`
	FileName      string         `db:"filename"`
	AltID         string         `db:"alternativeid"`
	PreviewLink   sql.NullString `db:"previewlink"`
	Amount        float32        `db:"amount"`
	Created       time.Time      `db:"created"`
	Modified      sql.NullTime   `db:"modified"`
	TagList       string         `db:"taglist"`
	SenderList    string         `db:"senderlist"`
	InvoiceNumber sql.NullString `db:"invoicenumber"`
}

// PagedDocuments wraps a list of documents and returns the total number of documents
type PagedDocuments struct {
	Documents []DocEntity
	Count     int
}

// SortDirection can either by ASC or DESC
type SortDirection uint

const (
	// ASC as ascending sort direction
	ASC SortDirection = iota
	// DESC is descending sort direction
	DESC
)

func (s SortDirection) String() string {
	str := ""
	switch s {
	case ASC:
		str = "ASC"
		break
	case DESC:
		str = "DESC"
		break
	}
	return str
}

// DocSearch is used to search for documents
type DocSearch struct {
	Title  string
	Tag    string
	Sender string
	From   time.Time
	Until  time.Time
	Limit  int
	Skip   int
}

// OrderBy is used to sort a result list
type OrderBy struct {
	Field string
	Order SortDirection
}

// Repository is the CRUD interface for documents in the persistence store
type Repository interface {
	shared.BaseRepository
	Get(id string) (d DocEntity, err error)
	Exists(id string, a persistence.Atomic) (filePath string, err error)
	Save(doc DocEntity, a persistence.Atomic) (d DocEntity, err error)
	Delete(id string, a persistence.Atomic) (err error)
	Search(s DocSearch, order []OrderBy) (PagedDocuments, error)
	SearchLists(s string, st SearchType) ([]string, error)
}

// compiler interface check
var _ Repository = (*dbRepository)(nil)

// NewRepository creates a new instance using an existing connection
func NewRepository(c persistence.Connection) (Repository, error) {
	if !c.Active {
		return nil, fmt.Errorf("no repository connection available")
	}
	return &dbRepository{c}, nil
}

type dbRepository struct {
	c persistence.Connection
}

// CreateAtomic returns a new atomic object
func (rw *dbRepository) CreateAtomic() (persistence.Atomic, error) {
	return rw.c.CreateAtomic()
}

// Save a document entry. Either create or update the entry, based on availability
// if a valid/active atomic object is supplied the transaction handling is done by the caller
// otherwise a new transaction is created for the scope of the method
func (rw *dbRepository) Save(doc DocEntity, a persistence.Atomic) (d DocEntity, err error) {
	var (
		atomic  *persistence.Atomic
		newEnty bool
		r       sql.Result
	)

	defer func() {
		err = persistence.HandleTX(!a.Active, atomic, err)
	}()

	if atomic, err = persistence.CheckTX(rw.c, &a); err != nil {
		return
	}

	// try to fetch a document if an ID is supplied
	// the supplied ID is checked against an existing item
	// if the item is not found the provided data is used to create a new entry
	newEnty = true
	if doc.ID != "" {
		var find DocEntity
		// use the database logic for row-locking to prevent issues concurrently updating entries
		err = rw.c.Get(&find, "SELECT id,title,filename,alternativeid,previewlink,amount,taglist,senderlist,created,modified,invoicenumber FROM DOCUMENTS WHERE id=? FOR UPDATE", doc.ID)
		if err != nil {
			log.Printf("could not get a Document by ID '%s' - a new entry will be created", doc.ID)
			newEnty = true
		} else {
			newEnty = false
			doc.Created = find.Created
		}
	}

	if newEnty {
		doc.ID = uuid.New().String()
		doc.Created = time.Now().UTC()
		doc.AltID = randomString(8)
		r, err = atomic.NamedExec("INSERT INTO DOCUMENTS (id,title,filename,alternativeid,previewlink,amount,taglist,senderlist,created,invoicenumber) VALUES (:id,:title,:filename,:alternativeid,:previewlink,:amount,:taglist,:senderlist,:created,:invoicenumber)", &doc)
	} else {
		m := sql.NullTime{Time: time.Now().UTC(), Valid: true}
		doc.Modified = m
		r, err = atomic.NamedExec("UPDATE DOCUMENTS SET title=:title,filename=:filename,alternativeid=:alternativeid,previewlink=:previewlink,amount=:amount,taglist=:taglist,senderlist=:senderlist,modified=:modified,invoicenumber=:invoicenumber WHERE id=:id", &doc)
	}

	if err != nil {
		err = fmt.Errorf("could not create new entry: %v", err)
		return
	}
	c, err := r.RowsAffected()
	if err != nil {
		err = fmt.Errorf("could not get affected rows: %v", err)
		return
	}
	if c != 1 {
		err = fmt.Errorf("invalid number of rows affected, got %d", c)
		return
	}

	return doc, nil
}

// Get retuns a document by the given id
func (rw *dbRepository) Get(id string) (d DocEntity, err error) {
	err = rw.c.Get(&d, "SELECT id,title,filename,alternativeid,previewlink,amount,taglist,senderlist,created,modified,invoicenumber FROM DOCUMENTS WHERE id=?", id)
	if err != nil {
		err = fmt.Errorf("cannot get document by id '%s': %v", id, err)
		return
	}
	return d, nil
}

// Exists checks if a given id is available
func (rw *dbRepository) Exists(id string, a persistence.Atomic) (filePath string, err error) {
	var (
		atomic *persistence.Atomic
	)

	defer func() {
		err = persistence.HandleTX(!a.Active, atomic, err)
	}()

	if atomic, err = persistence.CheckTX(rw.c, &a); err != nil {
		return
	}

	var filename string
	err = atomic.Get(&filename, "SELECT filename FROM DOCUMENTS WHERE id = ?", id)
	if err != nil {
		err = fmt.Errorf("cannot query document or document not available. %v", err)
		return
	}
	return filename, nil

}

// Delete a document by its id
func (rw *dbRepository) Delete(id string, a persistence.Atomic) (err error) {
	var (
		atomic *persistence.Atomic
	)

	defer func() {
		err = persistence.HandleTX(!a.Active, atomic, err)
	}()

	if atomic, err = persistence.CheckTX(rw.c, &a); err != nil {
		return
	}

	_, err = atomic.Exec("DELETE FROM DOCUMENTS WHERE id = ?", id)
	if err != nil {
		err = fmt.Errorf("cannot delete document item: %v", err)
	}
	return
}

// Search for documents based on the supplied search-object 'DocSearch'
// the slice of order-bys is used to defined the query sort-order
func (rw *dbRepository) Search(s DocSearch, order []OrderBy) (d PagedDocuments, err error) {
	var query string
	q := "SELECT id,title,filename,alternativeid,previewlink,amount,taglist,senderlist,created,modified,invoicenumber FROM DOCUMENTS"
	qc := "SELECT count(id) FROM DOCUMENTS"
	where := "\nWHERE 1=1"
	paging := ""
	orderby := orderBy(order)
	arg := make(map[string]interface{})

	// use the supplied search-object to create the query
	if s.Title != "" {
		where += "\nAND ( lower(title) LIKE :search OR lower(taglist) LIKE :search OR lower(senderlist) LIKE :search OR lower(invoicenumber) LIKE :search)"
		arg["search"] = "%" + strings.ToLower(s.Title) + "%"
	}
	if s.Tag != "" {
		where += "\nAND lower(taglist) LIKE :tag"
		arg["tag"] = "%" + strings.ToLower(s.Tag) + "%"
	}
	if s.Sender != "" {
		where += "\nAND lower(senderlist) LIKE :sender"
		arg["sender"] = "%" + strings.ToLower(s.Sender) + "%"
	}
	if !s.From.IsZero() {
		where += "\nAND created >= :from"
		arg["from"] = s.From
	}
	if !s.Until.IsZero() {
		where += "\nAND created <= :until"
		arg["until"] = s.Until
	}
	if s.Limit > 0 {
		paging += fmt.Sprintf("\nLIMIT %d", s.Limit)
	}
	if s.Skip > 0 {
		paging += fmt.Sprintf("\nOFFSET %d", s.Skip)
	}

	// get the number of affected documents
	query = qc + where
	var c int
	query, args, err := prepareQuery(rw.c, query, arg)
	if err != nil {
		return
	}

	if err = rw.c.Get(&c, query, args...); err != nil {
		err = fmt.Errorf("could not get the total number of documents: %v", err)
		return
	}

	// retrieve the documents
	query = q + where + orderby + paging
	log.Printf("QUERY: %s", query)
	query, args, err = prepareQuery(rw.c, query, arg)
	if err != nil {
		return
	}
	var docs []DocEntity
	if err = rw.c.Select(&docs, query, args...); err != nil {
		err = fmt.Errorf("could not get the documents: %v", err)
		return
	}
	return PagedDocuments{Documents: docs, Count: c}, nil
}

// SearchType is used to determine if the search is performend on tags or senders
type SearchType uint

const (
	// TAGS is used to search tags within the documents table
	TAGS SearchType = iota
	// SENDERS is used to search senders within the documents table
	SENDERS
)

// SearchLists collects all tag-entries from all documents and returns those elements which start with
// the given search term. The search is performed case insensitive
func (rw *dbRepository) SearchLists(s string, st SearchType) ([]string, error) {
	var (
		t      string
		result []string
		lookup map[string]int
		found  []string
	)

	search := make(map[SearchType]string)
	search[TAGS] = "taglist"
	search[SENDERS] = "senderlist"

	query := "SELECT distinct(%s) as search FROM DOCUMENTS WHERE lower(%s) LIKE ?"
	query = fmt.Sprintf(query, search[st], search[st])

	rows, err := rw.c.Queryx(query, "%"+strings.ToLower(s)+"%")
	if err != nil {
		err = fmt.Errorf("could not search for %s: %v", search[st], err)
		return nil, err
	}
	defer rows.Close()

	lookup = make(map[string]int)
	for rows.Next() {
		if err := rows.Scan(&t); err != nil {
			err = fmt.Errorf("could not get row contents: %v", err)
			return nil, err
		}
		parts := strings.Split(t, ";")
		for _, p := range parts {
			if _, found := lookup[p]; !found && p != "" {
				lookup[p] = 1
				result = append(result, p)
			}
		}
	}

	// now we have collected all search-elements of all documents
	// search for those which start with the given search term s
	s = strings.ToLower(s)
	for i := range result {
		if strings.HasPrefix(strings.ToLower(result[i]), s) {
			found = append(found, result[i])
		}
	}
	sort.Strings(found)
	return found, nil
}

func prepareQuery(c persistence.Connection, q string, args map[string]interface{}) (string, []interface{}, error) {
	namedq, namedargs, err := sqlx.Named(q, args)
	if err != nil {
		return "", nil, fmt.Errorf("query error: %v", err)
	}
	query := c.Rebind(namedq)
	return query, namedargs, nil
}

// found: https://www.admfactory.com/how-to-generate-a-fixed-length-random-string-using-golang/
func randomString(n int) string {
	var letter = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	b := make([]rune, n)
	for i := range b {
		b[i] = letter[rand.Intn(len(letter))]
	}
	return string(b)
}

func orderBy(order []OrderBy) string {
	orderby := ""
	if order != nil && len(order) > 0 {
		orderby = "\nORDER BY "
		for i, o := range order {
			if i > 0 {
				orderby += ", "
			}
			orderby += fmt.Sprintf("%s %s", o.Field, o.Order)
		}
	}
	return orderby
}
