package bookmarks_test

import (
	"testing"

	"github.com/google/uuid"
	"golang.binggl.net/monorepo/internal/bookmarks-new/app/bookmarks"
	"golang.binggl.net/monorepo/internal/bookmarks-new/app/store"
	"golang.binggl.net/monorepo/pkg/logging"
	"golang.binggl.net/monorepo/pkg/security"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const faviconPath = "/tmp"

var user = security.User{
	Username:    "test",
	Email:       "a.b@c.de",
	DisplayName: "Test",
}

func service(t *testing.T) bookmarks.Service {
	return bookmarks.Service{
		Store:       repository(t),
		Logger:      logging.NewNop(),
		FaviconPath: faviconPath,
	}
}

func repository(t *testing.T) store.Repository {
	var (
		DB  *gorm.DB
		err error
	)
	if DB, err = gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{}); err != nil {
		t.Fatalf("cannot create database connection: %v", err)
	}
	// Migrate the schema
	DB.AutoMigrate(&store.Bookmark{})
	logger := logging.NewNop()
	return store.Create(DB, logger)
}

func Test_GetBookmark_NotFound(t *testing.T) {
	svc := service(t)

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
	svc := service(t)

	folder, err := svc.CreateBookmark(bookmarks.Bookmark{
		Type:        bookmarks.Folder,
		Path:        "/",
		DisplayName: "a",
	}, user)
	if err != nil {
		t.Errorf("could not create bookmark; %v", err)
	}

	bm, err := svc.CreateBookmark(bookmarks.Bookmark{
		Type:        bookmarks.Node,
		Path:        "/" + folder.DisplayName,
		DisplayName: "test",
		URL:         "https://www.orf.at",
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

	// ---- error case ----
	_, err = svc.CreateBookmark(bookmarks.Bookmark{
		Type:        bookmarks.Folder,
		Path:        "",
		DisplayName: "a",
	}, user)
	if err == nil {
		t.Errorf("error expected")
	}

	_, err = svc.CreateBookmark(bookmarks.Bookmark{
		Type:        bookmarks.Folder,
		Path:        "/",
		DisplayName: "",
	}, user)
	if err == nil {
		t.Errorf("error expected")
	}
}

func Test_GetBookmarksByPath(t *testing.T) {
	svc := service(t)
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
	svc := service(t)
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
	svc := service(t)
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
	paths, _ = svc.GetAllPaths(security.User{})
	if len(paths) != 0 {
		t.Errorf("expected to get no results for unknown user - god %d!", len(paths))
	}
}

func Test_GetBookmarksByName(t *testing.T) {
	svc := service(t)
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
	svc := service(t)

	// first we need a base-bookmark ;)
	bm, err := svc.CreateBookmark(bookmarks.Bookmark{
		Type:        bookmarks.Node,
		Path:        "/",
		DisplayName: "test",
		URL:         "https://www.orf.at",
	}, user)
	if err != nil {
		t.Errorf("could not create bookmark; %v", err)
	}

	nodeType := store.Node
	if bm.Type == bookmarks.Folder {
		nodeType = store.Folder
	}

	bmEntity := store.Bookmark{
		ID:          bm.ID,
		Path:        bm.Path,
		DisplayName: bm.DisplayName,
		URL:         "https://www.orf.at",
		SortOrder:   bm.SortOrder,
		Type:        nodeType,
		UserName:    user.Username,
		Created:     bm.Created,
		Modified:    bm.Modified,
		ChildCount:  bm.ChildCount,
		Highlight:   bm.Highlight,
		Favicon:     bm.Favicon,
	}

	svc.FetchFavicon(bmEntity, user)
	svc.FetchFaviconURL("https://orf.at/mojo/1_4_1/storyserver//news/news/images/touch-icon-iphone.png", bmEntity, user)
}
