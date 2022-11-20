package store_test

import (
	"testing"

	"golang.binggl.net/monorepo/internal/bookmarks-new/app/store"
)

func Test_Bookmark_CRUD(t *testing.T) {
	repo := store.NewMockRepository()

	// ---- create ----

	bm := store.Bookmark{
		ID:          "1",
		Type:        store.Node,
		URL:         "http://www.test.com",
		DisplayName: "text",
		UserName:    "user",
	}
	b, err := repo.Create(bm)
	if err != nil {
		t.Fatalf("no error expected; %v", err)
	}
	if b.URL != bm.URL {
		t.Errorf("expected same URLs for created bookmark")
	}

	// ---- read id ----

	found, err := repo.GetBookmarkByID(bm.ID, bm.UserName)
	if err != nil {
		t.Fatalf("no error expected; %v", err)
	}
	if found.URL != bm.URL {
		t.Errorf("expected same URLs for created bookmark")
	}

	_, err = repo.GetBookmarkByID(bm.ID, "")
	if err == nil {
		t.Fatalf("error expected")
	}

	_, err = repo.GetBookmarkByID("0", bm.UserName)
	if err == nil {
		t.Fatalf("error expected")
	}

	// ---- read all ----

	bookmarks, err := repo.GetAllBookmarks(bm.UserName)
	if err != nil {
		t.Fatalf("no error expected; %v", err)
	}
	if len(bookmarks) != 1 {
		t.Fatal("expected some bookmarks")
	}

	_, err = repo.GetAllBookmarks("")
	if err == nil {
		t.Fatal("error expected")
	}

	// ---- update ----

	b.DisplayName = bm.DisplayName + "_update"
	b, err = repo.Update(b)
	if err != nil {
		t.Fatalf("no error expected; %v", err)
	}
	if b.DisplayName != bm.DisplayName+"_update" {
		t.Errorf("expected updated displayName")
	}

	// a different user is not found
	_, err = repo.Update(store.Bookmark{
		UserName: "differentUser",
	})
	if err == nil {
		t.Fatalf("error expected")
	}

	// a different ID is not found
	b.ID = "99"
	_, err = repo.Update(b)
	if err == nil {
		t.Fatalf("error expected")
	}

	// ---- delete ----

	err = repo.Delete(store.Bookmark{
		UserName: "differentUser",
	})
	if err == nil {
		t.Fatalf("error expected")
	}

	// a different ID is not found
	b.ID = "99"
	err = repo.Delete(b)
	if err == nil {
		t.Fatalf("error expected")
	}

	// remove the bookmark item
	err = repo.Delete(bm)
	if err != nil {
		t.Fatalf("could not remove bookmark; %v", err)
	}
}
