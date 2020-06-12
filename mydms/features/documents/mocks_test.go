package documents

import (
	"database/sql"
	"fmt"
	"time"

	"golang.binggl.net/monorepo/mydms/features/filestore"
	"golang.binggl.net/monorepo/mydms/features/upload"
	"golang.binggl.net/monorepo/mydms/persistence"
)

var errTx = fmt.Errorf("start transaction failed")

// --------------------------------------------------------------------------
// MOCK: documents.Repository
// --------------------------------------------------------------------------

var _ Repository = (*mockRepository)(nil)

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

func (m *mockRepository) Get(id string) (d DocumentEntity, err error) {
	m.callCount++
	if id == "" {
		return DocumentEntity{}, fmt.Errorf("no document")
	}
	return DocumentEntity{
		Modified:    sql.NullTime{Time: time.Now().UTC(), Valid: true},
		PreviewLink: sql.NullString{String: "string", Valid: true},
	}, m.errMap[m.callCount]
}

func (m *mockRepository) Save(doc DocumentEntity, a persistence.Atomic) (d DocumentEntity, err error) {
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

func (m *mockRepository) Search(s DocSearch, order []OrderBy) (PagedDocuments, error) {
	m.callCount++
	if s.Title == noResult {
		return PagedDocuments{}, fmt.Errorf("search error")
	}

	return PagedDocuments{
		Count: 2,
		Documents: []DocumentEntity{
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

func (m *mockRepository) SearchLists(s string, st SearchType) ([]string, error) {
	if m.fail {
		return nil, Err
	}
	return []string{"one", "two"}, nil
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

// SaveFile(file FileItem) error
// GetFile(filePath string) (FileItem, error)
// DeleteFile(filePath string) error

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
// MOCK: upload.Repository
// --------------------------------------------------------------------------

type mockUploadRepository struct {
	c         persistence.Connection
	errMap    map[int]error
	resultMap map[int]upload.Upload
	callCount int
}

func newUploadRepo() *mockUploadRepository {
	return &mockUploadRepository{
		errMap:    make(map[int]error),
		resultMap: make(map[int]upload.Upload),
	}
}

// persistence.BaseRepository
// Write(item Upload, a persistence.Atomic) (err error)
// Read(id string) (Upload, error)
// Delete(id string, a persistence.Atomic) (err error)

func (m *mockUploadRepository) Write(item upload.Upload, a persistence.Atomic) (err error) {
	m.callCount++
	return m.errMap[m.callCount]
}
func (m *mockUploadRepository) Read(id string) (upload.Upload, error) {
	m.callCount++
	return m.resultMap[m.callCount], m.errMap[m.callCount]
}
func (m *mockUploadRepository) Delete(id string, a persistence.Atomic) (err error) {
	m.callCount++
	return m.errMap[m.callCount]
}
func (m *mockUploadRepository) CreateAtomic() (persistence.Atomic, error) {
	m.callCount++
	if err := m.errMap[m.callCount]; err != nil {
		return persistence.Atomic{}, errTx
	}
	return m.c.CreateAtomic()
}
