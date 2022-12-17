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
		Store:          repository(t),
		Logger:         logging.NewNop(),
		FaviconPath:    faviconPath,
		DefaultFavicon: "default.icon",
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
	paths, _ = svc.GetAllPaths(security.User{Username: "unknown"})
	if len(paths) != 1 { // we always get "/"
		t.Errorf("expected to get no results for unknown user - got %d!", len(paths))
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

func Test_FetchBookmark(t *testing.T) {
	svc := service(t)
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
	svc := service(t)
	// /folder/bookmark

	folder := uuid.NewString()
	svc.CreateBookmark(bookmarks.Bookmark{
		Type:        bookmarks.Folder,
		DisplayName: folder,
		Path:        "/",
	}, user)

	url := "http://localhost"
	bookmark := uuid.NewString()
	bm, _ := svc.CreateBookmark(bookmarks.Bookmark{
		Type:        bookmarks.Node,
		DisplayName: bookmark,
		URL:         url,
		Path:        "/" + folder,
		Favicon:     "unknown.ico",
	}, user)

	path, err := svc.GetFaviconPath(bm.ID, user)
	if err != nil {
		t.Errorf("could not get path of favicon for ID: %s; %v", bm.ID, err)
	}

	if path == "" {
		t.Errorf("expected a path for the faviconGetFaviconPath")
	}

	// ---- error ----

	_, err = svc.GetFaviconPath("", user)
	if err == nil {
		t.Errorf("expected error for empty id")
	}

	_, err = svc.GetFaviconPath("unknown-id", user)
	if err == nil {
		t.Errorf("expected error for unknown id")
	}
}

func Test_DeleteBookmark(t *testing.T) {
	svc := service(t)
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
		Favicon:     "unknown.ico",
	}, user)

	// try to delete the folder which should not work because of the child-node
	err := svc.Delete(bmF.ID, user)
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

func Test_SortOrder(t *testing.T) {
	svc := service(t)
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
