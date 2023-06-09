// Package store is responsible to interact with the storage backend used for bookmarks
// this is done by implementing a repository for the datbase
package store

import (
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"golang.binggl.net/monorepo/pkg/logging"
)

// BookmarkRepository defines methods to interact with a store
type BookmarkRepository interface {
	InUnitOfWork(fn func(repo BookmarkRepository) error) error
	Create(item Bookmark) (Bookmark, error)
	Update(item Bookmark) (Bookmark, error)
	Delete(item Bookmark) error
	DeletePath(path, username string) error

	GetAllBookmarks(username string) ([]Bookmark, error)
	GetBookmarksByPath(path, username string) ([]Bookmark, error)
	GetBookmarksByPathStart(path, username string) ([]Bookmark, error)
	GetBookmarksByName(name, username string) ([]Bookmark, error)
	GetPathChildCount(path, username string) ([]NodeCount, error)
	GetAllPaths(username string) ([]string, error)

	GetBookmarkByID(id, username string) (Bookmark, error)
	GetFolderByPath(path, username string) (Bookmark, error)
}

// CreateBookmarkRepo a new repository
func CreateBookmarkRepo(db *gorm.DB, logger logging.Logger) BookmarkRepository {
	return &dbBookmarkRepository{
		transient: db,
		shared:    nil,
		logger:    logger,
	}
}

// --------------------------------------------------------------------------
// Implementation
// --------------------------------------------------------------------------

type dbBookmarkRepository struct {
	transient *gorm.DB
	shared    *gorm.DB
	logger    logging.Logger
	sync.Mutex
}

// InUnitOfWork uses a transaction to execute the supplied function
func (r *dbBookmarkRepository) InUnitOfWork(fn func(repo BookmarkRepository) error) error {
	return r.con().Transaction(func(tx *gorm.DB) error {
		// be sure the stop recursion here
		if r.shared != nil {
			return fmt.Errorf("a shared connection/transaction is already available, will not start a new one")
		}
		return fn(&dbBookmarkRepository{
			transient: r.transient,
			shared:    tx, // the transaction is used as the shared connection
			logger:    r.logger,
		})
	})
}

// query data
// --------------------------------------------------------------------------

// GetAllBookmarks retrieves all available bookmarks for the given user
func (r *dbBookmarkRepository) GetAllBookmarks(username string) ([]Bookmark, error) {
	var bookmarks []Bookmark
	h := r.con().Order("sort_order").Order("display_name").Where(&Bookmark{UserName: username}).Find(&bookmarks)
	return bookmarks, h.Error
}

// GetBookmarksByPath return the bookmark elements which have the given path
func (r *dbBookmarkRepository) GetBookmarksByPath(path, username string) ([]Bookmark, error) {
	var bookmarks []Bookmark
	h := r.con().Order("sort_order").Order("display_name").Where(&Bookmark{
		UserName: username,
		Path:     path,
	}).Find(&bookmarks)
	return bookmarks, h.Error
}

// GetBookmarksByPathStart return the bookmark elements which path starts with
func (r *dbBookmarkRepository) GetBookmarksByPathStart(path, username string) ([]Bookmark, error) {
	var bookmarks []Bookmark
	h := r.con().
		Order("type desc").
		Order("sort_order").
		Order("display_name").
		Where("user_name = ? AND path LIKE ?", username, path+"%").Find(&bookmarks)
	return bookmarks, h.Error
}

// GetBookmarksByName searches for bookmarks by the given name
func (r *dbBookmarkRepository) GetBookmarksByName(name, username string) ([]Bookmark, error) {
	var bookmarks []Bookmark
	h := r.con().Order("sort_order").Order("display_name").
		Where("user_name = ? AND lower(display_name) LIKE ?", username, "%"+strings.ToLower(name)+"%").Find(&bookmarks)
	return bookmarks, h.Error
}

// GetBookmarkByID returns the bookmark specified by the given id - for the user
func (r *dbBookmarkRepository) GetBookmarkByID(id, username string) (Bookmark, error) {
	var bookmark Bookmark
	h := r.con().Where(&Bookmark{ID: id, UserName: username}).First(&bookmark)
	return bookmark, h.Error
}

// GetFolderByPath returns the bookmark folder elements specified by path
func (r *dbBookmarkRepository) GetFolderByPath(path, username string) (Bookmark, error) {
	var bookmark Bookmark

	// ROOT path is virtual
	if path == "/" {
		return Bookmark{}, fmt.Errorf("cannot get folder for path '%s'", path)
	}

	parent, folderName, ok := pathAndFolder(path)
	if !ok {
		return Bookmark{}, fmt.Errorf("could not get parent/folder of path '%s'", path)
	}

	h := r.con().Where(&Bookmark{
		UserName:    username,
		Path:        parent,
		DisplayName: folderName,
		Type:        Folder,
	}).First(&bookmark)

	return bookmark, h.Error
}

// GetPathChildCount returns the number of child-elements for a given path
func (r *dbBookmarkRepository) GetPathChildCount(path, username string) ([]NodeCount, error) {
	if path == "" {
		return nil, fmt.Errorf("no path supplied")
	}

	query := `SELECT i.path as path, count(i.id) as count FROM BOOKMARKS i WHERE i.path IN (
                %s
                ) GROUP BY i.path %s ORDER BY i.path`

	query = fmt.Sprintf(query, nativeHierarchyQuery, "HAVING i.path = ?")

	// Folder, Username, Path
	var nodes []NodeCount
	h := r.con().Raw(query, Folder, username, path).Scan(&nodes)
	if h.Error != nil {
		return nil, fmt.Errorf("could not get nodecount for path '%s': %v", path, h.Error)
	}

	return nodes, nil
}

// GetAllPaths returns all available paths for the given username
func (r *dbBookmarkRepository) GetAllPaths(username string) ([]string, error) {
	return r.availablePaths(username)
}

// modify data
// --------------------------------------------------------------------------

// Create is used to save a new bookmark entry
func (r *dbBookmarkRepository) Create(item Bookmark) (Bookmark, error) {
	var (
		err       error
		hierarchy []string
	)

	defer func() {
		// release any previously created locks
		r.Unlock()
	}()
	// block concurrent requests
	r.Lock()

	if item.Path == "" {
		return Bookmark{}, fmt.Errorf("path is empty")
	}

	if item.ID == "" {
		item.ID = uuid.New().String()
	}
	item.Created = time.Now().UTC()

	r.logger.Debug(fmt.Sprintf("create new bookmark item: %+v", item))

	// if we create a new bookmark item using a specific path we need to ensure that
	// the parent-path is available. as this is a hierarchical structure this is quite tedious
	// the solution is to query the whole hierarchy and check if the given path is there

	if item.Path != "/" {
		hierarchy, err = r.availablePaths(item.UserName)
		if err != nil {
			return Bookmark{}, err
		}
		found := false
		for _, h := range hierarchy {
			if h == item.Path {
				found = true
				break
			}
		}
		if !found {
			r.logger.Error(fmt.Sprintf("cannot create the bookmark '%+v' because the parent path '%s' is not available!", item, item.Path))
			return Bookmark{}, fmt.Errorf("cannot create item because of missing path hierarchy '%s'", item.Path)
		}
	}

	if h := r.con().Create(&item); h.Error != nil {
		return Bookmark{}, h.Error
	}

	// this entry (either node or folder) was created with a given path. increment the number of child-elements
	// for this given path, and update the "parent" directory entry.
	// exception: if the path is ROOT, '/' no update needs to be done, because no dedicated ROOT, '/' entry
	if item.Path != "/" {
		err = r.calcChildCount(item.Path, item.UserName, func(c int) int {
			return c + 1
		})
		if err != nil {
			return Bookmark{}, fmt.Errorf("could not update the child-count for '%s': %v", item.Path, err)
		}
	}

	return item, nil
}

// Update changes an existing bookmark item
func (r *dbBookmarkRepository) Update(item Bookmark) (Bookmark, error) {
	var (
		err       error
		hierarchy []string
		bm        Bookmark
	)

	defer func() {
		// release any previously created locks
		r.Unlock()
	}()
	// block concurrent requests
	r.Lock()

	if item.Path == "" {
		return Bookmark{}, fmt.Errorf("path is empty")
	}

	h := r.con().Where(&Bookmark{ID: item.ID, UserName: item.UserName}).First(&bm)
	if h.Error != nil {
		return Bookmark{}, fmt.Errorf("cannot get bookmark by id '%s': %v", item.ID, h.Error)
	}

	r.logger.Debug(fmt.Sprintf("update bookmark item: %+v", item))

	// if we create a new bookmark item using a specific path we need to ensure that
	// the parent-path is available. as this is a hierarchical structure this is quite tedious
	// the solution is to query the whole hierarchy and check if the given path is there

	if item.Path != "/" {
		hierarchy, err = r.availablePaths(item.UserName)
		if err != nil {
			return Bookmark{}, err
		}
		found := false
		for _, h := range hierarchy {
			if h == item.Path {
				found = true
				break
			}
		}
		if !found {
			r.logger.Error(fmt.Sprintf("cannot update the bookmark '%+v' because the parent path '%s' is not available!", item, item.Path))
			return Bookmark{}, fmt.Errorf("cannot update item because of missing path hierarchy '%s'", item.Path)
		}
	}

	now := time.Now().UTC()
	bm.Modified = &now
	bm.DisplayName = item.DisplayName
	bm.Path = item.Path
	bm.SortOrder = item.SortOrder
	bm.URL = item.URL
	bm.Favicon = item.Favicon
	bm.Highlight = item.Highlight
	bm.ChildCount = item.ChildCount
	bm.InvertFaviconColor = item.InvertFaviconColor

	h = r.con().Save(&bm)
	if h.Error != nil {
		return Bookmark{}, fmt.Errorf("cannot update bookmark with id '%s': %v", item.ID, h.Error)
	}
	return bm, nil
}

// Delete removes the bookmark identified by id
func (r *dbBookmarkRepository) Delete(item Bookmark) error {
	var (
		bm  Bookmark
		err error
	)

	defer func() {
		// release any previously created locks
		r.Unlock()
	}()
	// block concurrent requests
	r.Lock()

	h := r.con().Where(&Bookmark{ID: item.ID, UserName: item.UserName}).First(&bm)
	if h.Error != nil {
		return fmt.Errorf("cannot get bookmark by id '%s': %v", item.ID, h.Error)
	}

	r.logger.Debug(fmt.Sprintf("delete bookmark item: %+v", item))

	// one item is removed from a given path, decrement the child-count for
	// the folder / path this item is located in
	if item.Path != "/" {
		err = r.calcChildCount(item.Path, item.UserName, func(c int) int {
			return c - 1
		})
		if err != nil {
			return fmt.Errorf("could not update the child-count for '%s': %v", item.Path, err)
		}
	}

	h = r.con().Delete(&bm)
	if h.Error != nil {
		return fmt.Errorf("cannot delete bookmark by id '%s': %v", item.ID, h.Error)
	}
	return nil
}

// DeletePath removes all bookmarks having the same path
func (r *dbBookmarkRepository) DeletePath(path, username string) error {

	if path == "" {
		return fmt.Errorf("path is empty")
	}
	if path == "/" {
		return fmt.Errorf("cannot delete the root path")
	}

	defer func() {
		// release any previously created locks
		r.Unlock()
	}()
	// block concurrent requests
	r.Lock()

	h := r.con().Where("user_name = ? AND path LIKE ?", username, path+"%").Delete(Bookmark{})
	if h.Error != nil {
		return fmt.Errorf("no bookmarks available for path '%s': %v", path, h.Error)
	}

	folder, err := r.GetFolderByPath(path, username)
	if err != nil {
		return fmt.Errorf("could not get folder of given path '%s'", path)
	}
	h = r.con().Delete(&folder)
	if h.Error != nil {
		return fmt.Errorf("cannot delete folder '%s': %v", path, h.Error)
	}

	// entries with path /pa/th were deleted - also the folder /pa - th needs to be deleted
	parentPath, _, ok := pathAndFolder(path)
	if !ok {
		return fmt.Errorf("could not get parent-path/folder of given path '%s'", path)
	}

	if parentPath != "/" {
		parentFolder, err := r.GetFolderByPath(parentPath, username)
		if err != nil {
			return fmt.Errorf("could not get folder of given path '%s'", parentPath)
		}
		nodes, err := r.GetPathChildCount(parentPath, username)
		if err != nil {
			return fmt.Errorf("could not get child-count of given path '%s'", parentPath)
		}

		var count int
		if len(nodes) > 0 {
			count = nodes[0].Count
		}

		return r.updateChildCount(&parentFolder, count)
	}

	return nil
}

// --------------------------------------------------------------------------
// internal logic / helpers
// --------------------------------------------------------------------------

func (r *dbBookmarkRepository) con() *gorm.DB {
	if r.shared != nil {
		return r.shared
	}
	if r.transient == nil {
		panic("no database connection is available")
	}
	return r.transient
}

func (r *dbBookmarkRepository) updateChildCount(folder *Bookmark, count int) error {
	if h := r.con().Model(folder).Updates(
		map[string]interface{}{"child_count": count, "modified": time.Now().UTC()}); h.Error != nil {
		return fmt.Errorf("cannot update item '%+v': %v", *folder, h.Error)
	}
	return nil
}

const nativeHierarchyQuery = `SELECT '/' as path

UNION ALL

SELECT a.path || '/' || a.display_name FROM (

    SELECT
        CASE ii.path
            WHEN '/' THEN ''
            ELSE ii.path
        END AS path, ii.display_name
    FROM BOOKMARKS ii WHERE
        ii.type = ? AND ii.user_name = ?
) a
GROUP BY a.path || '/' || a.display_name`

func (r *dbBookmarkRepository) availablePaths(username string) (paths []string, err error) {
	var (
		rows *sql.Rows
	)

	rows, err = r.con().Raw(nativeHierarchyQuery, Folder, username).Rows() // (*sql.Rows, error)
	defer func(ro *sql.Rows) {
		if ro != nil {
			err = ro.Close()
		}
	}(rows)

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var path string
		if err = rows.Scan(&path); err != nil {
			return nil, err
		}
		paths = append(paths, path)
	}
	return paths, nil
}

func (r *dbBookmarkRepository) calcChildCount(path, username string, fn func(i int) int) error {
	// the supplied path is of the form
	// /A/B/C => get the entry C (which is a folder) and inc/dec the child-count
	parentPath, parentName, ok := pathAndFolder(path)
	if !ok {
		return fmt.Errorf("invalid path encountered '%s'", path)
	}
	var bm Bookmark
	if h := r.con().Where(&Bookmark{
		UserName:    username,
		Path:        parentPath,
		Type:        Folder,
		DisplayName: parentName}).First(&bm); h.Error != nil {
		return fmt.Errorf("could not get parent item '%s, %s'", parentPath, parentName)
	}

	// update the found item
	count := fn(bm.ChildCount)
	return r.updateChildCount(&bm, count)
}

func pathAndFolder(fullPath string) (path string, folder string, valid bool) {
	i := strings.LastIndex(fullPath, "/")
	if i == -1 {
		return
	}

	parent := fullPath[0:i]
	if i == 0 || parent == "" {
		parent = "/"
	}

	name := fullPath[i+1:]

	return parent, name, true
}
