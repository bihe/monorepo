package store_test

import (
	"database/sql"
	"fmt"
	"math/rand"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.binggl.net/monorepo/internal/bookmarks/app/store"
	"golang.binggl.net/monorepo/pkg/persistence"
)

func repository(t *testing.T) (store.BookmarkRepository, *sql.DB) {
	return repositoryFile(":memory:", t)
}

func repositoryFile(file string, t *testing.T) (store.BookmarkRepository, *sql.DB) {
	var (
		err error
	)
	dbCon := persistence.MustCreateSqliteConn(file)
	con, err := persistence.CreateGormSqliteCon(dbCon)
	if err != nil {
		t.Fatalf("cannot create database connection: %v", err)
	}
	// Migrate the schema
	con.Write.AutoMigrate(&store.Bookmark{})
	con.Read.AutoMigrate(&store.Bookmark{})
	db, err := con.Write.DB()
	if err != nil {
		t.Fatalf("could not get DB handle; %v", err)
	}
	con.Read = con.Write
	repo := store.CreateBookmarkRepo(con, logger)
	return repo, db
}

func TestGetAllBookmarks(t *testing.T) {
	repo, db := repository(t)
	defer db.Close()

	userName := "test"
	bookmarks, err := repo.GetAllBookmarks(userName)
	if err != nil {
		t.Errorf("Could not get bookmarks: %v", err)
	}
	if len(bookmarks) != 0 {
		t.Errorf("Invalid number of bookmarks returned: %d", len(bookmarks))
	}
}

func TestCreateBookmark(t *testing.T) {
	repo, db := repository(t)
	defer db.Close()

	userName := "username"
	item := store.Bookmark{
		DisplayName:        "displayName",
		Path:               "/",
		SortOrder:          0,
		Type:               store.Node,
		URL:                "http://url",
		UserName:           userName,
		InvertFaviconColor: 1,
	}
	bm, err := repo.Create(item)
	if err != nil {
		t.Errorf("Could not create bookmarks: %v", err)
	}

	assert.NotEmpty(t, bm.ID)
	assert.Equal(t, "displayName", bm.DisplayName)
	assert.Equal(t, "http://url", bm.URL)
	assert.Equal(t, 1, bm.InvertFaviconColor)

	// check item.Path empty error
	item.Path = ""
	bm, err = repo.Create(item)
	if err == nil {
		t.Errorf("Expected error on empty path!")
	}
}

func TestCreatBookmarkAndHierarchy(t *testing.T) {
	repo, db := repository(t)
	defer db.Close()

	userName := "userName"

	// create a root folder and a "child-node"
	//
	// /Folder
	// /Folder/Node
	folder := store.Bookmark{
		DisplayName: "Folder",
		Path:        "/",
		SortOrder:   0,
		Type:        store.Folder,
		UserName:    userName,
	}
	bm, err := repo.Create(folder)
	if err != nil {
		t.Errorf("Could not create bookmarks: %v", err)
	}
	assert.NotEmpty(t, bm.ID)

	_, err = repo.GetBookmarkByID(bm.ID, userName)
	if err != nil {
		t.Errorf("Could not read bookmarks: %v", err)
	}

	node := store.Bookmark{
		DisplayName: "Node",
		Path:        "/Folder",
		SortOrder:   0,
		Type:        store.Node,
		URL:         "http://url",
		UserName:    userName,
	}
	bm, err = repo.Create(node)
	if err != nil {
		t.Errorf("Could not create bookmarks: %v", err)
	}
	assert.NotEmpty(t, bm.ID)

	n, err := repo.GetBookmarkByID(bm.ID, userName)
	if err != nil {
		t.Errorf("Could not read bookmarks: %v", err)
	}

	assert.NotEmpty(t, n.ID)
	assert.Equal(t, "Node", node.DisplayName)
	assert.Equal(t, "/Folder", node.Path)
	assert.Equal(t, store.Node, node.Type)
	assert.Equal(t, "http://url", node.URL)

	// error creating node because of missing folder
	node.Path = "/unknown/folder"
	bm, err = repo.Create(node)
	if err == nil {
		t.Errorf("Expected error for unknown path!")
	}
}

func TestCreateBookmarkInUnitOfWork(t *testing.T) {
	repo, db := repository(t)
	defer db.Close()

	userName := "userName"

	folder := store.Bookmark{
		DisplayName: "Folder",
		Path:        "/",
		SortOrder:   0,
		Type:        store.Folder,
		UserName:    userName,
	}

	id := ""
	err := repo.InUnitOfWork(func(r store.BookmarkRepository) error {
		bm, err := r.Create(folder)
		if err != nil {
			return err
		}
		assert.NotEmpty(t, bm.ID)
		id = bm.ID
		folder, err = r.GetBookmarkByID(bm.ID, bm.UserName)
		if err != nil {
			return err
		}
		assert.Equal(t, "Folder", folder.DisplayName)
		return nil
	})
	if err != nil {
		t.Errorf("Cannot execute in UnitOfWork: %v", err)
	}

	// get the created entry "out-of" the transaction
	folder, err = repo.GetBookmarkByID(id, userName)
	if err != nil {
		t.Errorf("could not get the created folder by id: %v", err)
	}
	assert.Equal(t, "Folder", folder.DisplayName)

	// test a rollback
	folderRollback := store.Bookmark{
		DisplayName: "Folder",
		Path:        "/",
		SortOrder:   0,
		Type:        store.Folder,
		UserName:    userName,
	}
	var bm store.Bookmark
	err = repo.InUnitOfWork(func(r store.BookmarkRepository) error {
		bm, err = r.Create(folderRollback)
		if err != nil {
			return err
		}
		assert.NotEmpty(t, bm.ID)
		id = bm.ID
		folder, err = r.GetBookmarkByID(bm.ID, bm.UserName)
		if err != nil {
			return err
		}
		assert.Equal(t, "Folder", folder.DisplayName)
		return fmt.Errorf("rollback")
	})
	if err == nil {
		t.Errorf("expected an error because of rollback")
	}

	// get the created entry "out-of" the transaction
	folder, err = repo.GetBookmarkByID(id, userName)
	if err == nil {
		t.Errorf("expected an error because item not available!")
	}
}

func TestUpdateBookmark(t *testing.T) {
	repo, db := repository(t)
	defer db.Close()

	userName := "username"
	item := store.Bookmark{
		DisplayName: "displayName",
		Path:        "/",
		SortOrder:   0,
		Type:        store.Node,
		URL:         "http://url",
		UserName:    userName,
	}
	bm, err := repo.Create(item)
	if err != nil {
		t.Errorf("Could not create bookmarks: %v", err)
	}

	assert.NotEmpty(t, bm.ID)
	assert.Equal(t, "displayName", bm.DisplayName)
	assert.Equal(t, "http://url", bm.URL)

	bm.DisplayName = bm.DisplayName + "_update"

	bm, err = repo.Update(bm)
	if err != nil {
		t.Errorf("Could not update bookmarks: %v", err)
	}
	assert.Equal(t, "displayName_update", bm.DisplayName)

	// update with empty-path
	bm.Path = ""
	bm, err = repo.Update(bm)
	if err == nil {
		t.Errorf("Expected error because of empty path")
	}

	// update with unknown-path
	bm.Path = "/unknown/path"
	bm, err = repo.Update(bm)
	if err == nil {
		t.Errorf("Expected error because of unknown path")
	}
}

func TestDeleteBookmark(t *testing.T) {
	repo, db := repository(t)
	defer db.Close()

	userName := "userName"

	// create a root folder and a "child-node"
	//
	// /Folder
	// /Folder/Node
	f := store.Bookmark{
		DisplayName: "Folder",
		Path:        "/",
		SortOrder:   0,
		Type:        store.Folder,
		UserName:    userName,
	}
	folder, err := repo.Create(f)
	if err != nil {
		t.Errorf("Could not create bookmarks: %v", err)
	}
	assert.NotEmpty(t, folder.ID)

	n := store.Bookmark{
		DisplayName: "Node",
		Path:        "/Folder",
		SortOrder:   0,
		Type:        store.Node,
		URL:         "http://url",
		UserName:    userName,
	}
	node, err := repo.Create(n)
	if err != nil {
		t.Errorf("Could not create bookmarks: %v", err)
	}
	assert.NotEmpty(t, node.ID)

	folder, err = repo.GetBookmarkByID(folder.ID, folder.UserName)
	if err != nil {
		t.Errorf("Could not read bookmark folder: %v", err)
	}
	assert.Equal(t, 1, folder.ChildCount)

	// remove the node
	err = repo.Delete(node)
	if err != nil {
		t.Errorf("Could not delete bookmark: %v", err)
	}

	// child-count has to be 0 again
	folder, err = repo.GetBookmarkByID(folder.ID, folder.UserName)
	if err != nil {
		t.Errorf("Could not read bookmark folder: %v", err)
	}
	assert.Equal(t, 0, folder.ChildCount)
}

func TestGetPathChildCount(t *testing.T) {
	repo, db := repository(t)
	defer db.Close()
	errStr := "Could not create bookmarks: %v"
	errStrNodes := "Could not get node-count: %v"

	userName := "userName"

	// create a folder structure
	_, err := repo.Create(store.Bookmark{
		Type:        store.Folder,
		DisplayName: "Folder",
		UserName:    userName,
		Path:        "/",
	})
	if err != nil {
		t.Errorf(errStr, err)
	}
	_, err = repo.Create(store.Bookmark{
		Type:        store.Folder,
		DisplayName: "Folder1",
		UserName:    userName,
		Path:        "/",
	})
	if err != nil {
		t.Errorf(errStr, err)
	}
	_, err = repo.Create(store.Bookmark{
		Type:        store.Node,
		DisplayName: "Node",
		UserName:    userName,
		Path:        "/Folder",
		URL:         "http://url",
	})
	if err != nil {
		t.Errorf(errStr, err)
	}
	_, err = repo.Create(store.Bookmark{
		Type:        store.Node,
		DisplayName: "Node1",
		UserName:    userName,
		Path:        "/Folder1",
		URL:         "http://url",
	})
	if err != nil {
		t.Errorf(errStr, err)
	}

	// Root
	nodes, err := repo.GetPathChildCount("/", userName)
	if err != nil {
		t.Errorf(errStrNodes, err)
	}
	assert.Equal(t, 1, len(nodes))
	assert.Equal(t, "/", nodes[0].Path)
	assert.Equal(t, 2, nodes[0].Count)

	// Folder
	nodes, err = repo.GetPathChildCount("/Folder", userName)
	if err != nil {
		t.Errorf(errStrNodes, err)
	}
	assert.Equal(t, 1, len(nodes))
	assert.Equal(t, "/Folder", nodes[0].Path)
	assert.Equal(t, 1, nodes[0].Count)

	// Folder1
	nodes, err = repo.GetPathChildCount("/Folder1", userName)
	if err != nil {
		t.Errorf(errStrNodes, err)
	}
	assert.Equal(t, 1, len(nodes))
	assert.Equal(t, "/Folder1", nodes[0].Path)
	assert.Equal(t, 1, nodes[0].Count)

	// no path
	_, err = repo.GetPathChildCount("", userName)
	if err == nil {
		t.Errorf("expected error because of missing path")
	}
}

func TestDeletePath(t *testing.T) {
	repo, db := repository(t)
	defer db.Close()
	errStr := "Could not create bookmarks: %v"

	userName := "userName"

	// create a folder structure
	_, err := repo.Create(store.Bookmark{
		Type:        store.Folder,
		DisplayName: "Folder",
		UserName:    userName,
		Path:        "/",
	})
	if err != nil {
		t.Errorf(errStr, err)
	}
	_, err = repo.Create(store.Bookmark{
		Type:        store.Folder,
		DisplayName: "Folder1",
		UserName:    userName,
		Path:        "/",
	})
	if err != nil {
		t.Errorf(errStr, err)
	}
	_, err = repo.Create(store.Bookmark{
		Type:        store.Node,
		DisplayName: "Node",
		UserName:    userName,
		Path:        "/Folder",
		URL:         "http://url",
	})
	if err != nil {
		t.Errorf(errStr, err)
	}
	_, err = repo.Create(store.Bookmark{
		Type:        store.Node,
		DisplayName: "Node1",
		UserName:    userName,
		Path:        "/Folder1",
		URL:         "http://url",
	})
	if err != nil {
		t.Errorf(errStr, err)
	}

	all, err := repo.GetAllBookmarks(userName)
	if err != nil {
		t.Errorf("could not get all bookmarks: %v", err)
	}
	assert.Equal(t, 4, len(all))

	// Delete the whole folder-path /Folder1 (1 x Node + 1 x Folder1)
	err = repo.DeletePath("/Folder1", userName)
	if err != nil {
		t.Errorf("could not delete folder path /Folder1: %v", err)
	}

	all, err = repo.GetAllBookmarks(userName)
	if err != nil {
		t.Errorf("could not get all bookmarks: %v", err)
	}
	assert.Equal(t, 2, len(all))

	// Delete the whole folder-path /Folder (1 x Node + 1 x Folder)
	err = repo.DeletePath("/Folder", userName)
	if err != nil {
		t.Errorf("could not delete folder path /Folder: %v", err)
	}

	all, err = repo.GetAllBookmarks(userName)
	if err != nil {
		t.Errorf("could not get all bookmarks: %v", err)
	}
	assert.Equal(t, 0, len(all))
}

func TestDeletePath2(t *testing.T) {
	repo, db := repository(t)
	defer db.Close()
	errStr := "Could not create bookmarks: %v"

	userName := "userName"

	// create a folder structure
	folder, err := repo.Create(store.Bookmark{
		Type:        store.Folder,
		DisplayName: "Folder",
		UserName:    userName,
		Path:        "/",
	})
	if err != nil {
		t.Errorf(errStr, err)
	}
	_, err = repo.Create(store.Bookmark{
		Type:        store.Folder,
		DisplayName: "Folder1",
		UserName:    userName,
		Path:        "/Folder",
	})
	if err != nil {
		t.Errorf(errStr, err)
	}
	_, err = repo.Create(store.Bookmark{
		Type:        store.Node,
		DisplayName: "Node",
		UserName:    userName,
		Path:        "/Folder",
		URL:         "http://url",
	})
	if err != nil {
		t.Errorf(errStr, err)
	}
	_, err = repo.Create(store.Bookmark{
		Type:        store.Node,
		DisplayName: "Node1",
		UserName:    userName,
		Path:        "/Folder/Folder1",
		URL:         "http://url",
	})
	if err != nil {
		t.Errorf(errStr, err)
	}

	all, err := repo.GetAllBookmarks(userName)
	if err != nil {
		t.Errorf("could not get all bookmarks: %v", err)
	}
	assert.Equal(t, 4, len(all))

	// get the child-count of folder
	bm, err := repo.GetBookmarkByID(folder.ID, folder.UserName)
	if err != nil {
		t.Errorf("could not get bookmark by id %s: %v", folder.ID, err)
	}
	assert.Equal(t, 2, bm.ChildCount)

	// Delete the whole folder-path /Folder/Folder1 (1 x Node + 1 x Folder1)
	err = repo.DeletePath("/Folder/Folder1", userName)
	if err != nil {
		t.Errorf("could not delete folder path /Folder1: %v", err)
	}

	bm, err = repo.GetBookmarkByID(folder.ID, folder.UserName)
	if err != nil {
		t.Errorf("could not get bookmark by id %s: %v", folder.ID, err)
	}
	assert.Equal(t, 1, bm.ChildCount)

	// error checks
	err = repo.DeletePath("", userName)
	if err == nil {
		t.Errorf("expected error empty path")
	}
	err = repo.DeletePath("/", userName)
	if err == nil {
		t.Errorf("expected error ROOT path")
	}
}

func TestGetBookmarksByPath(t *testing.T) {
	repo, db := repository(t)
	defer db.Close()
	errStr := "could not get bookmarks: %v"

	userName := "userName"

	// create a folder structure
	_, err := repo.Create(store.Bookmark{
		Type:        store.Folder,
		DisplayName: "Folder",
		UserName:    userName,
		Path:        "/",
	})
	if err != nil {
		t.Errorf(errStr, err)
	}
	_, err = repo.Create(store.Bookmark{
		Type:        store.Folder,
		DisplayName: "Folder1",
		UserName:    userName,
		Path:        "/Folder",
	})
	if err != nil {
		t.Errorf(errStr, err)
	}
	_, err = repo.Create(store.Bookmark{
		Type:        store.Folder,
		DisplayName: "Folder2",
		UserName:    userName,
		Path:        "/Folder/Folder1",
	})
	if err != nil {
		t.Errorf(errStr, err)
	}

	bookmarks, err := repo.GetBookmarksByPath("/", userName)
	if err != nil {
		t.Errorf(errStr, err)
	}

	assert.Equal(t, 1, len(bookmarks))
	assert.Equal(t, "Folder", bookmarks[0].DisplayName)

	bookmarks, err = repo.GetBookmarksByPath("/Folder", userName)
	if err != nil {
		t.Errorf(errStr, err)
	}
	assert.Equal(t, 1, len(bookmarks))
	assert.Equal(t, "Folder1", bookmarks[0].DisplayName)

	bookmarks, err = repo.GetBookmarksByPathStart("/Folder", userName)
	if err != nil {
		t.Errorf(errStr, err)
	}
	assert.Equal(t, 2, len(bookmarks))
	assert.Equal(t, "Folder1", bookmarks[0].DisplayName)
	assert.Equal(t, "Folder2", bookmarks[1].DisplayName)
}

func TestGetBookmarksByName(t *testing.T) {
	repo, db := repository(t)
	defer db.Close()
	errStr := "Could not create bookmarks: %v"
	errStr1 := "Could not get bookmarks: %v"

	userName := "userName"

	// create a folder structure
	_, err := repo.Create(store.Bookmark{
		Type:        store.Folder,
		DisplayName: "Folder",
		UserName:    userName,
		Path:        "/",
	})
	if err != nil {
		t.Errorf(errStr, err)
	}
	_, err = repo.Create(store.Bookmark{
		Type:        store.Node,
		DisplayName: "Node",
		UserName:    userName,
		Path:        "/Folder",
	})
	if err != nil {
		t.Errorf(errStr, err)
	}
	_, err = repo.Create(store.Bookmark{
		Type:        store.Node,
		DisplayName: "Node1",
		UserName:    userName,
		Path:        "/Folder",
	})
	if err != nil {
		t.Errorf(errStr, err)
	}

	bookmarks, err := repo.GetBookmarksByName("node", userName)
	if err != nil {
		t.Errorf(errStr1, err)
	}
	assert.Equal(t, 2, len(bookmarks))

	bookmarks, err = repo.GetBookmarksByName("Folder", userName)
	if err != nil {
		t.Errorf(errStr1, err)
	}
	assert.Equal(t, 1, len(bookmarks))

	bookmarks, err = repo.GetBookmarksByName("o", userName)
	if err != nil {
		t.Errorf(errStr1, err)
	}
	assert.Equal(t, 3, len(bookmarks))
}

func TestGetAllPath(t *testing.T) {
	repo, db := repository(t)
	defer db.Close()

	userName := "username"

	if _, err := repo.Create(store.Bookmark{
		DisplayName: "Z",
		Path:        "/",
		SortOrder:   0,
		Type:        store.Folder,
		UserName:    userName,
	}); err != nil {
		t.Errorf("Could not create bookmarks: %v", err)
	}

	if _, err := repo.Create(store.Bookmark{
		DisplayName: "A",
		Path:        "/",
		SortOrder:   0,
		Type:        store.Folder,
		UserName:    userName,
	}); err != nil {
		t.Errorf("Could not create bookmarks: %v", err)
	}

	if _, err := repo.Create(store.Bookmark{
		DisplayName: "B",
		Path:        "/A",
		SortOrder:   0,
		Type:        store.Folder,
		UserName:    userName,
	}); err != nil {
		t.Errorf("Could not create bookmarks: %v", err)
	}

	if _, err := repo.Create(store.Bookmark{
		DisplayName: "A",
		Path:        "/Z",
		SortOrder:   0,
		Type:        store.Folder,
		UserName:    userName,
	}); err != nil {
		t.Errorf("Could not create bookmarks: %v", err)
	}

	paths, err := repo.GetAllPaths(userName)
	if err != nil {
		t.Errorf("cannot get all paths: %v", err)
	}

	assert.Equal(t, 5, len(paths))
}

func TestGetFolderByPath(t *testing.T) {
	repo, db := repository(t)
	defer db.Close()

	userName := "username"

	if _, err := repo.Create(store.Bookmark{
		DisplayName: "Z",
		Path:        "/",
		SortOrder:   0,
		Type:        store.Folder,
		UserName:    userName,
	}); err != nil {
		t.Errorf("Could not create bookmarks: %v", err)
	}

	folder, err := repo.GetFolderByPath("/Z", userName)
	if err != nil {
		t.Errorf("cannot get folder for path '%s': %v", "/Z", err)
	}
	assert.Equal(t, "/", folder.Path)
	assert.Equal(t, "Z", folder.DisplayName)

	_, err = repo.GetFolderByPath("/", userName)
	if err == nil {
		t.Errorf("expected error for path '/'")
	}

	_, err = repo.GetFolderByPath("", userName)
	if err == nil {
		t.Errorf("expected error for path ''")
	}
}

func TestConcurrentWriteForConnection(t *testing.T) {
	// span a couple of go-routines to create bookmarks concurrently
	// this should work without any busy-connection errors from sqlite: https://www.sqlite.org/rescode.html#busy-connection
	var (
		wg sync.WaitGroup
	)
	// use a random filename
	f, err := os.CreateTemp("/tmp", "*.db")
	if err != nil {
		t.Fatalf("cannot create random filename")
	}
	fName := f.Name()
	f.Close()

	repo, db := repositoryFile(fName, t)
	defer func() {
		db.Close()
		os.Remove(fName)
	}()

	// spin-up a couple of go-routines and create bookmarks in parallel
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			var index = i

			item := store.Bookmark{
				DisplayName:        fmt.Sprintf("displayName%d", index),
				Path:               "/",
				SortOrder:          0,
				Type:               store.Node,
				URL:                "http://url",
				UserName:           "random",
				InvertFaviconColor: 1,
			}

			errGo := repo.InUnitOfWork(func(r store.BookmarkRepository) error {
				_, err = r.Create(item)
				return err
			})

			if errGo != nil {
				t.Errorf("could not create bookmark: %v", errGo)
			}
			v := rand.Intn(500-100) + 100
			time.Sleep(time.Duration(v) * time.Millisecond)

			wg.Done()
		}()
	}
	wg.Wait()
	if err != nil {
		t.Errorf("could not create the bookmarks: %v", err)
	}
}

func TestBookmarksReferencingSameFavicon(t *testing.T) {
	repo, db := repository(t)
	defer db.Close()

	userName := "userName"
	favicon := "favicon"

	folder1 := store.Bookmark{
		DisplayName: "Folder1",
		Path:        "/",
		Favicon:     favicon,
		Type:        store.Folder,
		UserName:    userName,
	}
	bm, err := repo.Create(folder1)
	if err != nil {
		t.Errorf("Could not create bookmarks: %v", err)
	}
	assert.NotEmpty(t, bm.ID)

	folder2 := store.Bookmark{
		DisplayName: "Folder2",
		Path:        "/",
		Favicon:     favicon,
		Type:        store.Folder,
		UserName:    userName,
	}
	bm, err = repo.Create(folder2)
	if err != nil {
		t.Errorf("Could not create bookmarks: %v", err)
	}
	assert.NotEmpty(t, bm.ID)

	folder3 := store.Bookmark{
		DisplayName: "Folder3",
		Path:        "/",
		Type:        store.Folder,
		UserName:    userName,
	}
	bm, err = repo.Create(folder3)
	if err != nil {
		t.Errorf("Could not create bookmarks: %v", err)
	}
	assert.NotEmpty(t, bm.ID)

	num, err := repo.NumBookmarksReferencingFavicon(favicon, userName)
	if err != nil {
		t.Errorf("could not count favicons; %v", err)
	}
	if num != 2 {
		t.Error("expected 2 items to reference the favicon")
	}

	num, err = repo.NumBookmarksReferencingFavicon("NO_ID", userName)
	if err != nil {
		t.Errorf("could not count favicons; %v", err)
	}
	if num == 2 {
		t.Error("expected 0 items to reference the favicon")
	}

}
