package store

import (
	"database/sql"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"golang.binggl.net/monorepo/pkg/logging"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const expectations = "there were unfulfilled expectations: %s"

var logger = logging.NewNop()

func mockRepository() (Repository, sqlmock.Sqlmock, error) {
	var (
		db   *sql.DB
		DB   *gorm.DB
		err  error
		mock sqlmock.Sqlmock
	)
	if db, mock, err = sqlmock.New(); err != nil {
		return nil, nil, err
	}
	if DB, err = gorm.Open(mysql.New(mysql.Config{Conn: db, SkipInitializeWithVersion: true}), &gorm.Config{}); err != nil {
		return nil, nil, err
	}
	return Create(DB, logger), mock, nil
}

func repository(t *testing.T) (Repository, *sql.DB) {
	var (
		DB  *gorm.DB
		err error
	)
	if DB, err = gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{}); err != nil {
		t.Fatalf("cannot create database connection: %v", err)
	}
	// Migrate the schema
	DB.AutoMigrate(&Bookmark{})
	db, err := DB.DB()
	if err != nil {
		t.Fatalf("could not get DB handle; %v", err)
	}
	return Create(DB, logger), db
}

func Test_Mock_GetAllBookmarks(t *testing.T) {
	repo, mock, err := mockRepository()
	if err != nil {
		t.Fatalf("Could not create Repository: %v", err)
	}

	userName := "test"
	now := time.Now().UTC()
	rowDef := []string{"id", "path", "display_name", "url", "sort_order", "type", "user_name", "created", "modified", "child_count", "access_count", "favicon"}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `BOOKMARKS` WHERE")).
		WithArgs(userName).
		WillReturnRows(sqlmock.NewRows(rowDef).
			AddRow("id", "path", "display_name", "url", 0, 0, userName, now, nil, 0, 0, ""))

	bookmarks, err := repo.GetAllBookmarks(userName)
	if err != nil {
		t.Errorf("Could not get bookmarks: %v", err)
	}
	if len(bookmarks) != 1 {
		t.Errorf("Invalid number of bookmarks returned: %d", len(bookmarks))
	}

	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf(expectations, err)
	}
}

func Test_Store_Availability(t *testing.T) {
	repo, db := repository(t)
	defer db.Close()

	err := repo.CheckStoreConnectivity(1000)
	assert.Nil(t, err, "store is not available")
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
	item := Bookmark{
		DisplayName: "displayName",
		Path:        "/",
		SortOrder:   0,
		Type:        Node,
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
	folder := Bookmark{
		DisplayName: "Folder",
		Path:        "/",
		SortOrder:   0,
		Type:        Folder,
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

	node := Bookmark{
		DisplayName: "Node",
		Path:        "/Folder",
		SortOrder:   0,
		Type:        Node,
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
	assert.Equal(t, Node, node.Type)
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

	folder := Bookmark{
		DisplayName: "Folder",
		Path:        "/",
		SortOrder:   0,
		Type:        Folder,
		UserName:    userName,
	}

	id := ""
	err := repo.InUnitOfWork(func(r Repository) error {
		bm, err := r.Create(folder)
		if err != nil {
			return err
		}
		assert.NotEmpty(t, bm.ID)
		id = bm.ID
		folder, err := r.GetBookmarkByID(bm.ID, bm.UserName)
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
	folderRollback := Bookmark{
		DisplayName: "Folder",
		Path:        "/",
		SortOrder:   0,
		Type:        Folder,
		UserName:    userName,
	}
	err = repo.InUnitOfWork(func(r Repository) error {
		bm, err := r.Create(folderRollback)
		if err != nil {
			return err
		}
		assert.NotEmpty(t, bm.ID)
		id = bm.ID
		folder, err := r.GetBookmarkByID(bm.ID, bm.UserName)
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
		t.Errorf("expected an error because item not availabel!")
	}
}

func TestUpdateBookmark(t *testing.T) {
	repo, db := repository(t)
	defer db.Close()

	userName := "username"
	item := Bookmark{
		DisplayName: "displayName",
		Path:        "/",
		SortOrder:   0,
		Type:        Node,
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
	f := Bookmark{
		DisplayName: "Folder",
		Path:        "/",
		SortOrder:   0,
		Type:        Folder,
		UserName:    userName,
	}
	folder, err := repo.Create(f)
	if err != nil {
		t.Errorf("Could not create bookmarks: %v", err)
	}
	assert.NotEmpty(t, folder.ID)

	n := Bookmark{
		DisplayName: "Node",
		Path:        "/Folder",
		SortOrder:   0,
		Type:        Node,
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
	_, err := repo.Create(Bookmark{
		Type:        Folder,
		DisplayName: "Folder",
		UserName:    userName,
		Path:        "/",
	})
	if err != nil {
		t.Errorf(errStr, err)
	}
	_, err = repo.Create(Bookmark{
		Type:        Folder,
		DisplayName: "Folder1",
		UserName:    userName,
		Path:        "/",
	})
	if err != nil {
		t.Errorf(errStr, err)
	}
	_, err = repo.Create(Bookmark{
		Type:        Node,
		DisplayName: "Node",
		UserName:    userName,
		Path:        "/Folder",
		URL:         "http://url",
	})
	if err != nil {
		t.Errorf(errStr, err)
	}
	_, err = repo.Create(Bookmark{
		Type:        Node,
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
	_, err := repo.Create(Bookmark{
		Type:        Folder,
		DisplayName: "Folder",
		UserName:    userName,
		Path:        "/",
	})
	if err != nil {
		t.Errorf(errStr, err)
	}
	_, err = repo.Create(Bookmark{
		Type:        Folder,
		DisplayName: "Folder1",
		UserName:    userName,
		Path:        "/",
	})
	if err != nil {
		t.Errorf(errStr, err)
	}
	_, err = repo.Create(Bookmark{
		Type:        Node,
		DisplayName: "Node",
		UserName:    userName,
		Path:        "/Folder",
		URL:         "http://url",
	})
	if err != nil {
		t.Errorf(errStr, err)
	}
	_, err = repo.Create(Bookmark{
		Type:        Node,
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
	folder, err := repo.Create(Bookmark{
		Type:        Folder,
		DisplayName: "Folder",
		UserName:    userName,
		Path:        "/",
	})
	if err != nil {
		t.Errorf(errStr, err)
	}
	_, err = repo.Create(Bookmark{
		Type:        Folder,
		DisplayName: "Folder1",
		UserName:    userName,
		Path:        "/Folder",
	})
	if err != nil {
		t.Errorf(errStr, err)
	}
	_, err = repo.Create(Bookmark{
		Type:        Node,
		DisplayName: "Node",
		UserName:    userName,
		Path:        "/Folder",
		URL:         "http://url",
	})
	if err != nil {
		t.Errorf(errStr, err)
	}
	_, err = repo.Create(Bookmark{
		Type:        Node,
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
	_, err := repo.Create(Bookmark{
		Type:        Folder,
		DisplayName: "Folder",
		UserName:    userName,
		Path:        "/",
	})
	if err != nil {
		t.Errorf(errStr, err)
	}
	_, err = repo.Create(Bookmark{
		Type:        Folder,
		DisplayName: "Folder1",
		UserName:    userName,
		Path:        "/Folder",
	})
	if err != nil {
		t.Errorf(errStr, err)
	}
	_, err = repo.Create(Bookmark{
		Type:        Folder,
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
	_, err := repo.Create(Bookmark{
		Type:        Folder,
		DisplayName: "Folder",
		UserName:    userName,
		Path:        "/",
	})
	if err != nil {
		t.Errorf(errStr, err)
	}
	_, err = repo.Create(Bookmark{
		Type:        Node,
		DisplayName: "Node",
		UserName:    userName,
		Path:        "/Folder",
	})
	if err != nil {
		t.Errorf(errStr, err)
	}
	_, err = repo.Create(Bookmark{
		Type:        Node,
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

func TestGetMostRecentBookmarks(t *testing.T) {
	repo, db := repository(t)
	defer db.Close()
	errStr := "Could not create bookmarks: %v"

	userName := "userName"

	// create a folder structure
	_, err := repo.Create(Bookmark{
		Type:        Node,
		DisplayName: "Node",
		UserName:    userName,
		Path:        "/",
		AccessCount: 5,
	})
	if err != nil {
		t.Errorf(errStr, err)
	}
	_, err = repo.Create(Bookmark{
		Type:        Node,
		DisplayName: "Node1",
		UserName:    userName,
		Path:        "/",
		AccessCount: 10,
	})
	if err != nil {
		t.Errorf(errStr, err)
	}
	_, err = repo.Create(Bookmark{
		Type:        Node,
		DisplayName: "Node2",
		UserName:    userName,
		Path:        "/",
	})
	if err != nil {
		t.Errorf(errStr, err)
	}

	all, err := repo.GetAllBookmarks(userName)
	if err != nil {
		t.Errorf("could not get bookmarks: %v", err)
	}
	assert.Equal(t, 3, len(all))

	recent, err := repo.GetMostRecentBookmarks(userName, 10)
	if err != nil {
		t.Errorf("could not get bookmarks: %v", err)
	}
	assert.Equal(t, 2, len(recent))
	assert.Equal(t, "Node1", recent[0].DisplayName)
	assert.Equal(t, "Node", recent[1].DisplayName)

	recent, err = repo.GetMostRecentBookmarks(userName, 1)
	if err != nil {
		t.Errorf("could not get bookmarks: %v", err)
	}
	assert.Equal(t, 1, len(recent))
	assert.Equal(t, "Node1", recent[0].DisplayName)
}

func TestGetAllPath(t *testing.T) {
	repo, db := repository(t)
	defer db.Close()

	userName := "username"

	if _, err := repo.Create(Bookmark{
		DisplayName: "Z",
		Path:        "/",
		SortOrder:   0,
		Type:        Folder,
		UserName:    userName,
	}); err != nil {
		t.Errorf("Could not create bookmarks: %v", err)
	}

	if _, err := repo.Create(Bookmark{
		DisplayName: "A",
		Path:        "/",
		SortOrder:   0,
		Type:        Folder,
		UserName:    userName,
	}); err != nil {
		t.Errorf("Could not create bookmarks: %v", err)
	}

	if _, err := repo.Create(Bookmark{
		DisplayName: "B",
		Path:        "/A",
		SortOrder:   0,
		Type:        Folder,
		UserName:    userName,
	}); err != nil {
		t.Errorf("Could not create bookmarks: %v", err)
	}

	if _, err := repo.Create(Bookmark{
		DisplayName: "A",
		Path:        "/Z",
		SortOrder:   0,
		Type:        Folder,
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

	if _, err := repo.Create(Bookmark{
		DisplayName: "Z",
		Path:        "/",
		SortOrder:   0,
		Type:        Folder,
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
