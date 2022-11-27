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
