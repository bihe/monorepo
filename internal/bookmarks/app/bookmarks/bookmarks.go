package bookmarks

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"golang.binggl.net/monorepo/internal/bookmarks/app"
	"golang.binggl.net/monorepo/internal/bookmarks/app/favicon"
	"golang.binggl.net/monorepo/internal/bookmarks/app/store"
	"golang.binggl.net/monorepo/internal/common/upload"
	"golang.binggl.net/monorepo/pkg/logging"
	"golang.binggl.net/monorepo/pkg/security"
)

// Application implements the bookmarks logic
type Application struct {
	// Logger instance to use for structured logging
	Logger logging.Logger
	// BookmarkStore defines a bookmarks repository
	BookmarkStore store.BookmarkRepository
	// FavStore specifies a Favicon repository
	FavStore store.FaviconRepository
	// FaviconPath defines a path on disk to store a favicon
	FaviconPath string
	// FileStore persists files
	FileStore store.FileRepository
	// UploadSvc manages file uploads
	UploadSvc upload.Service
}

// ExistingFavicon is used to prefix a favicon ID which is already available
const ExistingFavicon = "existing://"

// CreateBookmark stores a new bookmark
func (s *Application) CreateBookmark(bm Bookmark, user security.User) (*Bookmark, error) {
	var (
		savedItem store.Bookmark
		t         store.NodeType
		err       error
		fileID    *string
	)

	if bm.Path == "" || bm.DisplayName == "" {
		s.Logger.Error("required fields of bookmarks missing (Path or DisplayName)")
		return nil, app.ErrValidation("invalid request data supplied, missing Path or DisplayName")
	}

	switch bm.Type {
	case Node:
		if bm.URL == "" {
			return nil, app.ErrValidation("invalid request data supplied, missing URL for bookmark")
		}
		if bm.FileID != "" {
			return nil, app.ErrValidation("invalid request data supplied, a node does not have a file")
		}
		t = store.Node
	case Folder:
		t = store.Folder
	case FileItem:
		if bm.FileID == "" {
			return nil, app.ErrValidation("invalid request data supplied, missing File for bookmark")
		}
		t = store.FileItem
	}

	// when a favicon is supplied, fetch the local temp favicon with the given id
	// store the favicon in the repository and remove the old stored favicon
	if bm.Favicon != "" {
		bm.Favicon, err = s.saveFavicon(bm.Favicon)
		if err != nil {
			return nil, err
		}
	}

	// take care if a file was uploaded. the id provided is a token to fetch/store the binary payload
	if bm.FileID != "" {
		savedFileId, err := s.processFile(bm.FileID)
		if err != nil {
			return nil, err
		}
		fileID = &savedFileId
	}

	if err := s.BookmarkStore.InUnitOfWork(func(repo store.BookmarkRepository) error {
		item, err := repo.Create(store.Bookmark{
			DisplayName:        bm.DisplayName,
			Path:               bm.Path,
			Type:               t,
			URL:                bm.URL,
			UserName:           user.Username,
			Favicon:            bm.Favicon,
			SortOrder:          bm.SortOrder,
			Highlight:          bm.Highlight,
			InvertFaviconColor: bm.InvertFaviconColor,
			FileID:             fileID,
		})
		if err != nil {
			return err
		}
		savedItem = item
		return nil
	}); err != nil {
		s.Logger.Error(fmt.Sprintf("could not create a new bookmark: %v", err))
		return nil, fmt.Errorf("error creating a new bookmark: %v", err)
	}
	id := savedItem.ID
	s.Logger.Info(fmt.Sprintf("bookmark created with ID: %s", id))
	return entityToModel(savedItem), nil
}

// GetBookmarkByID retrieves a bookmark for the given user
func (s *Application) GetBookmarkByID(id string, user security.User) (*Bookmark, error) {
	if id == "" {
		return nil, app.ErrValidation("no id supplied to fetch bookmark")
	}
	var (
		bm  store.Bookmark
		err error
	)

	bm, err = s.BookmarkStore.GetBookmarkByID(id, user.Username)
	if err != nil {
		return nil, app.ErrNotFound(fmt.Sprintf("could not fetch bookmark; %v", err))
	}
	return entityToModel(bm), nil
}

// GetBookmarkPayloadByID returns the payload of a given bookmark
func (s *Application) GetBookmarkPayloadByID(id string, user security.User) ([]byte, error) {
	if id == "" {
		return nil, app.ErrValidation("no id supplied to fetch bookmark")
	}
	var (
		bm  store.Bookmark
		err error
	)

	bm, err = s.BookmarkStore.GetBookmarkByID(id, user.Username)
	if err != nil {
		return nil, app.ErrNotFound(fmt.Sprintf("could not fetch bookmark; %v", err))
	}

	if bm.File == nil {
		return nil, fmt.Errorf("cannot fetch payload, no file availabe")
	}
	if bm.File.FileObjectID == nil {
		return nil, fmt.Errorf("cannot fetch payload, no file availabe")
	}
	payloadID := *bm.File.FileObjectID
	file, err := s.FileStore.Get(payloadID)
	if err != nil {
		return nil, fmt.Errorf("could not fetch payload; %v", err)
	}
	if file.FileObject == nil {
		return nil, fmt.Errorf("cannot fetch payload, no file availabe")
	}

	return file.FileObject.Payload, nil
}

// GetBookmarksByPath retrieves bookmarks by a given path
func (s *Application) GetBookmarksByPath(path string, user security.User) ([]Bookmark, error) {
	if path == "" {
		return nil, app.ErrValidation("missing path")
	}

	s.Logger.Info(fmt.Sprintf("get bookmarks by path: '%s' for user: '%s'", path, user.Username))

	bms, err := s.BookmarkStore.GetBookmarksByPath(path, user.Username)
	if err != nil {
		s.Logger.Error(fmt.Sprintf("cannot get bookmark by path: '%s', %v", path, err))
	}
	return entityListToModel(bms), nil
}

// GetBookmarksFolderByPath returns the folder identified by the given path
func (s *Application) GetBookmarksFolderByPath(path string, user security.User) (*Bookmark, error) {
	if path == "" {
		return nil, app.ErrValidation("missing path parameter")
	}

	s.Logger.Info(fmt.Sprintf("get bookmarks-folder by path: '%s' for user: '%s'", path, user.Username))
	if path == "/" {
		// special treatment for the root path. This path is ALWAYS available
		// and does not have a specific storage entry - this is by convention
		return &Bookmark{
			DisplayName: "Root",
			Path:        "/",
			Type:        Folder,
			ID:          fmt.Sprintf("%s_ROOT", user.Username),
		}, nil
	}
	bm, err := s.BookmarkStore.GetFolderByPath(path, user.Username)
	if err != nil {
		s.Logger.Error(fmt.Sprintf("cannot get bookmark folder by path: '%s', %v", path, err))
		return nil, app.ErrNotFound(fmt.Sprintf("no folder for path '%s' found", path))
	}

	return entityToModel(bm), nil
}

// GetAllPaths returns all available stored paths
func (s *Application) GetAllPaths(user security.User) ([]string, error) {
	paths, err := s.BookmarkStore.GetAllPaths(user.Username)
	if err != nil {
		s.Logger.Error(fmt.Sprintf("cannot get all bookmark paths: %v", err))
		return make([]string, 0), fmt.Errorf("cannot get all bookmark paths")
	}
	return paths, nil
}

// GetBookmarksByName returns the given bookmark(s) by the supplied name
func (s *Application) GetBookmarksByName(name string, user security.User) ([]Bookmark, error) {
	if name == "" {
		return make([]Bookmark, 0), app.ErrValidation("missing name parameter")
	}

	s.Logger.Info(fmt.Sprintf("get bookmarks by name: '%s' for user: '%s'", name, user.Username))
	bms, err := s.BookmarkStore.GetBookmarksByName(name, user.Username)
	var bookmarks []Bookmark
	if err != nil {
		s.Logger.Error(fmt.Sprintf("cannot get bookmark by name: '%s', %v", name, err))
	}
	bookmarks = entityListToModel(bms)
	return bookmarks, nil
}

// FetchAndForward retrieves a bookmark by id and forwards to the url of the bookmark
func (s *Application) FetchAndForward(id string, user security.User) (string, error) {
	if id == "" {
		return "", app.ErrValidation("missing id parameter")
	}

	s.Logger.Info(fmt.Sprintf("try to fetch bookmark with ID '%s'", id))
	redirectURL := ""
	if err := s.BookmarkStore.InUnitOfWork(func(repo store.BookmarkRepository) error {
		existing, err := repo.GetBookmarkByID(id, user.Username)
		if err != nil {
			s.Logger.Error(fmt.Sprintf("could not find bookmark by id '%s': %v", id, err))
			return err
		}

		if existing.Type == store.Folder {
			s.Logger.Error(fmt.Sprintf("accessCount and redirect only valid for Nodes: ID '%s'", id))
			return app.ErrValidation(fmt.Sprintf("cannot fetch and forward folder - ID '%s'", id))
		}

		// once accessed the highlight flag is removed
		existing.Highlight = 0
		if _, err := repo.Update(existing); err != nil {
			s.Logger.Error(fmt.Sprintf("could not update bookmark '%s': %v", id, err))
			return err
		}
		redirectURL = existing.URL
		return nil
	}); err != nil {
		s.Logger.Error(fmt.Sprintf("could not fetch and update bookmark by ID '%s': %v", id, err))
		return "", fmt.Errorf("error fetching and updating bookmark: %w", err)
	}

	s.Logger.Info(fmt.Sprintf("will redirect to bookmark URL '%s'", redirectURL))
	return redirectURL, nil
}

// Delete a bookmark by id
func (s *Application) Delete(id string, user security.User) error {
	if id == "" {
		return app.ErrValidation("missing id parameter")
	}

	faviconId := ""
	s.Logger.Info(fmt.Sprintf("will try to delete bookmark with ID '%s'", id))
	if err := s.BookmarkStore.InUnitOfWork(func(repo store.BookmarkRepository) error {
		// 1) fetch the existing bookmark by id
		existing, err := repo.GetBookmarkByID(id, user.Username)
		if err != nil {
			s.Logger.Error(fmt.Sprintf("could not find bookmark by id '%s': %v", id, err))
			return err
		}
		faviconId = existing.Favicon

		// if the element is a folder and there are child-elements
		// prevent the deletion - this can only be done via a recursive deletion like rm -rf
		if existing.Type == store.Folder && existing.ChildCount > 0 {
			return fmt.Errorf("cannot delete folder '%s' because of existing child-elements %d",
				ensureFolderPath(existing.Path, existing.DisplayName), existing.ChildCount)
		}
		err = repo.Delete(existing)
		if err != nil {
			return err
		}

		return nil
	}); err != nil {
		s.Logger.Error(fmt.Sprintf("could not delete bookmark because of error: %v", err))
		return fmt.Errorf("error deleting bookmark: %v", err)
	}

	// when a favicon is available remove it, if it is the only remaining one
	if faviconId != "" {
		numFavicon, err := s.BookmarkStore.NumBookmarksReferencingFavicon(faviconId, user.Username)
		if err != nil {
			s.Logger.Error(fmt.Sprintf("could not check favicon references; %v", err))
			return fmt.Errorf("error counting favicon references; %v", err)
		}
		if numFavicon != 0 {
			return nil
		}
		err = s.FavStore.InUnitOfWork(func(favRepo store.FaviconRepository) error {
			obj, e := favRepo.Get(faviconId)
			if e != nil {
				s.Logger.Warn(fmt.Sprintf("the specified favicon '%s' was not available in the store", faviconId))
				// exit here, nothing more to do
				return nil
			}
			return favRepo.Delete(obj)
		})
		if err != nil {
			s.Logger.Error(fmt.Sprintf("got an error while trying to delete the favicon; %v", err))
			return fmt.Errorf("could not delete the specified favicon; %v", err)
		}
	}

	return nil
}

// DeletePath is used for folders to remove a whole "structure" of bookmarks
func (s *Application) DeletePath(id string, user security.User) error {
	if id == "" {
		return app.ErrValidation("missing id parameter")
	}

	faviconIDs := make([]string, 0)

	s.Logger.Info(fmt.Sprintf("will try to delete bookmark-path for ID '%s'", id))
	if err := s.BookmarkStore.InUnitOfWork(func(repo store.BookmarkRepository) error {
		// 1) fetch the existing bookmark by id
		existing, err := repo.GetBookmarkByID(id, user.Username)
		if err != nil {
			s.Logger.Error(fmt.Sprintf("could not find bookmark by id '%s': %v", id, err))
			return err
		}
		if existing.Type != store.Folder {
			s.Logger.Warn("this call is only applicable for folders")
			return fmt.Errorf("DeletePath is only possible for folders")
		}

		// 2) fetch all items which start with the given path
		startPath := ensureFolderPath(existing.Path, existing.DisplayName)
		bookmarks, err := repo.GetBookmarksByPathStart(startPath, user.Username)
		if err != nil {
			s.Logger.Error(fmt.Sprintf("could not find bookmark starting by path '%s': %v", startPath, err))
			return err
		}

		// 3) collect the favicon IDs to remove them later if necessary
		if existing.Favicon != "" {
			faviconIDs = append(faviconIDs, existing.Favicon)
		}
		for _, b := range bookmarks {
			if b.Favicon != "" {
				faviconIDs = append(faviconIDs, b.Favicon)
			}
		}

		// 4) delete the bookmark path
		err = repo.DeletePath(startPath, user.Username)
		if err != nil {
			s.Logger.Error(fmt.Sprintf("could not delete bookmark-path for '%s': %v", startPath, err))
			return err
		}

		return nil
	}); err != nil {
		s.Logger.Error(fmt.Sprintf("could not delete bookmark path because of error: %v", err))
		return fmt.Errorf("error deleting bookmark path: %v", err)
	}

	// favicon cleanup
	for _, fi := range faviconIDs {
		// when a favicon is available remove it, if it is the only remaining one
		numFavicon, err := s.BookmarkStore.NumBookmarksReferencingFavicon(fi, user.Username)
		if err != nil {
			s.Logger.Error(fmt.Sprintf("could not check favicon references; %v", err))
			return fmt.Errorf("error counting favicon references; %v", err)
		}
		if numFavicon != 0 {
			return nil
		}
		err = s.FavStore.InUnitOfWork(func(favRepo store.FaviconRepository) error {
			obj, e := favRepo.Get(fi)
			if e != nil {
				s.Logger.Warn(fmt.Sprintf("the specified favicon '%s' was not available in the store", fi))
				// exit here, nothing more to do
				return nil
			}
			return favRepo.Delete(obj)
		})
		if err != nil {
			s.Logger.Error(fmt.Sprintf("got an error while trying to delete the favicon; %v", err))
			return fmt.Errorf("could not delete the specified favicon; %v", err)
		}
	}

	return nil
}

// DeleteBookmarkFile removes an existing file stored with a bookmark
func (s *Application) DeleteBookmarkFile(bookmarkID string, user security.User) error {
	if bookmarkID == "" {
		return app.ErrValidation("missing bookmarkId parameter")
	}

	bm, err := s.BookmarkStore.GetBookmarkByID(bookmarkID, user.Username)
	if err != nil {
		return app.ErrNotFound(fmt.Sprintf("could not fetch bookmark; %v", err))
	}

	fileID := *bm.FileID
	s.Logger.Info(fmt.Sprintf("will try to delete file with ID '%s'", fileID))
	if err := s.FileStore.InUnitOfWork(func(repo store.FileRepository) error {
		existing, err := repo.Get(fileID)
		if err != nil {
			return fmt.Errorf("no file available with given id: %v", err)
		}
		err = repo.Delete(existing)
		if err != nil {
			return err
		}

		return nil
	}); err != nil {
		s.Logger.Error(fmt.Sprintf("could not delete bookmark because of error: %v", err))
		return fmt.Errorf("error deleting bookmark: %v", err)
	}
	return nil
}

// UpdateSortOrder modifies the display sort-order
func (s *Application) UpdateSortOrder(sort BookmarksSortOrder, user security.User) (int, error) {
	if len(sort.IDs) != len(sort.SortOrder) {
		return 0, app.ErrValidation(fmt.Sprintf("the number of IDs (%d) does not correspond the number of SortOrder entries (%d)", len(sort.IDs), len(sort.SortOrder)))
	}

	var updates int
	if err := s.BookmarkStore.InUnitOfWork(func(repo store.BookmarkRepository) error {
		for i, item := range sort.IDs {
			bm, err := repo.GetBookmarkByID(item, user.Username)
			if err != nil {
				s.Logger.Error(fmt.Sprintf("could not get bookmark by id '%s', %v", item, err))
				return err
			}
			s.Logger.Info(fmt.Sprintf("will update sortOrder of bookmark '%s' with value %d", bm.DisplayName, sort.SortOrder[i]))
			bm.SortOrder = sort.SortOrder[i]
			_, err = repo.Update(bm)
			if err != nil {
				s.Logger.Error(fmt.Sprintf("could not update bookmark: %v", err))
				return err
			}
			updates++
		}
		return nil
	}); err != nil {
		s.Logger.Error(fmt.Sprintf("could not update the sort-order for bookmark: %v", err))
		return 0, fmt.Errorf("error updating sort-order of bookmark: %v", err)
	}
	return updates, nil
}

// UpdateBookmark a bookmark
func (s *Application) UpdateBookmark(bm Bookmark, user security.User) (*Bookmark, error) {
	var (
		id     string
		item   store.Bookmark
		fileID *string
	)
	if bm.Path == "" || bm.DisplayName == "" || bm.ID == "" {
		s.Logger.Error("required fields of bookmarks missing")
		return nil, app.ErrValidation("invalid request data supplied, missing ID, Path or DisplayName")

	}

	// 1) fetch the existing bookmark by id
	existing, err := s.BookmarkStore.GetBookmarkByID(bm.ID, user.Username)
	if err != nil {
		s.Logger.Error(fmt.Sprintf("could not find bookmark by id '%s': %v", bm.ID, err))
		return nil, err
	}

	// when a favicon is supplied, fetch the local temp favicon with the given id
	// store the favicon in the repository and remove the old stored favicon
	favicon := bm.Favicon
	if favicon != "" && favicon != existing.Favicon {
		favicon, err = s.saveFavicon(bm.Favicon)
		if err != nil {
			return nil, err
		}
	} else {
		// if no favicon change was found re-use the existing favicon
		favicon = existing.Favicon
	}

	// if there is a new file during the update, process the new file
	existingFile := bm.FileID
	existingFileID := ""
	if existing.FileID != nil {
		existingFileID = *existing.FileID
	}
	if existingFile != "" && existingFile != existingFileID {
		savedFileId, err := s.processFile(bm.FileID)
		if err != nil {
			return nil, err
		}
		fileID = &savedFileId
	} else {
		fileID = existing.FileID
	}

	s.Logger.Info(fmt.Sprintf("will try to update existing bookmark entry: '%s (%s)'", bm.DisplayName, bm.ID))
	if err := s.BookmarkStore.InUnitOfWork(func(repo store.BookmarkRepository) error {
		childCount := existing.ChildCount
		if existing.Type == store.Folder {
			// 2) ensure that the existing folder is not moved to itself
			folderPath := ensureFolderPath(existing.Path, existing.DisplayName)
			if bm.Path == folderPath {
				s.Logger.Error(fmt.Sprintf("a folder cannot be moved into itself: folder-path: '%s', destination: '%s'", folderPath, bm.Path))
				return app.ErrValidation("cannot move folder into itself")
			}

			// 3) get the folder child-count
			// on save of a folder, update the child-count
			parentPath := existing.Path
			path := ensureFolderPath(parentPath, existing.DisplayName)
			nodeCount, err := repo.GetPathChildCount(path, user.Username)
			if err != nil {
				s.Logger.Error(fmt.Sprintf("could not get child-count of path '%s': %v", path, err))
				return err
			}
			if len(nodeCount) > 0 {
				for _, n := range nodeCount {
					if n.Path == path {
						childCount = n.Count
						break
					}
				}
			}
		}

		existingDisplayName := existing.DisplayName
		existingPath := existing.Path

		// 4) update the bookmark
		item, err = repo.Update(store.Bookmark{
			ID:                 bm.ID,
			Created:            existing.Created,
			DisplayName:        bm.DisplayName,
			Path:               bm.Path,
			Type:               existing.Type, // it does not make any sense to change the type of a bookmark!
			URL:                bm.URL,
			SortOrder:          bm.SortOrder,
			UserName:           user.Username,
			ChildCount:         childCount,
			Favicon:            favicon,
			Highlight:          bm.Highlight,
			InvertFaviconColor: bm.InvertFaviconColor,
			FileID:             fileID,
		})
		if err != nil {
			s.Logger.Error(fmt.Sprintf("could not update bookmark: %v", err))
			return err
		}
		id = item.ID

		if existing.Type == store.Folder && (existingDisplayName != bm.DisplayName || existingPath != bm.Path) {
			// if we have a folder and change the display-name or the parent-path, this also affects ALL sub-elements
			// therefore all paths of sub-elements where this folder-path is present, need to be updated
			newPath := ensureFolderPath(bm.Path, bm.DisplayName)
			oldPath := ensureFolderPath(existingPath, existingDisplayName)

			// if the folder is moved to the target-path we need to treat the special case, that
			// a folder with the same name might already exists in the target path.
			// without any logic we would end up with two folders
			// as we identify folders just by name (and not by ID) the logic below to update the
			// child-count updates the entries to the "new" folder (even though that is no update, as
			// we just deal with paths, which are the same /A/B vs /A/B)
			//
			// the "solution" is to get rid of one bookmark-folder. it is not important which one, because
			// we just address by path (/A/B).
			// the idea is to keep the folder which we are working on (this one) and remove the "other one"
			sameFolders, err := repo.GetBookmarksByPath(bm.Path, user.Username)
			if err != nil {
				s.Logger.Error(fmt.Sprintf("could not get bookmarks by path '%s': %v", oldPath, err))
				return err
			}
			sameFolderBookmarks := make([]store.Bookmark, 0)
			for _, f := range sameFolders {
				if f.Type == store.Folder && f.DisplayName == bm.DisplayName && f.ID != bm.ID {
					sameFolderBookmarks = append(sameFolderBookmarks, f)
				}
			}
			if len(sameFolderBookmarks) > 0 {
				// we have found folders with the same name, remove the folder to have only one remaining
				for _, f := range sameFolderBookmarks {
					err = repo.Delete(f)
					if err != nil {
						s.Logger.Error(fmt.Sprintf("could not merge folders '%s': %v", f.DisplayName, err))
						return err
					}
				}
			}

			s.Logger.Info(fmt.Sprintf("will update all old paths '%s' to new path '%s'", oldPath, newPath))
			bookmarks, err := repo.GetBookmarksByPathStart(oldPath, user.Username)
			if err != nil {
				s.Logger.Error(fmt.Sprintf("could not get bookmarks by path '%s': %v", oldPath, err))
				return err
			}

			// when we rename a bunch of items in bookmarks the sort-order is relevant
			// the bookmarks need to be sorted by path, otherwise an issue arises with "children" and the check
			// for the existence of the parent path
			sort.Slice(bookmarks, func(i, j int) bool {
				return bookmarks[i].Path < bookmarks[j].Path
			})

			for _, updateBm := range bookmarks {
				updatePath := strings.ReplaceAll(updateBm.Path, oldPath, newPath)
				if _, err := repo.Update(store.Bookmark{
					ID:                 updateBm.ID,
					Created:            updateBm.Created,
					DisplayName:        updateBm.DisplayName,
					Path:               updatePath,
					SortOrder:          updateBm.SortOrder,
					Type:               updateBm.Type,
					URL:                updateBm.URL,
					UserName:           user.Username,
					ChildCount:         updateBm.ChildCount,
					Highlight:          updateBm.Highlight,
					Favicon:            updateBm.Favicon,
					InvertFaviconColor: updateBm.InvertFaviconColor,
					FileID:             updateBm.FileID,
				}); err != nil {
					s.Logger.Error(fmt.Sprintf("cannot update bookmark path: %v", err))
					return err
				}
			}

			// after the update above, correct the child-count of this folder
			if err := s.updateChildCountOfPath(newPath, user.Username, repo); err != nil {
				s.Logger.Error(fmt.Sprintf("could not update child-count of folder-path '%s': %v", newPath, err))
				return err
			}
		}

		// if the path has changed - update the child-count of affected paths
		if existingPath != bm.Path {
			// the affected paths are the origin-path and the destination-path
			if err := s.updateChildCountOfPath(existingPath, user.Username, repo); err != nil {
				s.Logger.Error(fmt.Sprintf("could not update child-count of origin-path '%s': %v", existingPath, err))
				return err
			}
			// and the destination-path
			if err := s.updateChildCountOfPath(bm.Path, user.Username, repo); err != nil {
				s.Logger.Error(fmt.Sprintf("could not update child-count of destination-path '%s': %v", bm.Path, err))
				return err
			}
		}

		return nil
	}); err != nil {
		s.Logger.Error(fmt.Sprintf("could not update bookmark because of error: %v", err))
		return nil, fmt.Errorf("error updating bookmark: %w", err)
	}

	s.Logger.Info(fmt.Sprintf("updated bookmark with ID '%s'", id))
	return entityToModel(item), nil
}

func (s *Application) updateChildCountOfPath(path, username string, repo store.BookmarkRepository) error {
	if path == "/" {
		s.Logger.Info("skip the ROOT path '/'")
		return nil
	}

	folder, err := repo.GetFolderByPath(path, username)
	if err != nil {
		s.Logger.Error(fmt.Sprintf("cannot get the folder for path '%s': %v", path, err))
		return err
	}

	nodeCount, err := repo.GetPathChildCount(path, username)
	if err != nil {
		s.Logger.Error(fmt.Sprintf("could not get the child-count of the path '%s': %v", path, err))
		return err
	}

	var childCount int
	if len(nodeCount) > 0 {
		// use the first element, because the count was queried for a specific path
		childCount = nodeCount[0].Count
	}

	folder.ChildCount = childCount

	if _, err := repo.Update(folder); err != nil {
		s.Logger.Error(fmt.Sprintf("could not update the child-count of folder '%s': %v", path, err))
		return err
	}
	return nil
}

// ---- Favicon Logic ----
var defaultFaviconModTime = time.Date(2022, 12, 31, 0, 0, 0, 0, time.UTC)

const defaultFaviconName = "default_bookmark_favicon.svg"

// GetBookmarkFavicon returns the payload of the found favicon or the default favicon
func (s *Application) GetBookmarkFavicon(bookmarkID string, user security.User) (*ObjectInfo, error) {
	if bookmarkID == "" {
		return nil, app.ErrValidation("missing bookmark id parameter")
	}

	s.Logger.Info(fmt.Sprintf("try to fetch bookmark with ID '%s'", bookmarkID))
	existing, err := s.BookmarkStore.GetBookmarkByID(bookmarkID, user.Username)
	if err != nil {
		s.Logger.Error(fmt.Sprintf("could not find bookmark by id '%s': %v", bookmarkID, err))
		return nil, app.ErrNotFound(fmt.Sprintf("could not find bookmark with ID '%s'", bookmarkID))
	}

	defaultFavicon := app.DefaultFavicon
	switch existing.Type {
	case store.Node:
		defaultFavicon = app.DefaultFavicon
	case store.Folder:
		defaultFavicon = app.DefaultIconFolder
	case store.FileItem:
		defaultFavicon = app.DefaultIconFile
	}

	fi := ObjectInfo{
		Payload:  defaultFavicon,
		Name:     defaultFaviconName,
		Modified: defaultFaviconModTime,
	}

	if existing.Favicon == "" {
		return &fi, nil
	}

	favicon, err := s.FavStore.Get(existing.Favicon)
	if err != nil {
		s.Logger.Error(fmt.Sprintf("a favicon was defined '%s' but is not available", existing.Favicon))
		return &fi, nil
	}

	fi.Name = favicon.ID
	fi.Modified = favicon.LastModified
	fi.Payload = favicon.Payload

	return &fi, nil
}

// GetAvailableFavicons retrieves all available favicons of the stored bookmarks.
// It does not return duplicates, but focuses on unique favicons by comparing the stored payload.
// A optional search-term is provided to filter bookmarks by name and as a result the given favicons.
func (s *Application) GetAvailableFavicons(user security.User, search string) ([]ObjectInfo, error) {
	var (
		bookmarks []store.Bookmark
		err       error
	)
	if search != "" {
		bookmarks, err = s.BookmarkStore.GetBookmarksByName(search, user.Username)
	} else {
		bookmarks, err = s.BookmarkStore.GetAllBookmarks(user.Username)

	}
	if err != nil {
		return nil, fmt.Errorf("could not retrieve bookmarks for user '%s'; %v", user.Username, err)
	}

	// iterate of the full list and determine the faviconIDs which are unique
	uniqueFavicons := make(map[string]int16)
	for _, b := range bookmarks {
		if b.Favicon == "" {
			continue
		}
		if _, avail := uniqueFavicons[b.Favicon]; !avail {
			uniqueFavicons[b.Favicon] = 1
		}
	}

	// determine via a hash-function if the payload is unique as well
	uniquePayload := make(map[string]store.Favicon)
	for faviconID := range uniqueFavicons {
		favicon, err := s.FavStore.Get(faviconID)
		if err != nil {
			continue
		}
		if len(favicon.Payload) == 0 {
			continue
		}

		hashVal := hash(favicon.Payload)
		if _, avail := uniqueFavicons[hashVal]; !avail {
			uniquePayload[hashVal] = favicon
		}
	}

	// finally we have the complete list of unique favicons
	favicons := make([]ObjectInfo, len(uniquePayload))
	i := 0
	for key := range uniquePayload {
		favicon := uniquePayload[key]
		favicons[i] = ObjectInfo{
			// do not copy the payload around
			//Payload:  favicon.Payload,
			Name:     favicon.ID,
			Modified: favicon.LastModified,
		}
		i += 1
	}

	// be consistent in sorting
	sort.Slice(favicons, func(i, j int) bool {
		a := favicons[i]
		b := favicons[j]
		return a.Name < b.Name
	})

	return favicons, nil
}

// GetFaviconByID retrieves the favicon payload from the store
func (s *Application) GetFaviconByID(faviconID string) (*ObjectInfo, error) {
	if faviconID == "" {
		return nil, app.ErrValidation("missing favicon id parameter")
	}

	s.Logger.Info(fmt.Sprintf("try to fetch favicon with ID '%s'", faviconID))
	favicon, err := s.FavStore.Get(faviconID)
	if err != nil {
		s.Logger.Error(fmt.Sprintf("could not find favicon by id '%s': %v", faviconID, err))
		return nil, app.ErrNotFound(fmt.Sprintf("could not find favicon with ID '%s'", faviconID))
	}

	fi := ObjectInfo{
		Payload:  favicon.Payload,
		Name:     favicon.ID,
		Modified: favicon.LastModified,
	}

	return &fi, nil
}

// local favicon handling // store the retrieved favicon on disk
// --------------------------------------------------------------------------

// GetLocalFaviconByID reads to favicon payload from the local store
func (s *Application) GetLocalFaviconByID(id string) (*ObjectInfo, error) {
	if id == "" {
		return nil, app.ErrValidation("missing id parameter")
	}

	fullPath := path.Join(s.FaviconPath, id)
	meta, err := os.Stat(fullPath)
	if err != nil {
		s.Logger.Error(fmt.Sprintf("the specified favicon '%s' is not available", fullPath))
		// nevertheless we return a favicon, the standard one
		return &ObjectInfo{
			Payload:  app.DefaultFavicon,
			Name:     defaultFaviconName,
			Modified: defaultFaviconModTime,
		}, nil
	}
	payload, err := os.ReadFile(fullPath)
	if err != nil {
		s.Logger.Error(fmt.Sprintf("could not read payload for path '%s'", fullPath))
		return nil, fmt.Errorf("could not read the favicon for id '%s'", id)
	}
	if len(payload) == 0 {
		s.Logger.Error(fmt.Sprintf("got a Zero payload for path '%s'", fullPath))
		return nil, fmt.Errorf("could not read the favicon for id '%s'", id)
	}

	fi := ObjectInfo{
		Name:     meta.Name(),
		Payload:  payload,
		Modified: meta.ModTime(),
	}

	return &fi, nil
}

// RemoveLocalFavicon is used to delete the locally stored favicon payload if it is no longer needed
func (s *Application) RemoveLocalFavicon(id string) error {
	if id == "" {
		return app.ErrValidation("missing id parameter")
	}

	fullPath := path.Join(s.FaviconPath, id)
	if _, err := os.Stat(fullPath); err != nil {
		return fmt.Errorf("no favicon available for id '%s'", id)
	}
	return os.Remove(fullPath)
}

// LocalFetchFaviconURL retrieves the favicon from the given URL and stores the payload locally
func (s *Application) LocalFetchFaviconURL(url string) (*ObjectInfo, error) {
	if url == "" {
		return nil, app.ErrValidation("empty URL provided")
	}

	content, err := favicon.FetchURL(url, favicon.FetchImage)
	if err != nil {
		return nil, fmt.Errorf("cannot fetch favicon from URL '%s': %v", url, err)
	}
	if len(content.Payload) == 0 {
		return nil, fmt.Errorf("not payload for favicon from URL '%s'", url)
	}

	s.Logger.Info(fmt.Sprintf("got favicon payload length of '%d' for URL '%s'", len(content.Payload), url))
	return s.WriteLocalFavicon(content.FileName, content.MimeType, content.Payload)
}

// LocalExtractFaviconFromURL retrieves the Favicon from the bookmarks URL and stores the payload locally
func (s *Application) LocalExtractFaviconFromURL(baseURL string) (*ObjectInfo, error) {
	if baseURL == "" {
		return nil, app.ErrValidation("empty URL provided")
	}

	content, err := favicon.GetFaviconFromURL(baseURL)
	if err != nil {
		return nil, fmt.Errorf("cannot fetch favicon from URL '%s': %v", baseURL, err)
	}
	if len(content.Payload) == 0 {
		return nil, fmt.Errorf("no payload for favicon from URL '%s'", baseURL)
	}

	s.Logger.Info(fmt.Sprintf("got favicon payload length of '%d' for URL '%s'", len(content.Payload), baseURL))
	return s.WriteLocalFavicon(content.FileName, content.MimeType, content.Payload)
}

const faviconSizeX = 50

// WriteLocalFavicon takes a payload and stores it locally using a generated hash-code
func (s *Application) WriteLocalFavicon(name, mimeType string, payload []byte) (*ObjectInfo, error) {
	content := favicon.Content{
		FileName: name,
		MimeType: mimeType,
		Payload:  payload,
	}
	resized := s.resize(content, faviconSizeX, 0 /* keep aspect ratio */)

	hashPayload, err := hashInput(resized.Payload)
	if err != nil {
		return nil, fmt.Errorf("could not hash payload: '%v'", err)
	}
	hashFilename, err := hashInput([]byte(name))
	if err != nil {
		return nil, fmt.Errorf("could not hash filename: '%v'", err)
	}
	ext := filepath.Ext(name)
	filename := fmt.Sprintf("%s_%s%s", hashFilename, hashPayload, ext)
	fullPath := path.Join(s.FaviconPath, filename)

	if err := os.WriteFile(fullPath, resized.Payload, 0644); err != nil {
		return nil, fmt.Errorf("could not write favicon to file '%s': %v", fullPath, err)
	}

	fi := ObjectInfo{
		Name:     filename,
		Payload:  resized.Payload,
		Modified: time.Now(),
	}
	return &fi, nil
}

// ---- Internals ----

func (s *Application) saveFavicon(providedID string) (string, error) {
	faviconID := providedID

	// an existing favicon, which does not need to be processed is indicated with the given Prefix
	// if this is found, remote the prefix and return to caller
	if strings.HasPrefix(faviconID, ExistingFavicon) {
		return strings.ReplaceAll(faviconID, ExistingFavicon, ""), nil
	}

	obj, err := s.GetLocalFaviconByID(providedID)
	if err != nil {
		s.Logger.Error(fmt.Sprintf("wanted to access the provided favicon '%s' but got an error; %v", providedID, err))
		return "", fmt.Errorf("could not access the specified favicon")
	}
	// got a payload which should be persisted in the store
	err = s.FavStore.InUnitOfWork(func(favRepo store.FaviconRepository) error {
		if _, e := favRepo.Save(store.Favicon{
			ID:           obj.Name,
			Payload:      obj.Payload,
			LastModified: obj.Modified,
		}); e != nil {
			return e
		}
		// the favicon was saved in the store - we can delete the local one
		if e := s.RemoveLocalFavicon(obj.Name); e != nil {
			s.Logger.Error(fmt.Sprintf("could not remove the local favicon; %v", err))
			// we do not fail here, not worth the fail the overall operation if only the
			// removal of the local file is not possible
		}
		return nil
	})
	if err != nil {
		s.Logger.Error(fmt.Sprintf("got an error while trying to store the favicon; %v", err))
		return "", fmt.Errorf("could not store the specified favicon; %v", err)
	}
	return faviconID, nil
}

func (s *Application) processFile(bookmarkFileID string) (string, error) {
	var (
		fileID string
	)

	if bookmarkFileID == "" {
		return fileID, nil
	}

	// the file was upload before and we have a reference of the id
	// we validate the existence of the ID and store the file.
	upload, err := s.UploadSvc.Read(bookmarkFileID)
	if err != nil {
		s.Logger.Error(fmt.Sprintf("the given file is not available %v", err), logging.LogV("FileID", bookmarkFileID))
		return "", fmt.Errorf("could not fetch the file identified by '%s'", bookmarkFileID)
	}
	err = s.FileStore.InUnitOfWork(func(repo store.FileRepository) error {
		file, e := repo.Save(store.File{
			Name:     upload.FileName,
			MimeType: upload.MimeType,
			FileObject: &store.FileObject{
				Payload: upload.Payload,
			},
			Size: len(upload.Payload),
		})
		if e != nil {
			return e
		}
		fileID = file.ID
		s.Logger.Info("saved a new file", logging.LogV("ID", fileID))

		e = s.UploadSvc.Delete(bookmarkFileID)
		if e != nil {
			// fail silently: it is not important for the overall logic, that
			// the uploaded file is deleted.
			s.Logger.Error(fmt.Sprintf("could not cleanup uploaded file; %v", err))
		}
		return nil
	})
	if err != nil {
		s.Logger.Error(fmt.Sprintf("got an error while trying to store the file; %v", err))
		return fileID, fmt.Errorf("could not store the specified file; %v", err)
	}
	return fileID, nil

}

func (s *Application) resize(content favicon.Content, x, y int) favicon.Content {
	resized, err := favicon.ResizeImage(content, x, y)
	if err != nil {
		s.Logger.Error("could not resize favicon", logging.ErrV(err))
		return content
	}
	return resized
}

func hashInput(input []byte) (string, error) {
	hash := sha1.New()
	if _, err := hash.Write(input); err != nil {
		return "", fmt.Errorf("could not hash input: %v", err)
	}
	bs := hash.Sum(nil)
	return fmt.Sprintf("%x", bs), nil
}

// EnsureFolderPath takes care that the supplied path is valid
// e.g. it does not start with two slashes '//' and that the resulting
// path is valid, with all necessary delimiters
func ensureFolderPath(path, displayName string) string {
	folderPath := path
	if !strings.HasSuffix(path, "/") {
		folderPath = path + "/"
	}
	if strings.HasPrefix(folderPath, "//") {
		folderPath = strings.ReplaceAll(folderPath, "//", "/")
	}
	folderPath += displayName
	return folderPath
}

func hash(payload []byte) string {
	// Create a new SHA-1 hash object
	hasher := sha1.New()
	hasher.Write(payload)
	checksum := hasher.Sum(nil)
	hexChecksum := hex.EncodeToString(checksum)
	return hexChecksum
}
