package store_test

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.binggl.net/monorepo/internal/bookmarks/app/store"
	"golang.binggl.net/monorepo/pkg/persistence"
)

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
const mimeType = "application/pdf"

func fileRepo(t *testing.T) (store.FileRepository, store.BookmarkRepository, *sql.DB) {
	var (
		err error
	)
	dbCon := persistence.MustCreateSqliteConn(":memory:")
	con, err := persistence.CreateGormSqliteCon(dbCon)
	if err != nil {
		t.Fatalf("cannot create database connection: %v", err)
	}
	// Migrate the schema
	con.W().AutoMigrate(&store.File{}, &store.Bookmark{})
	db, err := con.W().DB()
	if err != nil {
		t.Fatalf("could not get DB handle; %v", err)
	}
	// use the write connection for both read/write to prevent issues with in-memory sqlite
	con.Read = con.Write
	fileRepo := store.CreateFileRepo(con, logger)
	bmRepo := store.CreateBookmarkRepo(con, logger)
	return fileRepo, bmRepo, db
}

func Test_Create_And_Get_File(t *testing.T) {
	repo, _, db := fileRepo(t)
	defer db.Close()

	// create
	file, err := repo.Save(store.File{
		Name:     "test.pdf",
		MimeType: mimeType,
		Payload:  []byte(pdfPayload),
		Size:     len([]byte(pdfPayload)),
	})
	if err != nil {
		t.Errorf("could not save a file %v", err)
	}
	if file.Payload == nil || file.Size == 0 {
		t.Errorf("the file was not saved correctly %v", err)
	}

	// fetch it
	file, err = repo.Get(file.ID)
	if err != nil {
		t.Errorf("could not get file by id: %v", err)
	}

	assert.Equal(t, file.Size, len(file.Payload))
	assert.Equal(t, "test.pdf", file.Name)
	assert.Equal(t, "application/pdf", file.MimeType)

	// change it
	file.Name = "changed.pdf"
	file, err = repo.Save(file)
	if err != nil {
		t.Errorf("could not save a file %v", err)
	}
	assert.Equal(t, "changed.pdf", file.Name)

	// input checks
	_, err = repo.Save(store.File{
		ID: "test",
	})
	assert.NotNil(t, err)

	_, err = repo.Save(store.File{
		ID:      "test",
		Payload: []byte(pdfPayload),
		Size:    0,
	})
	assert.NotNil(t, err)
}

func Test_Bookmark_Referencing_File(t *testing.T) {
	fileRepo, bmRepo, db := fileRepo(t)
	defer db.Close()

	// create
	file, err := fileRepo.Save(store.File{
		Name:     "test.pdf",
		MimeType: mimeType,
		Payload:  []byte(pdfPayload),
		Size:     len([]byte(pdfPayload)),
	})
	if err != nil {
		t.Errorf("could not save a file %v", err)
	}
	if file.Payload == nil || file.Size == 0 {
		t.Errorf("the file was not saved correctly %v", err)
	}

	userName := "username"
	item := store.Bookmark{
		DisplayName:        "displayName",
		Path:               "/",
		SortOrder:          0,
		Type:               store.FileItem,
		UserName:           userName,
		InvertFaviconColor: 1,
		File:               &file,
	}
	savedBm, err := bmRepo.Create(item)
	if err != nil {
		t.Errorf("Could not create bookmarks: %v", err)
	}

	// retrieve the bookmark
	bm, err := bmRepo.GetBookmarkByID(savedBm.ID, userName)
	if err != nil {
		t.Errorf("Could not get bookmarks by ID: %v", err)
	}
	assert.Equal(t, "displayName", bm.DisplayName)

	// delete the file
	err = fileRepo.Delete(file)
	if err != nil {
		t.Errorf("Could not delete file by ID: %v", err)
	}

	// retrieve bookmark again, the file should be nil
	bm, err = bmRepo.GetBookmarkByID(savedBm.ID, userName)
	if err != nil {
		t.Errorf("Could not get bookmarks by ID: %v", err)
	}
	assert.Equal(t, "displayName", bm.DisplayName)

}

func Test_FileRepo_InUnitOfWork(t *testing.T) {
	r, _, db := fileRepo(t)
	defer db.Close()
	err := r.InUnitOfWork(func(repo store.FileRepository) error {
		file, err := repo.Save(store.File{
			Name:     "test.pdf",
			MimeType: mimeType,
			Payload:  []byte(pdfPayload),
			Size:     len([]byte(pdfPayload)),
		})
		if err != nil {
			return err
		}

		assert.Equal(t, file.Size, len(file.Payload))
		assert.Equal(t, "test.pdf", file.Name)
		assert.Equal(t, "application/pdf", file.MimeType)

		return nil
	})
	if err != nil {
		t.Errorf("could not execute in UnitOfWork; %v", err)
	}

}
