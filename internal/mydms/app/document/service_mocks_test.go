package document_test

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"golang.binggl.net/monorepo/internal/mydms/app/document"
	"golang.binggl.net/monorepo/internal/mydms/app/filestore"
	"golang.binggl.net/monorepo/internal/mydms/app/upload"
	"golang.binggl.net/monorepo/pkg/persistence"
)

var errTx = fmt.Errorf("start transaction failed")
var Err = fmt.Errorf("error")

const notExists = "!exists"
const noDelete = "!delete"
const noResult = "!result"
const noFileDelete = "!fileDelete"

// rather small PDF payload
// https://stackoverflow.com/questions/17279712/what-is-the-smallest-possible-valid-pdf
const pdfPayload = `%PDF-1.0
1 0 obj<</Type/Catalog/Pages 2 0 R>>endobj 2 0 obj<</Type/Pages/Kids[3 0 R]/Count 1>>endobj 3 0 obj<</Type/Page/MediaBox[0 0 3 3]>>endobj
xref
0 4
0000000000 65535 f
0000000010 00000 n
0000000053 00000 n
0000000102 00000 n
trailer<</Size 4/Root 1 0 R>>
startxref
149
%EOF
`

// --------------------------------------------------------------------------
// MOCK: documents.Repository
// --------------------------------------------------------------------------

var _ document.Repository = (*mockRepository)(nil)

type mockRepository struct {
	c         persistence.Connection
	fail      bool
	errMap    map[int]error
	callCount int
}

func newDocRepo(c persistence.Connection) *mockRepository {
	return &mockRepository{
		c:      c,
		errMap: make(map[int]error),
	}
}

func (m *mockRepository) Get(id string) (d document.DocEntity, err error) {
	m.callCount++
	if id == "" {
		return document.DocEntity{}, fmt.Errorf("no document")
	}
	return document.DocEntity{
		ID:          id,
		Modified:    sql.NullTime{Time: time.Now().UTC(), Valid: true},
		PreviewLink: sql.NullString{String: "string", Valid: true},
	}, m.errMap[m.callCount]
}

func (m *mockRepository) Save(doc document.DocEntity, a persistence.Atomic) (d document.DocEntity, err error) {
	m.callCount++
	return doc, m.errMap[m.callCount]
}

func (m *mockRepository) Delete(id string, a persistence.Atomic) (err error) {
	m.callCount++
	if id == noDelete {
		return fmt.Errorf("delete error")
	}
	return m.errMap[m.callCount]
}

func (m *mockRepository) Search(s document.DocSearch, order []document.OrderBy) (document.PagedDocResult, error) {
	m.callCount++
	if s.Title == noResult {
		return document.PagedDocResult{}, fmt.Errorf("search error")
	}

	return document.PagedDocResult{
		Count: 2,
		Documents: []document.DocEntity{
			{
				Title:       "title1",
				FileName:    "filename1",
				Amount:      1,
				TagList:     "taglist1",
				SenderList:  "senderlist1",
				PreviewLink: sql.NullString{String: "previewlink", Valid: true},
				Created:     time.Now().UTC(),
				Modified:    sql.NullTime{Time: time.Now().UTC(), Valid: true},
			},
			{
				Title:      "title2",
				FileName:   "filename2",
				Amount:     2,
				TagList:    "taglist2",
				SenderList: "senderlist2",
				Created:    time.Now().UTC(),
			},
		},
	}, nil
}

func (m *mockRepository) Exists(id string, a persistence.Atomic) (filePath string, err error) {
	m.callCount++
	if id == notExists {
		return "", fmt.Errorf("exists error")
	}
	if id == noFileDelete {
		return noFileDelete, nil
	}
	return "file", nil
}

func (m *mockRepository) CreateAtomic() (persistence.Atomic, error) {
	m.callCount++
	if m.fail {
		return persistence.Atomic{}, errTx
	}
	return m.c.CreateAtomic()
}

func (m *mockRepository) SearchLists(s string, st document.SearchType) ([]string, error) {
	if m.fail {
		return nil, Err
	}
	return []string{"one", "two"}, nil
}

func GetMockConn(t *testing.T) (persistence.Connection, *sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("could not create a new SQL mock; %v", err)
	}

	dbx := sqlx.NewDb(db, "mysql")
	con := persistence.NewFromDB(dbx)

	return con, db, mock
}

// --------------------------------------------------------------------------
// MOCK: filestore.FileService
// --------------------------------------------------------------------------

type mockFileService struct {
	errMap    map[int]error
	callCount int
}

func newFileService() *mockFileService {
	return &mockFileService{
		errMap: make(map[int]error),
	}
}

// compile time assertions for our service
var (
	_ filestore.FileService = &mockFileService{}
)

func (m *mockFileService) InitClient() (err error) {
	m.callCount++
	return m.errMap[m.callCount]
}

func (m *mockFileService) SaveFile(file filestore.FileItem) error {
	m.callCount++
	return m.errMap[m.callCount]
}

func (m *mockFileService) GetFile(filePath string) (filestore.FileItem, error) {
	m.callCount++
	return filestore.FileItem{
		FileName:   "test.pdf",
		FolderName: "PATH",
		MimeType:   "application/pdf",
		Payload:    []byte(pdfPayload),
	}, m.errMap[m.callCount]
}

func (m *mockFileService) DeleteFile(filePath string) error {
	m.callCount++
	return m.errMap[m.callCount]
}

// --------------------------------------------------------------------------
// MOCK: upload.Client
// --------------------------------------------------------------------------

type mockUploadClient struct {
	getFail    bool
	deleteFail bool
}

func (m *mockUploadClient) Get(id, authToken string) (upload.Upload, error) {
	if m.getFail {
		return upload.Upload{}, fmt.Errorf("error")
	}
	return upload.Upload{}, nil
}

// Delete removes the upload-object specified by the given ID
func (m *mockUploadClient) Delete(id, authToken string) error {
	if m.deleteFail {
		return fmt.Errorf("error")
	}
	return nil
}

var _ upload.Client = &mockUploadClient{}
