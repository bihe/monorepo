package bookmarks

import (
	"crypto/sha1"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"golang.binggl.net/monorepo/internal/bookmarks-new/app/favicon"
	"golang.binggl.net/monorepo/internal/bookmarks-new/app/store"
	"golang.binggl.net/monorepo/pkg/logging"
	"golang.binggl.net/monorepo/pkg/security"
)

// Service implements the bookmarks logic
type Service struct {
	// Logger instance to use for structured logging
	Logger logging.Logger
	// Store defines a repository for persistance
	Store store.Repository
	// FaviconPath defines a path on disk to store a favicon
	FaviconPath string
	// DefaultFavicon is used if the bookmark item does not have a favicon
	DefaultFavicon string
}

// CreateBookmark stores a new bookmark
func (s *Service) CreateBookmark(bm Bookmark, user security.User) (*Bookmark, error) {
	var (
		savedItem store.Bookmark
		t         store.NodeType
	)

	if bm.Path == "" || bm.DisplayName == "" {
		s.Logger.Error("required fields of bookmarks missing (Path or DisplayName)")
		return nil, fmt.Errorf("invalid request data supplied, missing Path or DisplayName")
	}

	if bm.Type == Folder {
		t = store.Folder
	}

	if err := s.Store.InUnitOfWork(func(repo store.Repository) error {
		item, err := repo.Create(store.Bookmark{
			DisplayName: bm.DisplayName,
			Path:        bm.Path,
			Type:        t,
			URL:         bm.URL,
			UserName:    user.Username,
			Favicon:     bm.Favicon,
			SortOrder:   bm.SortOrder,
			Highlight:   bm.Highlight,
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

	if bm.CustomFavicon != "" {
		// fire&forget, run this in background and do not wait for the result
		go s.FetchFaviconURL(bm.CustomFavicon, savedItem, user)
	} else {
		// if no specific favicon was supplied, immediately fetch one
		if savedItem.Favicon == "" && savedItem.Type == store.Node {
			// fire&forget, run this in background and do not wait for the result
			go s.FetchFavicon(savedItem, user)
		}
	}

	id := savedItem.ID
	s.Logger.Info(fmt.Sprintf("bookmark created with ID: %s", id))
	return entityToModel(savedItem), nil
}

// GetBookmarkByID retrieves a bookmark for the given user
func (s *Service) GetBookmarkByID(id string, user security.User) (*Bookmark, error) {
	if id == "" {
		return nil, fmt.Errorf("no id supplied to fetch bookmark")
	}

	bm, err := s.Store.GetBookmarkByID(id, user.Username)
	if err != nil {
		return nil, fmt.Errorf("could not fetch bookmark; %v", err)
	}
	return entityToModel(bm), nil
}

// GetBookmarksByPath retrieves bookmarks by a given path
func (s *Service) GetBookmarksByPath(path string, user security.User) ([]Bookmark, error) {
	if path == "" {
		return nil, fmt.Errorf("missing path")
	}

	s.Logger.Info(fmt.Sprintf("get bookmarks by path: '%s' for user: '%s'", path, user.Username))

	bms, err := s.Store.GetBookmarksByPath(path, user.Username)
	if err != nil {
		s.Logger.Error(fmt.Sprintf("cannot get bookmark by path: '%s', %v", path, err))
	}
	return entityListToModel(bms), nil
}

// GetBookmarksFolderByPath returns the folder identified by the given path
func (s *Service) GetBookmarksFolderByPath(path string, user security.User) (*Bookmark, error) {
	if path == "" {
		return nil, fmt.Errorf("missing path parameter")
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
	bm, err := s.Store.GetFolderByPath(path, user.Username)
	if err != nil {
		s.Logger.Error(fmt.Sprintf("cannot get bookmark folder by path: '%s', %v", path, err))
		return nil, fmt.Errorf("no folder for path '%s' found", path)
	}

	return entityToModel(bm), nil
}

// GetAllPaths returns all available stored paths
func (s *Service) GetAllPaths(user security.User) ([]string, error) {
	paths, err := s.Store.GetAllPaths(user.Username)
	if err != nil {
		s.Logger.Error(fmt.Sprintf("cannot get all bookmark paths: %v", err))
		return make([]string, 0), fmt.Errorf("cannot get all bookmark paths")
	}
	return paths, nil
}

// GetBookmarksByName returns the given bookmark(s) by the supplied name
func (s *Service) GetBookmarksByName(name string, user security.User) ([]Bookmark, error) {
	if name == "" {
		return make([]Bookmark, 0), fmt.Errorf("missing name parameter")
	}

	s.Logger.Info(fmt.Sprintf("get bookmarks by name: '%s' for user: '%s'", name, user.Username))
	bms, err := s.Store.GetBookmarksByName(name, user.Username)
	var bookmarks []Bookmark
	if err != nil {
		s.Logger.Error(fmt.Sprintf("cannot get bookmark by name: '%s', %v", name, err))
	}
	bookmarks = entityListToModel(bms)
	return bookmarks, nil
}

// FetchAndForward retrieves a bookmark by id and forwards to the url of the bookmark
func (s *Service) FetchAndForward(id string, user security.User) (string, error) {
	if id == "" {
		return "", fmt.Errorf("missing id parameter")
	}

	s.Logger.Info(fmt.Sprintf("try to fetch bookmark with ID '%s'", id))
	redirectURL := ""
	if err := s.Store.InUnitOfWork(func(repo store.Repository) error {
		existing, err := repo.GetBookmarkByID(id, user.Username)
		if err != nil {
			s.Logger.Error(fmt.Sprintf("could not find bookmark by id '%s': %v", id, err))
			return err
		}

		if existing.Type == store.Folder {
			s.Logger.Error(fmt.Sprintf("accessCount and redirect only valid for Nodes: ID '%s'", id))
			return fmt.Errorf("cannot fetch and forward folder - ID '%s'", id)
		}

		// once accessed the highlight flag is removed
		existing.Highlight = 0
		if _, err := repo.Update(existing); err != nil {
			s.Logger.Error(fmt.Sprintf("could not update bookmark '%s': %v", id, err))
			return err
		}
		redirectURL = existing.URL

		if existing.Favicon == "" {
			// fire&forget, run this in background and do not wait for the result
			go s.FetchFavicon(existing, user)
		}
		return nil
	}); err != nil {
		s.Logger.Error(fmt.Sprintf("could not fetch and update bookmark by ID '%s': %v", id, err))
		return "", fmt.Errorf("error fetching and updating bookmark: %v", err)
	}

	s.Logger.Info(fmt.Sprintf("will redirect to bookmark URL '%s'", redirectURL))
	return redirectURL, nil
}

// Delete a bookmark by id
func (s *Service) Delete(id string, user security.User) error {
	if id == "" {
		return fmt.Errorf("missing id parameter")
	}

	s.Logger.Info(fmt.Sprintf("will try to delete bookmark with ID '%s'", id))
	if err := s.Store.InUnitOfWork(func(repo store.Repository) error {
		// 1) fetch the existing bookmark by id
		existing, err := repo.GetBookmarkByID(id, user.Username)
		if err != nil {
			s.Logger.Error(fmt.Sprintf("could not find bookmark by id '%s': %v", id, err))
			return err
		}

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

	return nil
}

// UpdateSortOrder modifies the display sort-order
func (s *Service) UpdateSortOrder(sort BookmarksSortOrder, user security.User) (int, error) {
	if len(sort.IDs) != len(sort.SortOrder) {
		return 0, fmt.Errorf("the number of IDs (%d) does not correspond the number of SortOrder entries (%d)", len(sort.IDs), len(sort.SortOrder))
	}

	var updates int
	if err := s.Store.InUnitOfWork(func(repo store.Repository) error {
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
		s.Logger.Error(fmt.Sprintf("could not update the sortorder for bookmark: %v", err))
		return 0, fmt.Errorf("error updating sortorder of bookmark: %v", err)
	}
	return updates, nil
}

// ---- Favicon Logic ----

// GetFaviconPath for the specified bookmark, return the path to the found favicon or the default favicon
func (s *Service) GetFaviconPath(id string, user security.User) (string, error) {
	if id == "" {
		return "", fmt.Errorf("missing id parameter")
	}

	s.Logger.Info(fmt.Sprintf("try to fetch bookmark with ID '%s'", id))
	existing, err := s.Store.GetBookmarkByID(id, user.Username)
	if err != nil {
		s.Logger.Error(fmt.Sprintf("could not find bookmark by id '%s': %v", id, err))
		return "", fmt.Errorf("could not find bookmark with ID '%s'", id)
	}

	if existing.Favicon == "" {
		return path.Join(s.FaviconPath, s.DefaultFavicon), nil
	}

	fullPath := path.Join(s.FaviconPath, existing.Favicon)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		s.Logger.Error(fmt.Sprintf("the specified favicon '%s' is not available", fullPath))
		// not found - use default
		fullPath = path.Join(s.FaviconPath, s.DefaultFavicon)
	}
	return fullPath, nil
}

// FetchFaviconURL retrieves the favicon from the given URL
func (s *Service) FetchFaviconURL(url string, bm store.Bookmark, user security.User) {
	payload, err := favicon.FetchURL(url)
	if err != nil {
		s.Logger.Error(fmt.Sprintf("cannot fetch favicon from URL '%s': %v", url, err))
		return
	}
	favi := ""
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		favi = parts[len(parts)-1] // last element
	} else {
		s.Logger.Error(fmt.Sprintf("cannot get favicon-filename from URL '%s': %v", url, err))
		return
	}

	if len(payload) == 0 {
		s.Logger.Error(fmt.Sprintf("not payload for favicon from URL '%s'", url))
		return
	}

	s.Logger.Info(fmt.Sprintf("got favicon payload length of '%d' for URL '%s'", len(payload), url))
	hashPayload, err := hashInput(payload)
	if err != nil {
		s.Logger.Error(fmt.Sprintf("could not hash payload: '%v'", err))
		return
	}
	hashFilename, err := hashInput([]byte(favi))
	if err != nil {
		s.Logger.Error(fmt.Sprintf("could not hash filename: '%v'", err))
		return
	}
	ext := filepath.Ext(favi)
	filename := fmt.Sprintf("%s_%s%s", hashFilename, hashPayload, ext)
	fullPath := path.Join(s.FaviconPath, filename)

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		if err := os.WriteFile(fullPath, payload, 0644); err != nil {
			s.Logger.Error(fmt.Sprintf("could not write favicon to file '%s': %v", fullPath, err))
			return
		}
	}

	// also update the favi for the bookmark
	if err := s.Store.InUnitOfWork(func(repo store.Repository) error {
		bm.Favicon = filename
		_, err := repo.Update(bm)
		return err
	}); err != nil {
		s.Logger.Error(fmt.Sprintf("could not update bookmark with favicon '%s': %v", filename, err))
	}
}

// FetchFavicon retrieves the Favicon from the bookmarks URL
func (s *Service) FetchFavicon(bm store.Bookmark, user security.User) {
	favi, payload, err := favicon.GetFaviconFromURL(bm.URL)
	if err != nil {
		s.Logger.Error(fmt.Sprintf("cannot fetch favicon from URL '%s': %v", bm.URL, err))
		return
	}

	if len(payload) == 0 {
		s.Logger.Error(fmt.Sprintf("no payload for favicon from URL '%s'", bm.URL))
		return
	}

	s.Logger.Info(fmt.Sprintf("got favicon payload length of '%d' for URL '%s'", len(payload), bm.URL))
	hashPayload, err := hashInput(payload)
	if err != nil {
		s.Logger.Error(fmt.Sprintf("could not hash payload: '%v'", err))
		return
	}
	hashFilename, err := hashInput([]byte(favi))
	if err != nil {
		s.Logger.Error(fmt.Sprintf("could not hash filename: '%v'", err))
		return
	}
	ext := filepath.Ext(favi)
	filename := fmt.Sprintf("%s_%s%s", hashFilename, hashPayload, ext)
	fullPath := path.Join(s.FaviconPath, filename)

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		if err := os.WriteFile(fullPath, payload, 0644); err != nil {
			s.Logger.Error(fmt.Sprintf("could not write favicon to file '%s': %v", fullPath, err))
			return
		}
	}

	// also update the favi for the bookmark
	if err := s.Store.InUnitOfWork(func(repo store.Repository) error {
		bm.Favicon = filename
		_, err := repo.Update(bm)
		return err
	}); err != nil {
		s.Logger.Error(fmt.Sprintf("could not update bookmark with favicon '%s': %v", filename, err))
	}

}

// ---- Internals ----

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

// // Update a bookmark
// // swagger:operation PUT /api/v1/bookmarks bookmarks UpdateBookmark
// //
// // update a bookmark
// //
// // use the supplied payload to update a existing bookmark
// //
// // ---
// // consumes:
// // - application/json
// // produces:
// // - application/json
// // responses:
// //   '200':
// //     description: Result
// //     schema:
// //       "$ref": "#/definitions/Result"
// //   '400':
// //     description: ProblemDetail
// //     schema:
// //       "$ref": "#/definitions/ProblemDetail"
// //   '401':
// //     description: ProblemDetail
// //     schema:
// //       "$ref": "#/definitions/ProblemDetail"
// //   '403':
// //     description: ProblemDetail
// //     schema:
// //       "$ref": "#/definitions/ProblemDetail"
// func (b *BookmarksAPI) Update(user security.User, w http.ResponseWriter, r *http.Request) error {
// 	var (
// 		id      string
// 		payload *BookmarkRequest
// 	)

// 	payload = &BookmarkRequest{}
// 	if err := render.Bind(r, payload); err != nil {
// 		b.Log.ErrorRequest(fmt.Sprintf("cannot bind payload: '%v'", err), r)
// 		return errors.BadRequestError{Err: fmt.Errorf("invalid request data supplied"), Request: r}
// 	}

// 	if payload.Path == "" || payload.DisplayName == "" || payload.ID == "" {
// 		b.Log.ErrorRequest("required fields of bookmarks missing", r)
// 		return errors.BadRequestError{Err: fmt.Errorf("invalid request data supplied, missing ID, Path or DisplayName"), Request: r}

// 	}

// 	b.Log.InfoRequest(fmt.Sprintf("will try to update existing bookmark entry: '%s'", payload), r)
// 	if err := b.Repository.InUnitOfWork(func(repo store.Repository) error {
// 		// 1) fetch the existing bookmark by id
// 		existing, err := repo.GetBookmarkByID(payload.ID, user.Username)
// 		if err != nil {
// 			b.Log.ErrorRequest(fmt.Sprintf("could not find bookmark by id '%s': %v", payload.ID, err), r)
// 			return err
// 		}
// 		childCount := existing.ChildCount
// 		if existing.Type == store.Folder {
// 			// 2) ensure that the existing folder is not moved to itself
// 			folderPath := ensureFolderPath(existing.Path, existing.DisplayName)
// 			if payload.Path == folderPath {
// 				b.Log.ErrorRequest(fmt.Sprintf("a folder cannot be moved into itself: folder-path: '%s', destination: '%s'", folderPath, payload.Path), r)
// 				return errors.BadRequestError{Err: fmt.Errorf("cannot move folder into itself"), Request: r}
// 			}

// 			// 3) get the folder child-count
// 			// on save of a folder, update the child-count
// 			parentPath := existing.Path
// 			path := ensureFolderPath(parentPath, existing.DisplayName)
// 			nodeCount, err := repo.GetPathChildCount(path, user.Username)
// 			if err != nil {
// 				b.Log.ErrorRequest(fmt.Sprintf("could not get child-count of path '%s': %v", path, err), r)
// 				return err
// 			}
// 			if len(nodeCount) > 0 {
// 				for _, n := range nodeCount {
// 					if n.Path == path {
// 						childCount = n.Count
// 						break
// 					}
// 				}
// 			}
// 		}

// 		existingDisplayName := existing.DisplayName
// 		existingPath := existing.Path

// 		// 4) update the bookmark
// 		item, err := repo.Update(store.Bookmark{
// 			ID:          payload.ID,
// 			Created:     existing.Created,
// 			DisplayName: payload.DisplayName,
// 			Path:        payload.Path,
// 			Type:        existing.Type, // it does not make any sense to change the type of a bookmark!
// 			URL:         payload.URL,
// 			SortOrder:   payload.SortOrder,
// 			UserName:    user.Username,
// 			ChildCount:  childCount,
// 			Favicon:     payload.Favicon,
// 			Highlight:   payload.Highlight,
// 		})
// 		if err != nil {
// 			b.Log.ErrorRequest(fmt.Sprintf("could not update bookmark: %v", err), r)
// 			return err
// 		}
// 		id = item.ID

// 		// also update the favicon if not available
// 		if item.Favicon == "" {
// 			// fire&forget, run this in background and do not wait for the result
// 			go b.FetchFavicon(item, user)
// 		}
// 		if payload.CustomFavicon != "" {
// 			// fire&forget, run this in background and do not wait for the result
// 			go b.FetchFaviconURL(payload.CustomFavicon, item, user)
// 		}

// 		if existing.Type == store.Folder && (existingDisplayName != payload.DisplayName || existingPath != payload.Path) {
// 			// if we have a folder and change the displayname or the parent-path, this also affects ALL sub-elements
// 			// therefore all paths of sub-elements where this folder-path is present, need to be updated
// 			newPath := ensureFolderPath(payload.Path, payload.DisplayName)
// 			oldPath := ensureFolderPath(existingPath, existingDisplayName)

// 			b.Log.InfoRequest(fmt.Sprintf("will update all old paths '%s' to new path '%s'", oldPath, newPath), r)
// 			bookmarks, err := repo.GetBookmarksByPathStart(oldPath, user.Username)
// 			if err != nil {
// 				b.Log.ErrorRequest(fmt.Sprintf("could not get bookmarks by path '%s': %v", oldPath, err), r)
// 				return err
// 			}

// 			for _, bm := range bookmarks {
// 				updatePath := strings.ReplaceAll(bm.Path, oldPath, newPath)
// 				if _, err := repo.Update(store.Bookmark{
// 					ID:          bm.ID,
// 					Created:     bm.Created,
// 					DisplayName: bm.DisplayName,
// 					Path:        updatePath,
// 					SortOrder:   bm.SortOrder,
// 					Type:        bm.Type,
// 					URL:         bm.URL,
// 					UserName:    user.Username,
// 					ChildCount:  bm.ChildCount,
// 					Highlight:   bm.Highlight,
// 					Favicon:     bm.Favicon,
// 				}); err != nil {
// 					b.Log.ErrorRequest(fmt.Sprintf("cannot update bookmark path: %v", err), r)
// 					return err
// 				}
// 			}
// 		}

// 		// if the path has changed - update the childcount of affected paths
// 		if existingPath != payload.Path {
// 			// the affected paths are the origin-path and the destination-path
// 			if err := b.updateChildCountOfPath(r, existingPath, user.Username, repo); err != nil {
// 				b.Log.ErrorRequest(fmt.Sprintf("could not update child-count of origin-path '%s': %v", existingPath, err), r)
// 				return err
// 			}
// 			// and the destination-path
// 			if err := b.updateChildCountOfPath(r, payload.Path, user.Username, repo); err != nil {
// 				b.Log.ErrorRequest(fmt.Sprintf("could not update child-count of destination-path '%s': %v", payload.Path, err), r)
// 				return err
// 			}
// 		}

// 		return nil
// 	}); err != nil {
// 		b.Log.ErrorRequest(fmt.Sprintf("could not update bookmark because of error: %v", err), r)

// 		var badRequest errors.BadRequestError
// 		if er.As(err, &badRequest) {
// 			return badRequest
// 		}
// 		return errors.ServerError{Err: fmt.Errorf("error updating bookmark: %v", err), Request: r}
// 	}

// 	b.Log.InfoRequest(fmt.Sprintf("updated bookmark with ID '%s'", id), r)
// 	return render.Render(w, r, ResultResponse{
// 		Result: &Result{
// 			Success: true,
// 			Message: fmt.Sprintf("Bookmark with ID '%s' was updated", id),
// 			Value:   id,
// 		},
// 	})
// }

// func (b *BookmarksAPI) updateChildCountOfPath(r *http.Request, path, username string, repo store.Repository) error {
// 	if path == "/" {
// 		b.Log.InfoRequest("skip the ROOT path '/'", r)
// 		return nil
// 	}

// 	folder, err := repo.GetFolderByPath(path, username)
// 	if err != nil {
// 		b.Log.ErrorRequest(fmt.Sprintf("cannot get the folder for path '%s': %v", path, err), r)
// 		return err
// 	}

// 	nodeCount, err := repo.GetPathChildCount(path, username)
// 	if err != nil {
// 		b.Log.ErrorRequest(fmt.Sprintf("could not get the child-count of the path '%s': %v", path, err), r)
// 		return err
// 	}

// 	var childCount int
// 	if len(nodeCount) > 0 {
// 		// use the firest element, because the count was queried for a specific path
// 		childCount = nodeCount[0].Count
// 	}

// 	folder.ChildCount = childCount

// 	if _, err := repo.Update(folder); err != nil {
// 		b.Log.ErrorRequest(fmt.Sprintf("could not update the child-count of folder '%s': %v", path, err), r)
// 		return err
// 	}
// 	return nil
// }
