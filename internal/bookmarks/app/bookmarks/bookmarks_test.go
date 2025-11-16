package bookmarks_test

import (
	"bytes"
	_ "embed"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"golang.binggl.net/monorepo/internal/bookmarks/app/bookmarks"
	"golang.binggl.net/monorepo/internal/bookmarks/app/store"
	"golang.binggl.net/monorepo/internal/common/upload"
	"golang.binggl.net/monorepo/pkg/logging"
	"golang.binggl.net/monorepo/pkg/persistence"
	"golang.binggl.net/monorepo/pkg/security"
)

const faviconPath = "/tmp"
const maxUploadSize = 5000000

var allowedFileTypes = []string{"pdf"}

var user = security.User{
	Username:    "test",
	Email:       "a.b@c.de",
	DisplayName: "Test",
}

type mockUploadStore struct {
	fail   bool
	upload upload.Upload
}

func (m *mockUploadStore) Write(item upload.Upload) (err error) {
	if m.fail {
		return fmt.Errorf("error")
	}
	m.upload = item
	return nil
}

func (m *mockUploadStore) Read(id string) (upload.Upload, error) {
	if m.fail {
		return upload.Upload{}, fmt.Errorf("error")
	}
	if id == m.upload.ID {
		return m.upload, nil
	}
	return upload.Upload{}, fmt.Errorf("error")
}

func (m *mockUploadStore) Delete(id string) (err error) {
	if m.fail {
		return fmt.Errorf("error")
	}
	m.upload = upload.Upload{}
	return nil
}

var _ upload.Store = &mockUploadStore{}

func app(t *testing.T) bookmarks.Application {
	bRepo, fRepo, fileRepo := repositories(t)
	uploadSvc := upload.NewService(upload.ServiceOptions{
		Logger:           logging.NewNop(),
		Crypter:          nil,
		MaxUploadSize:    maxUploadSize,
		AllowedFileTypes: allowedFileTypes,
		Store:            &mockUploadStore{},
	})
	return bookmarks.Application{
		BookmarkStore: bRepo,
		FavStore:      fRepo,
		Logger:        logging.NewNop(),
		FaviconPath:   faviconPath,
		FileStore:     fileRepo,
		UploadSvc:     uploadSvc,
	}
}

func repositories(t *testing.T) (store.BookmarkRepository, store.FaviconRepository, store.FileRepository) {
	var (
		err error
	)
	dbCon := persistence.MustCreateSqliteConn(":memory:")
	con, err := persistence.CreateGormSqliteCon(dbCon)
	if err != nil {
		t.Fatalf("cannot create database connection: %v", err)
	}
	// Migrate the schema
	con.W().AutoMigrate(&store.Bookmark{}, &store.Favicon{}, &store.File{})
	con.Read = con.Write
	logger := logging.NewNop()
	return store.CreateBookmarkRepo(con, logger), store.CreateFaviconRepo(con, logger), store.CreateFileRepo(con, logger)
}

func Test_GetBookmark_NotFound(t *testing.T) {
	svc := app(t)

	_, err := svc.GetBookmarkByID("", security.User{})
	if err == nil {
		t.Errorf("error expected if no id supplied")
	}

	_, err = svc.GetBookmarkByID("not-found-id", security.User{})
	if err == nil {
		t.Errorf("error expected if no id supplied")
	}
}

func Test_CreateBookmark(t *testing.T) {
	svc := app(t)

	folder, err := svc.CreateBookmark(bookmarks.Bookmark{
		Type:        bookmarks.Folder,
		Path:        "/",
		DisplayName: "a",
	}, user)
	if err != nil {
		t.Errorf("could not create bookmark; %v", err)
	}

	bm, err := svc.CreateBookmark(bookmarks.Bookmark{
		Type:               bookmarks.Node,
		Path:               "/" + folder.DisplayName,
		DisplayName:        "test",
		URL:                "http://www.example.com",
		InvertFaviconColor: 1,
	}, user)
	if err != nil {
		t.Errorf("could not create bookmark; %v", err)
	}

	fetched, err := svc.GetBookmarkByID(bm.ID, user)
	if err != nil {
		t.Errorf("could not fetch bookmark; %v", err)
	}

	if fetched.DisplayName != bm.DisplayName {
		t.Errorf("the fetched bookmark is not the requested one")
	}
	if fetched.InvertFaviconColor != bm.InvertFaviconColor {
		t.Errorf("the fetched bookmark does not have the flag InvertFaviconColor set")
	}

	// ---- error case ----

	// missing path
	_, err = svc.CreateBookmark(bookmarks.Bookmark{
		Type:        bookmarks.Folder,
		Path:        "",
		DisplayName: "a",
	}, user)
	if err == nil {
		t.Errorf("error expected")
	}

	// missing name
	_, err = svc.CreateBookmark(bookmarks.Bookmark{
		Type:        bookmarks.Folder,
		Path:        "/",
		DisplayName: "",
	}, user)
	if err == nil {
		t.Errorf("error expected")
	}

	// missing URL
	_, err = svc.CreateBookmark(bookmarks.Bookmark{
		Type:        bookmarks.Node,
		Path:        "/",
		DisplayName: "test",
	}, user)
	if err == nil {
		t.Errorf("error expected")
	}

	// missing File
	_, err = svc.CreateBookmark(bookmarks.Bookmark{
		Type:        bookmarks.FileItem,
		Path:        "/",
		DisplayName: "test",
	}, user)
	if err == nil {
		t.Errorf("error expected")
	}
}

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

func Test_CreateBookmark_File(t *testing.T) {
	svc := app(t)

	fileID, err := svc.UploadSvc.Save(upload.File{
		Name:     "test.pdf",
		MimeType: "application/pdf",
		File:     bytes.NewReader([]byte(pdfPayload)),
		Size:     int64(len([]byte(pdfPayload))),
	})
	if fileID == "" {
		t.Errorf("could not save file; %v", err)
	}

	bm, err := svc.CreateBookmark(bookmarks.Bookmark{
		Type:        bookmarks.FileItem,
		Path:        "/",
		DisplayName: "test",
		FileID:      fileID,
	}, user)
	if err != nil {
		t.Errorf("could not create bookmark; %v", err)
	}

	if bm.Type != bookmarks.FileItem {
		t.Errorf("the type of the created item is wrong, expected '%v', got '%v'", bookmarks.FileItem, bm.Type)
	}
	if bm.FileID == "" {
		t.Errorf("the bookmark is missing a file")
	}

	// retrieve the bookmark again
	fetched, err := svc.GetBookmarkByID(bm.ID, user)
	if err != nil {
		t.Errorf("could not fetch bookmark; %v", err)
	}

	if fetched.FileID == "" {
		t.Errorf("the fetched bookmark is missing a fileID")
	}
	if fetched.FilePayload.Payload == nil {
		t.Errorf("the fetched bookmark is missing a file payload")
	}

	// ---- error case ----

	// missing path
	_, err = svc.CreateBookmark(bookmarks.Bookmark{
		Type:        bookmarks.Node,
		Path:        "/",
		DisplayName: "a",
		URL:         "https://abc.de",
		FileID:      "a file was supplied",
	}, user)
	if err == nil {
		t.Errorf("error expected")
	}
}

func Test_GetBookmarksByPath(t *testing.T) {
	svc := app(t)
	// create a bookmark path and retrieve the bookmarks of the given path
	// Path: /a
	path := uuid.NewString()
	svc.CreateBookmark(bookmarks.Bookmark{
		Type:        bookmarks.Folder,
		DisplayName: path,
		Path:        "/",
	}, user)
	// create two nodes
	svc.CreateBookmark(bookmarks.Bookmark{
		Type:        bookmarks.Node,
		DisplayName: "test1",
		Path:        "/" + path,
		URL:         "http://localhost",
	}, user)
	svc.CreateBookmark(bookmarks.Bookmark{
		Type:        bookmarks.Node,
		DisplayName: "test2",
		Path:        "/" + path,
		URL:         "http://localhost",
	}, user)

	// ---- fetch the nodes by path ----
	bms, err := svc.GetBookmarksByPath("/"+path, user)
	if err != nil {
		t.Errorf("could not fetch bookmarks by path; %v", err)
	}
	if len(bms) != 2 {
		t.Error("expected to get two bookmarks in return")
	}

	// ---- error case ----
	_, err = svc.GetBookmarksByPath("", user)
	if err == nil {
		t.Error("error expected")
	}
	bms, _ = svc.GetBookmarksByPath("we-will-not-find-this-path", user)
	if len(bms) != 0 {
		t.Error("there should be no bookmarks found by path")
	}
}

func Test_GetBookmarksFolderByPath(t *testing.T) {
	svc := app(t)
	// create a bookmark path and retrieve the bookmarks of the given path
	// Path: /folder/folder1
	folder := uuid.NewString()
	svc.CreateBookmark(bookmarks.Bookmark{
		Type:        bookmarks.Folder,
		DisplayName: folder,
		Path:        "/",
	}, user)
	folder1 := uuid.NewString()
	svc.CreateBookmark(bookmarks.Bookmark{
		Type:        bookmarks.Folder,
		DisplayName: folder1,
		Path:        "/" + folder,
	}, user)

	// get the folder by path
	f, err := svc.GetBookmarksFolderByPath("/"+folder+"/"+folder1, user)
	if err != nil {
		t.Errorf("did not expect error for GetBookmarksFolderByPath; %v", err)
	}
	if f.DisplayName != folder1 {
		t.Errorf("wrong folder returned; %s", f.DisplayName)
	}

	// ---- ROOT ----
	f, err = svc.GetBookmarksFolderByPath("/", user)
	if err != nil {
		t.Errorf("did not expect error for GetBookmarksFolderByPath; %v", err)
	}
	if f.DisplayName != "Root" {
		t.Errorf("wrong folder returned; expected Root - got %s", f.DisplayName)
	}

	// ---- error ----
	_, err = svc.GetBookmarksFolderByPath("", user)
	if err == nil {
		t.Error("error expected for empty path")
	}

	_, err = svc.GetBookmarksFolderByPath("/unknown", user)
	if err == nil {
		t.Error("error expected for unknown path")
	}
}

func Test_GetAllPaths(t *testing.T) {
	svc := app(t)
	// /folder/folder1
	// /folder/folder2
	// /folder3
	folder := uuid.NewString()
	svc.CreateBookmark(bookmarks.Bookmark{
		Type:        bookmarks.Folder,
		DisplayName: folder,
		Path:        "/",
	}, user)
	folder1 := uuid.NewString()
	svc.CreateBookmark(bookmarks.Bookmark{
		Type:        bookmarks.Folder,
		DisplayName: folder1,
		Path:        "/" + folder,
	}, user)
	folder2 := uuid.NewString()
	svc.CreateBookmark(bookmarks.Bookmark{
		Type:        bookmarks.Folder,
		DisplayName: folder2,
		Path:        "/" + folder,
	}, user)
	folder3 := uuid.NewString()
	svc.CreateBookmark(bookmarks.Bookmark{
		Type:        bookmarks.Folder,
		DisplayName: folder3,
		Path:        "/",
	}, user)

	paths, err := svc.GetAllPaths(user)
	if err != nil {
		t.Errorf("did not expect error for GetAllPaths; %v", err)
	}
	if len(paths) < 5 {
		t.Errorf("expected number of entries but got %d", len(paths))
	}

	// ---- error ----
	paths, _ = svc.GetAllPaths(security.User{Username: "unknown"})
	if len(paths) != 1 { // we always get "/"
		t.Errorf("expected to get no results for unknown user - got %d!", len(paths))
	}
}

func Test_GetBookmarksByName(t *testing.T) {
	svc := app(t)
	// /folder
	// /folder/$aaa
	// /folder/$bbb
	// /folder/$aab
	folder := uuid.NewString()
	svc.CreateBookmark(bookmarks.Bookmark{
		Type:        bookmarks.Folder,
		DisplayName: folder,
		Path:        "/",
	}, user)
	aaa := uuid.NewString()
	svc.CreateBookmark(bookmarks.Bookmark{
		Type:        bookmarks.Node,
		DisplayName: "$aaa" + aaa,
		URL:         "http://www.aaa.com",
		Path:        "/" + folder,
	}, user)
	bbb := uuid.NewString()
	svc.CreateBookmark(bookmarks.Bookmark{
		Type:        bookmarks.Node,
		DisplayName: "$bbb" + bbb,
		URL:         "http://www.bbb.com",
		Path:        "/" + folder,
	}, user)
	aab := uuid.NewString()
	svc.CreateBookmark(bookmarks.Bookmark{
		Type:        bookmarks.Node,
		DisplayName: "$aab" + aab,
		URL:         "http://www.aab.com",
		Path:        "/" + folder,
	}, user)

	bms, err := svc.GetBookmarksByName("$aa", user)
	if err != nil {
		t.Errorf("did not expect an error for GetBookmarksByName; %v", err)
	}
	if len(bms) != 2 {
		t.Errorf("expected 2 entries but got %d", len(bms))
	}
	bms, err = svc.GetBookmarksByName("$bbb", user)
	if err != nil {
		t.Errorf("did not expect an error for GetBookmarksByName; %v", err)
	}
	if len(bms) != 1 {
		t.Errorf("expected 1 entries but got %d", len(bms))
	}

	// ---- error ----
	_, err = svc.GetBookmarksByName("", user)
	if err == nil {
		t.Errorf("expected error for empty name")
	}

}

func Test_FetchFavicon(t *testing.T) {
	svc := app(t)
	server := hostTestWebserver(t)
	defer server.Close()

	obj, err := svc.LocalFetchFaviconURL(server.URL + "/Wikipedia-logo.png")
	if err != nil {
		t.Errorf("error fetching favicon: %v", err)
	}
	assert.True(t, len(obj.Payload) > 0)

	obj, err = svc.LocalExtractFaviconFromURL(server.URL + "/orf.at")
	if err != nil {
		t.Errorf("error fetching favicon: %v", err)
	}
	assert.True(t, len(obj.Payload) > 0)

	favicon, err := svc.GetLocalFaviconByID(obj.Name)
	if err != nil {
		t.Errorf("could not get local favicon; %v", err)
	}
	assert.True(t, len(favicon.Payload) > 0)

	svc.RemoveLocalFavicon(obj.Name)
	_, err = os.Stat(path.Join(faviconPath, obj.Name))
	if err == nil {
		t.Errorf("the file should have been deleted")
	}

	// ---- validation ----

	_, err = svc.GetLocalFaviconByID("")
	if err == nil {
		t.Errorf("error expected")
	}

	favicon, err = svc.GetLocalFaviconByID("anyid")
	if err != nil {
		t.Errorf("should return the default favicon")
	}
	assert.True(t, len(favicon.Payload) > 0)

	_, err = svc.LocalFetchFaviconURL("")
	if err == nil {
		t.Errorf("error expected")
	}

	_, err = svc.LocalExtractFaviconFromURL("")
	if err == nil {
		t.Errorf("error expected")
	}

	err = svc.RemoveLocalFavicon("")
	if err == nil {
		t.Errorf("error expected")
	}

}

func Test_FetchBookmark(t *testing.T) {
	svc := app(t)
	// /folder/bookmark

	folder := uuid.NewString()
	f, _ := svc.CreateBookmark(bookmarks.Bookmark{
		Type:        bookmarks.Folder,
		DisplayName: folder,
		Path:        "/",
	}, user)

	url := "http://www.example.com"
	bookmark := uuid.NewString()
	bm, err := svc.CreateBookmark(bookmarks.Bookmark{
		Type:        bookmarks.Node,
		DisplayName: bookmark,
		URL:         url,
		Path:        "/" + folder,
		Highlight:   1,
	}, user)

	if err != nil {
		t.Errorf("could not create a bookmark; got error; %v", err)
	}

	// fetch-and-forward
	bmURL, err := svc.FetchAndForward(bm.ID, user)
	if err != nil {
		t.Errorf("should not return an error for fetching ID: %v", err)
	}
	if bmURL != url {
		t.Errorf("expected the url %s - got url %s", url, bmURL)
	}

	bm, _ = svc.GetBookmarkByID(bm.ID, user)
	if bm.Highlight == 1 {
		t.Errorf("expected the highlight flag to be 0 after a fetch")
	}

	// ---- error ----

	_, err = svc.FetchAndForward("", user)
	if err == nil {
		t.Errorf("expected an error when empty ID supplied")
	}

	_, err = svc.FetchAndForward("unknown-id", user)
	if err == nil {
		t.Errorf("expected an error for unknown ID supplied")
	}

	_, err = svc.FetchAndForward(f.ID, user)
	if err == nil {
		t.Errorf("expected an error for supplied folder ID")
	}
}

func Test_GetFavicon(t *testing.T) {
	svc := app(t)
	server := hostTestWebserver(t)
	defer server.Close()

	obj, err := svc.LocalFetchFaviconURL(server.URL + "/Wikipedia-logo.png")
	if err != nil {
		t.Errorf("error fetching favicon: %v", err)
	}

	// /folder/bookmark

	folder := uuid.NewString()
	svc.CreateBookmark(bookmarks.Bookmark{
		Type:        bookmarks.Folder,
		DisplayName: folder,
		Path:        "/",
	}, user)

	url := "http://localhost"
	bookmark := uuid.NewString()
	bm, err := svc.CreateBookmark(bookmarks.Bookmark{
		Type:        bookmarks.Node,
		DisplayName: bookmark,
		URL:         url,
		Path:        "/" + folder,
		Favicon:     obj.Name,
	}, user)
	if err != nil {
		t.Errorf("could not store bookmark; %v", err)
	}

	favicon, err := svc.GetBookmarkFavicon(bm.ID, user)
	if err != nil {
		t.Errorf("could not get path of favicon for ID: %s; %v", bm.ID, err)
	}

	if len(favicon.Payload) == 0 {
		t.Errorf("expected a path for the faviconGetFaviconPath")
	}

	// ---- error ----

	_, err = svc.GetBookmarkFavicon("", user)
	if err == nil {
		t.Errorf("expected error for empty id")
	}

	_, err = svc.GetBookmarkFavicon("unknown-id", user)
	if err == nil {
		t.Errorf("expected error for unknown id")
	}
}

func Test_DeleteBookmark(t *testing.T) {
	svc := app(t)
	server := hostTestWebserver(t)
	defer server.Close()

	obj, err := svc.LocalFetchFaviconURL(server.URL + "/Wikipedia-logo.png")
	if err != nil {
		t.Errorf("error fetching favicon: %v", err)
	}

	// /folder/bookmark

	folder := uuid.NewString()
	bmF, _ := svc.CreateBookmark(bookmarks.Bookmark{
		Type:        bookmarks.Folder,
		DisplayName: folder,
		Path:        "/",
	}, user)

	bookmark := uuid.NewString()
	bm, _ := svc.CreateBookmark(bookmarks.Bookmark{
		Type:        bookmarks.Node,
		DisplayName: bookmark,
		URL:         "http://localhost",
		Path:        "/" + folder,
		Favicon:     obj.Name,
	}, user)

	// try to delete the folder which should not work because of the child-node
	err = svc.Delete(bmF.ID, user)
	if err == nil {
		t.Error("expected an error because cannot delete a bookmark with child entries")
	}

	err = svc.Delete(bm.ID, user)
	if err != nil {
		t.Errorf("could not delete bookmark with ID; %s; %v", bm.ID, err)
	}

	err = svc.Delete(bmF.ID, user)
	if err != nil {
		t.Errorf("could not delete bookmark with ID; %s; %v", bmF.ID, err)
	}

	_, err = svc.GetBookmarkByID(bm.ID, user)
	if err == nil {
		t.Errorf("expected an error because deleted bookmark with id %s", bm.ID)
	}

	_, err = svc.GetBookmarkByID(bmF.ID, user)
	if err == nil {
		t.Errorf("expected an error because deleted bookmark with id %s", bmF.ID)
	}

	// ---- error ----

	err = svc.Delete("", user)
	if err == nil {
		t.Errorf("expected error for empty id")
	}

	err = svc.Delete("unknown-id", user)
	if err == nil {
		t.Errorf("expected error for unknown id")
	}
}

func Test_DeleteBookmarkPath(t *testing.T) {
	svc := app(t)
	server := hostTestWebserver(t)
	defer server.Close()

	obj, err := svc.LocalFetchFaviconURL(server.URL + "/Wikipedia-logo.png")
	if err != nil {
		t.Errorf("error fetching favicon: %v", err)
	}

	// /folder/bookmark

	folder := uuid.NewString()
	bmF, _ := svc.CreateBookmark(bookmarks.Bookmark{
		Type:        bookmarks.Folder,
		DisplayName: folder,
		Path:        "/",
	}, user)

	bookmark := uuid.NewString()
	bm, _ := svc.CreateBookmark(bookmarks.Bookmark{
		Type:        bookmarks.Node,
		DisplayName: bookmark,
		URL:         "http://localhost",
		Path:        "/" + folder,
		Favicon:     obj.Name,
	}, user)

	err = svc.DeletePath(bmF.ID, user)
	if err != nil {
		t.Error("could not delete the folder-path", err)
	}

	_, err = svc.GetBookmarkByID(bm.ID, user)
	if err == nil {
		t.Errorf("expected an error because deleted bookmark with id %s", bm.ID)
	}

	_, err = svc.GetBookmarkByID(bmF.ID, user)
	if err == nil {
		t.Errorf("expected an error because deleted bookmark with id %s", bmF.ID)
	}

	// ---- error ----

	err = svc.DeletePath("", user)
	if err == nil {
		t.Errorf("expected error for empty id")
	}

	err = svc.DeletePath("unknown-id", user)
	if err == nil {
		t.Errorf("expected error for unknown id")
	}

	// ---- Delete Item ----

	bookmark = uuid.NewString()
	bm, _ = svc.CreateBookmark(bookmarks.Bookmark{
		Type:        bookmarks.Node,
		DisplayName: bookmark,
		URL:         "http://localhost",
		Path:        "/",
	}, user)
	err = svc.DeletePath(bm.ID, user)
	if err == nil {
		t.Errorf("expected error for type Node")
	}

}

func Test_SortOrder(t *testing.T) {
	svc := app(t)
	// /sort
	// /sort/bookmark1
	// /sort/bookmark2
	// /sort/bookmark3
	url := "http://www.example.com"

	sortFolder := uuid.NewString()
	svc.CreateBookmark(bookmarks.Bookmark{
		Type:        bookmarks.Folder,
		DisplayName: sortFolder,
		Path:        "/",
	}, user)
	bookmark1 := uuid.NewString()
	b1, _ := svc.CreateBookmark(bookmarks.Bookmark{
		Type:        bookmarks.Node,
		DisplayName: "1_" + bookmark1,
		Path:        "/" + sortFolder,
		URL:         url,
	}, user)
	bookmark2 := uuid.NewString()
	b2, _ := svc.CreateBookmark(bookmarks.Bookmark{
		Type:        bookmarks.Node,
		DisplayName: "2_" + bookmark2,
		Path:        "/" + sortFolder,
		URL:         url,
	}, user)
	bookmark3 := uuid.NewString()
	b3, _ := svc.CreateBookmark(bookmarks.Bookmark{
		Type:        bookmarks.Node,
		DisplayName: "3_" + bookmark3,
		Path:        "/" + sortFolder,
		URL:         url,
	}, user)

	// get the initial sort-order
	bms, err := svc.GetBookmarksByPath("/"+sortFolder, user)
	if err != nil {
		t.Errorf("could not get bookmarks unsorted! %v", err)
	}
	if len(bms) == 0 {
		t.Errorf("should have received 3 bookmarks got %d", len(bms))
	}

	// create the sort-order structure
	so := bookmarks.BookmarksSortOrder{
		IDs: []string{b1.ID, b2.ID, b3.ID},
	}

	// 1) reverse
	so.SortOrder = []int{2, 1, 0}
	count, err := svc.UpdateSortOrder(so, user)
	if err != nil {
		t.Errorf("could not update the sortorder of bookmarks; %v", err)
	}
	if count != 3 {
		t.Errorf("should have updated 3 bookmarks, but the number was %d", count)
	}
	bms, _ = svc.GetBookmarksByPath("/"+sortFolder, user)
	if len(bms) != 3 {
		t.Errorf("should have received 3 bookmarks got %d", len(bms))
	}
	if bms[0].ID != so.IDs[2] && bms[2].ID != so.IDs[0] {
		t.Errorf("the update of the sortorder dit not work, got mixed-up results")
	}

	// 2) other
	so.SortOrder = []int{0, 2, 1}
	_, err = svc.UpdateSortOrder(so, user)
	if err != nil {
		t.Errorf("could not update the sortorder of bookmarks; %v", err)
	}
	bms, _ = svc.GetBookmarksByPath("/"+sortFolder, user)
	if len(bms) != 3 {
		t.Errorf("should have received 3 bookmarks got %d", len(bms))
	}
	if bms[0].ID != so.IDs[0] && bms[2].ID != so.IDs[1] && bms[1].ID != so.IDs[2] {
		t.Errorf("the update of the sortorder dit not work, got mixed-up results")
	}

	// empty slice, no updates
	so.IDs = make([]string, 0)
	so.SortOrder = make([]int, 0)
	count, _ = svc.UpdateSortOrder(so, user)
	if count != 0 {
		t.Errorf("expected 0 updates but got %d", count)
	}

	// ---- error ----

	// unkown ID
	so.IDs = []string{b1.ID, b2.ID, "unknownd-ID"}
	so.SortOrder = []int{2, 1, 0}
	_, err = svc.UpdateSortOrder(so, user)
	if err == nil {
		t.Errorf("expected an error for unknown-ID")
	}

	// unbalanced slices
	so.IDs = []string{b1.ID, b2.ID, b3.ID}
	so.SortOrder = []int{2, 1}
	_, err = svc.UpdateSortOrder(so, user)
	if err == nil {
		t.Errorf("expected an error for unknown-ID")
	}
}

func Test_UpdateBookmark(t *testing.T) {
	svc := app(t)

	// /update/bookmark
	url := "http://www.example.com"

	updateFolder := uuid.NewString()
	f, _ := svc.CreateBookmark(bookmarks.Bookmark{
		Type:        bookmarks.Folder,
		DisplayName: updateFolder,
		Path:        "/",
	}, user)
	updateFolder2 := uuid.NewString()
	svc.CreateBookmark(bookmarks.Bookmark{
		Type:        bookmarks.Folder,
		DisplayName: updateFolder2,
		Path:        "/",
	}, user)
	bookmark := uuid.NewString()
	bm, _ := svc.CreateBookmark(bookmarks.Bookmark{
		Type:        bookmarks.Node,
		DisplayName: bookmark,
		Path:        "/" + updateFolder,
		URL:         url,
	}, user)

	// get the bookmark
	b, _ := svc.GetBookmarkByID(bm.ID, user)
	if b == nil {
		t.Fatalf("could not get bookmark by id %s", bm.ID)
	}

	// update the bookmark
	b.DisplayName = bookmark + "_update"
	b.InvertFaviconColor = 1
	_, err := svc.UpdateBookmark(*b, user)
	if err != nil {
		t.Errorf("could not update bookmark, %v", err)
	}
	b, _ = svc.GetBookmarkByID(bm.ID, user)

	if b.DisplayName != bookmark+"_update" {
		t.Errorf("the bookmark was not correctly updated, got %s", b.DisplayName)
	}
	if b.InvertFaviconColor != 1 {
		t.Errorf("the bookmark was not correctly updated, got %d", b.InvertFaviconColor)
	}

	// update the folder
	f, _ = svc.GetBookmarkByID(f.ID, user)
	f.DisplayName = f.DisplayName + "_update"
	_, err = svc.UpdateBookmark(*f, user)
	if err != nil {
		t.Errorf("could not update bookmark, %v", err)
	}

	// fetch the bookmark again
	b, _ = svc.GetBookmarkByID(bm.ID, user)
	if b.Path != "/"+updateFolder+"_update" {
		t.Errorf("the bookmark-path was not correctly updated; %s", b.Path)
	}

	// change the path of the bookmark
	b, _ = svc.GetBookmarkByID(bm.ID, user)
	b.Path = "/" + updateFolder2
	_, err = svc.UpdateBookmark(*b, user)
	if err != nil {
		t.Errorf("could not update bookmark, %v", err)
	}

	// ---- error: empty ----
	_, err = svc.UpdateBookmark(bookmarks.Bookmark{}, user)
	if err == nil {
		t.Errorf("could not update bookmark, %v", err)
	}
	// ---- error: unknown-id ----
	_, err = svc.UpdateBookmark(bookmarks.Bookmark{
		Path:        "/",
		DisplayName: "abc",
		ID:          "unknown-ID",
	}, user)
	if err == nil {
		t.Error("expected an error because of unknown-ID")
	}
	// ---- error: move folder to self ----
	f, _ = svc.GetBookmarkByID(f.ID, user)
	f.Path = f.Path + f.DisplayName
	_, err = svc.UpdateBookmark(*f, user)
	if err == nil {
		t.Errorf("could not update bookmark, %v", err)
	}
}

func Test_UpateBookmarks_WithFavicons(t *testing.T) {
	svc := app(t)
	server := hostTestWebserver(t)
	defer server.Close()

	obj, err := svc.LocalFetchFaviconURL(server.URL + "/Wikipedia-logo.png")
	if err != nil {
		t.Errorf("error fetching favicon: %v", err)
	}

	// /bookmark

	bookmark := uuid.NewString()
	bm, err := svc.CreateBookmark(bookmarks.Bookmark{
		Type:        bookmarks.Node,
		DisplayName: bookmark,
		URL:         "http://localhost",
		Path:        "/",
		Favicon:     obj.Name,
	}, user)
	if err != nil {
		t.Errorf("could not create bookmark; %v", err)
	}

	// fetch a new favicon
	obj, err = svc.LocalFetchFaviconURL(server.URL + "/Example.png")
	if err != nil {
		t.Errorf("error fetching favicon: %v", err)
	}
	bm.Favicon = obj.Name

	bm, err = svc.UpdateBookmark(*bm, user)
	if err != nil {
		t.Errorf("could not update bookmark; %v", err)
	}
	assert.Equal(t, obj.Name, bm.Favicon)
}

func Test_UpdateBookmarks_WithFile(t *testing.T) {
	svc := app(t)

	fileID, _ := svc.UploadSvc.Save(upload.File{
		Name:     "test.pdf",
		MimeType: "application/pdf",
		File:     bytes.NewReader([]byte(pdfPayload)),
		Size:     int64(len([]byte(pdfPayload))),
	})
	bm, _ := svc.CreateBookmark(bookmarks.Bookmark{
		Type:        bookmarks.FileItem,
		Path:        "/",
		DisplayName: "test.pdf",
		FileID:      fileID,
	}, user)

	// get the bookmark
	b, _ := svc.GetBookmarkByID(bm.ID, user)
	if b == nil {
		t.Fatalf("could not get bookmark by id %s", bm.ID)
	}

	// update the bookmark by changing the file
	fileID, _ = svc.UploadSvc.Save(upload.File{
		Name:     "test_updated.pdf",
		MimeType: "application/pdf",
		File:     bytes.NewReader([]byte(pdfPayload)),
		Size:     int64(len([]byte(pdfPayload))),
	})
	b.FileID = fileID

	bm, err := svc.UpdateBookmark(*b, user)
	if err != nil {
		t.Errorf("could not update bookmark, %v", err)
	}

	// retrieve the bookmark again
	fetched, err := svc.GetBookmarkByID(bm.ID, user)
	if err != nil {
		t.Errorf("could not fetch bookmark; %v", err)
	}

	if fetched.FileID == "" {
		t.Errorf("the fetched bookmark is missing a fileID")
	}
	if fetched.FilePayload.Payload == nil {
		t.Errorf("the fetched bookmark is missing a file payload")
	}
	if fetched.FilePayload.Name != "test_updated.pdf" {
		t.Errorf("the file payload of the bookmark was not saved")
	}

}

func Test_GetAllAvailableFavicons(t *testing.T) {
	svc := app(t)
	server := hostTestWebserver(t)
	defer server.Close()

	var favUser = security.User{
		Username:    uuid.NewString(),
		Email:       "a.b@c.de",
		DisplayName: "Test_" + uuid.NewString(),
	}

	// 1)
	obj, err := svc.LocalFetchFaviconURL(server.URL + "/Wikipedia-logo.png")
	if err != nil {
		t.Errorf("error fetching favicon: %v", err)
	}

	_, err = svc.CreateBookmark(bookmarks.Bookmark{
		Type:        bookmarks.Node,
		DisplayName: uuid.NewString(),
		URL:         "http://localhost1",
		Path:        "/",
		Favicon:     obj.Name,
	}, favUser)
	if err != nil {
		t.Errorf("could not store bookmark; %v", err)
	}

	// 2)
	obj, err = svc.LocalFetchFaviconURL(server.URL + "/Example.png")
	if err != nil {
		t.Errorf("error fetching favicon: %v", err)
	}

	_, err = svc.CreateBookmark(bookmarks.Bookmark{
		Type:        bookmarks.Node,
		DisplayName: uuid.NewString(),
		URL:         "http://localhost2",
		Path:        "/",
		Favicon:     obj.Name,
	}, favUser)
	if err != nil {
		t.Errorf("could not store bookmark; %v", err)
	}

	// 3)
	obj, err = svc.LocalFetchFaviconURL(server.URL + "/Wikipedia-logo.png")
	if err != nil {
		t.Errorf("error fetching favicon: %v", err)
	}

	_, err = svc.CreateBookmark(bookmarks.Bookmark{
		Type:        bookmarks.Node,
		DisplayName: uuid.NewString(),
		URL:         "http://localhost3",
		Path:        "/",
		Favicon:     obj.Name,
	}, favUser)
	if err != nil {
		t.Errorf("could not store bookmark; %v", err)
	}

	// 4)
	obj, err = svc.LocalFetchFaviconURL(server.URL + "/Example.png")
	if err != nil {
		t.Errorf("error fetching favicon: %v", err)
	}

	_, err = svc.CreateBookmark(bookmarks.Bookmark{
		Type:        bookmarks.Node,
		DisplayName: uuid.NewString(),
		URL:         "http://localhost4",
		Path:        "/",
		Favicon:     obj.Name,
	}, favUser)
	if err != nil {
		t.Errorf("could not store bookmark; %v", err)
	}

	// 5)
	_, err = svc.CreateBookmark(bookmarks.Bookmark{
		Type:        bookmarks.Folder,
		DisplayName: uuid.NewString(),
		Path:        "/",
	}, favUser)
	if err != nil {
		t.Errorf("could not store bookmark; %v", err)
	}

	// we created 4 bookmarks with 2 different favicons. The result of all available favicons should be just 2
	// 4were saved but 2 used the same favicon, therefor the unique result should be only 2.

	favicons, err := svc.GetAvailableFavicons(favUser, "")
	if err != nil {
		t.Errorf("could not get available favicons; %v", err)
	}

	if len(favicons) != 2 {
		t.Errorf("expected 2 results but got %d", len(favicons))
	}

	// single favicon
	// -------------------------------------------------------------------------------------------------
	favicon, err := svc.GetFaviconByID(favicons[0].Name)
	if err != nil {
		t.Errorf("could not get a favicon by it's ID; %v", err)
	}
	if len(favicon.Payload) <= 0 {
		t.Errorf("the fetched favicon has no payload")
	}

	_, err = svc.GetFaviconByID("random_ID")
	if err == nil {
		t.Errorf("expected error because of wrong ID")
	}

	_, err = svc.GetFaviconByID("")
	if err == nil {
		t.Errorf("expected error because of missing ID")
	}

	// no favicons
	// -------------------------------------------------------------------------------------------------
	var unknownUser = security.User{
		Username:    uuid.NewString(),
		Email:       "a.b@c.de",
		DisplayName: "Unknown_" + uuid.NewString(),
	}
	favicons, err = svc.GetAvailableFavicons(unknownUser, "")
	if err != nil {
		t.Errorf("could not get available favicons; %v", err)
	}
	if len(favicons) != 0 {
		t.Errorf("expected 0 results but got %d", len(favicons))
	}
}

//go:embed Wikipedia-logo.png
var wikipediaLogo []byte

//go:embed Example.png
var examplePng []byte

//go:embed orf.at.html
var orfAT []byte

//go:embed favicon-32x32.png
var orfFavicon []byte

// hostTestWebserver is used to prevent the use of external resources when downloading images and other content
func hostTestWebserver(t *testing.T) *httptest.Server {
	// setup a test-server
	// ------------------------------------------------------------------
	mux := http.NewServeMux()
	writePayload := func(w http.ResponseWriter, payload []byte) {
		w.Header().Add("content-length", fmt.Sprintf("%d", len(payload)))
		if _, err := w.Write(payload); err != nil {
			t.Fatalf("%v", err)
		}
	}
	writeHTMLPayload := func(w http.ResponseWriter, payload []byte) {
		w.Header().Add("content-length", fmt.Sprintf("%d", len(payload)))
		w.Header().Add("content-type", "text/html")
		if _, err := w.Write(payload); err != nil {
			t.Fatalf("%v", err)
		}
	}

	mux.HandleFunc("/Wikipedia-logo.png", func(w http.ResponseWriter, r *http.Request) { writePayload(w, wikipediaLogo) })
	mux.HandleFunc("/Example.png", func(w http.ResponseWriter, r *http.Request) { writePayload(w, examplePng) })
	mux.HandleFunc("/orf.at/favicon-32x32.png", func(w http.ResponseWriter, r *http.Request) { writePayload(w, orfFavicon) })
	mux.HandleFunc("/orf.at", func(w http.ResponseWriter, r *http.Request) { writeHTMLPayload(w, orfAT) })

	return httptest.NewServer(mux)
}

// found an error when a folder named 'A' is moved to a location which already has a folder 'A'
// no overwrite is performed but the result is that two folders 'A' are available side-by-side.
func Test_MoveSameNamedFolder(t *testing.T) {
	svc := app(t)
	server := hostTestWebserver(t)
	defer server.Close()

	var user = security.User{
		Username:    uuid.NewString(),
		Email:       "a.b@c.de",
		DisplayName: "Test_" + uuid.NewString(),
	}

	// create a TARGET-ROOT folder where we have an
	// existing folder named Existing
	//   /TARGET-ROOT
	//    - /SubFolder
	//	- Bookmark1
	targetRoot := "TARGET-ROOT"
	_, err := svc.CreateBookmark(bookmarks.Bookmark{
		Type:        bookmarks.Folder,
		DisplayName: targetRoot,
		Path:        "/",
	}, user)
	if err != nil {
		t.Errorf("could not create targetRoot; %v", err)
	}
	subFolderName := "Subfolder"
	_, err = svc.CreateBookmark(bookmarks.Bookmark{
		Type:        bookmarks.Folder,
		DisplayName: subFolderName,
		Path:        "/" + targetRoot,
	}, user)
	if err != nil {
		t.Errorf("could not create subFolderName in targetRoot; %v", err)
	}
	_, err = svc.CreateBookmark(bookmarks.Bookmark{
		Type:        bookmarks.Node,
		DisplayName: "Bookmark1",
		URL:         "http://example.com",
		Path:        "/" + targetRoot + "/" + subFolderName,
	}, user)
	if err != nil {
		t.Errorf("could not create Bookmark in targetRoot/subFolderName; %v", err)
	}

	// create the same subfolder in ROOT
	//   /SubFolder		<- move this folder to TARGET-ROOT
	//	- Bookmark2
	rootSubFolder, err := svc.CreateBookmark(bookmarks.Bookmark{
		Type:        bookmarks.Folder,
		DisplayName: subFolderName,
		Path:        "/", // <-- create the same folder in ROOT
	}, user)
	if err != nil {
		t.Errorf("could not create subFolderName in targetRoot; %v", err)
	}
	_, err = svc.CreateBookmark(bookmarks.Bookmark{
		Type:        bookmarks.Node,
		DisplayName: "Bookmark2",
		URL:         "http://example.com",
		Path:        "/" + subFolderName,
	}, user)
	if err != nil {
		t.Errorf("could not create Bookmark in targetRoot/subFolderName; %v", err)
	}

	// setting the stage done (longe one!)

	// if the Folder /SubFolder is moved to /TARGET-ROOT/SubFolder the two folders
	// with the same name need to be merged and the one remaining folder should have
	// all the bookmarks of the two folders combined
	rootSubFolder.Path = "/" + targetRoot
	_, err = svc.UpdateBookmark(*rootSubFolder, user)
	if err != nil {
		t.Errorf("could not move rootSubFolder to targetRoot; %v", err)
	}

	folders, err := svc.GetBookmarksByPath("/"+targetRoot, user)
	if err != nil {
		t.Errorf("could not get bookmark folders for path /TARGET-ROOT; %v", err)
	}
	if len(folders) != 1 {
		var names string
		for _, f := range folders {
			if names != "" {
				names += ", "
			}
			names += fmt.Sprintf("name=%s (type=%s)", f.DisplayName, f.Type)
		}
		t.Errorf("expected on merged folder but got %d folders; %s", len(folders), names)
	}

	// the remaining folder should have two bookmarks
	//   /TARGET-ROOT
	//    - /SubFolder
	//	- Bookmark1
	//	- Bookmark2
	bookmarks, err := svc.GetBookmarksByPath("/"+targetRoot+"/"+subFolderName, user)
	if err != nil {
		t.Errorf("could not get bookmark folders for path /TARGET-ROOT/Subfolder; %v", err)
	}
	if len(bookmarks) != 2 {
		t.Errorf("expected to have 2 bookmarks but only got %d", len(bookmarks))
	}
	var bmNames string
	for _, bm := range bookmarks {
		if bmNames != "" {
			bmNames += ","
		}
		bmNames += bm.DisplayName
	}
	if !(strings.Contains(bmNames, "Bookmark1") && strings.Contains(bmNames, "Bookmark2")) {
		t.Errorf("expected to have 2 bookmarks Bookmark1/Bookmark2 but got %s", bmNames)
	}

	// the child-count of the SubFolder should be 2
	subFolder, err := svc.GetBookmarkByID(rootSubFolder.ID, user)
	if err != nil {
		t.Fatalf("could not get bookmark folder 'SubFolder' by ID; %v", err)
	}
	if subFolder.ChildCount != 2 {
		t.Errorf("expected a childcount of %d, but got %d", 2, subFolder.ChildCount)
	}
}
