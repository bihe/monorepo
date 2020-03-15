// Package api implements the HTTP API of the bookmarks-application.
//
// the purpose of this application is to provide bookmarks management independent of browsers
//
// Terms Of Service:
//
//     Schemes: https
//     Host: bookmarks.binggl.net
//     BasePath: /api/v1
//     Version: 1.0.0
//     License: Apache 2.0 https://opensource.org/licenses/Apache-2.0
//
//     Consumes:
//     - application/json
//
//     Produces:
//     - application/json
//
// swagger:meta
package api

import (
	"crypto/sha1"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	er "errors"

	"github.com/bihe/bookmarks/internal/favicon"
	"github.com/bihe/bookmarks/internal/store"
	"golang.binggl.net/commons/errors"
	"golang.binggl.net/commons/handler"
	"golang.binggl.net/commons/security"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

// BookmarksAPI implement the handler for the API
type BookmarksAPI struct {
	handler.Handler
	Repository     store.Repository
	BasePath       string
	FaviconPath    string
	DefaultFavicon string
}

// swagger:operation GET /api/v1/bookmarks/{id} bookmarks GetBookmarkByID
//
// get a bookmark by id
//
// returns a single bookmark specified by it's ID
//
// ---
// produces:
// - application/json
// parameters:
// - name: id
//   in: path
// responses:
//   '200':
//     description: Bookmark
//     schema:
//       "$ref": "#/definitions/Bookmark"
//   '400':
//     description: ProblemDetail
//     schema:
//       "$ref": "#/definitions/ProblemDetail"
//   '404':
//     description: ProblemDetail
//     schema:
//       "$ref": "#/definitions/ProblemDetail"
//   '401':
//     description: ProblemDetail
//     schema:
//       "$ref": "#/definitions/ProblemDetail"
//   '403':
//     description: ProblemDetail
//     schema:
//       "$ref": "#/definitions/ProblemDetail"
func (b *BookmarksAPI) GetBookmarkByID(user security.User, w http.ResponseWriter, r *http.Request) error {
	id := chi.URLParam(r, "id")

	if id == "" {
		return errors.BadRequestError{Err: fmt.Errorf("missing id parameter"), Request: r}
	}

	handler.LogFunction("api.GetBookmarkByID").Debugf("try to get bookmark by ID: '%s' for user: '%s'", id, user.Username)

	bookmark, err := b.Repository.GetBookmarkById(id, user.Username)
	if err != nil {
		handler.LogFunction("api.GetBookmarkByID").Warnf("cannot get bookmark by ID: '%s', %v", id, err)
		return errors.NotFoundError{Err: fmt.Errorf("no bookmark with ID '%s' avaliable", id), Request: r}
	}

	return render.Render(w, r, BookmarkResponse{Bookmark: entityToModel(bookmark)})
}

// swagger:operation GET /api/v1/bookmarks/bypath bookmarks GetBookmarksByPath
//
// get bookmarks by path
//
// returns a list of bookmarks for a given path
//
// ---
// produces:
// - application/json
// parameters:
// - name: path
//   in: query
// responses:
//   '200':
//     description: BookmarkList
//     schema:
//       "$ref": "#/definitions/BookmarkList"
//   '400':
//     description: ProblemDetail
//     schema:
//       "$ref": "#/definitions/ProblemDetail"
//   '401':
//     description: ProblemDetail
//     schema:
//       "$ref": "#/definitions/ProblemDetail"
//   '403':
//     description: ProblemDetail
//     schema:
//       "$ref": "#/definitions/ProblemDetail"
func (b *BookmarksAPI) GetBookmarksByPath(user security.User, w http.ResponseWriter, r *http.Request) error {
	path := r.URL.Query().Get("path")

	if path == "" {
		return errors.BadRequestError{Err: fmt.Errorf("missing path parameter"), Request: r}
	}

	handler.LogFunction("api.GetBookmarksByPath").Debugf("get bookmarks by path: '%s' for user: '%s'", path, user.Username)

	bms, err := b.Repository.GetBookmarksByPath(path, user.Username)
	var bookmarks []Bookmark
	if err != nil {
		handler.LogFunction("api.GetBookmarksByPath").Warnf("cannot get bookmark by path: '%s', %v", path, err)
	}
	bookmarks = entityListToModel(bms)
	count := len(bookmarks)
	result := BookmarkList{
		Success: true,
		Count:   count,
		Message: fmt.Sprintf("Found %d items.", count),
		Value:   bookmarks,
	}

	return render.Render(w, r, BookmarkListResponse{BookmarkList: &result})
}

// swagger:operation GET /api/v1/bookmarks/folder bookmarks GetBookmarksFolderByPath
//
// get bookmark folder by path
//
// returns the folder identified by the given path
//
// ---
// produces:
// - application/json
// parameters:
// - name: path
//   in: query
// responses:
//   '200':
//     description: BookmarkResult
//     schema:
//       "$ref": "#/definitions/BookmarkResult"
//   '400':
//     description: ProblemDetail
//     schema:
//       "$ref": "#/definitions/ProblemDetail"
//   '404':
//     description: ProblemDetail
//     schema:
//       "$ref": "#/definitions/ProblemDetail"
//   '401':
//     description: ProblemDetail
//     schema:
//       "$ref": "#/definitions/ProblemDetail"
//   '403':
//     description: ProblemDetail
//     schema:
//       "$ref": "#/definitions/ProblemDetail"
func (b *BookmarksAPI) GetBookmarksFolderByPath(user security.User, w http.ResponseWriter, r *http.Request) error {
	path := r.URL.Query().Get("path")

	if path == "" {
		return errors.BadRequestError{Err: fmt.Errorf("missing path parameter"), Request: r}
	}

	handler.LogFunction("api.GetBookmarksFolderByPath").Debugf("get bookmarks-folder by path: '%s' for user: '%s'", path, user.Username)

	if path == "/" {
		// special treatment for the root path. This path is ALWAYS available
		// and does not have a specific storage entry - this is by convention
		return render.Render(w, r, BookmarResultResponse{BookmarkResult: &BookmarkResult{
			Success: true,
			Message: fmt.Sprintf("Found bookmark folder for path %s.", path),
			Value: Bookmark{
				DisplayName: "Root",
				Path:        "/",
				Type:        Folder,
				ID:          fmt.Sprintf("%s_ROOT", user.Username),
			},
		}})
	}

	bm, err := b.Repository.GetFolderByPath(path, user.Username)
	if err != nil {
		handler.LogFunction("api.GetBookmarksFolderByPath").Warnf("cannot get bookmark folder by path: '%s', %v", path, err)
		return errors.NotFoundError{Err: fmt.Errorf("no folder for path '%s' found", path), Request: r}
	}

	return render.Render(w, r, BookmarResultResponse{BookmarkResult: &BookmarkResult{
		Success: true,
		Message: fmt.Sprintf("Found bookmark folder for path %s.", path),
		Value:   *entityToModel(bm),
	}})
}

// swagger:operation GET /api/v1/bookmarks/allpaths bookmarks GetAllPaths
//
// return all paths
//
// determine all available paths for the given user
//
// ---
// produces:
// - application/json
// responses:
//   '200':
//     description: BookmarksPathsResponse
//     schema:
//       "$ref": "#/definitions/BookmarksPathsResponse"
//   '400':
//     description: ProblemDetail
//     schema:
//       "$ref": "#/definitions/ProblemDetail"
//   '404':
//     description: ProblemDetail
//     schema:
//       "$ref": "#/definitions/ProblemDetail"
//   '401':
//     description: ProblemDetail
//     schema:
//       "$ref": "#/definitions/ProblemDetail"
//   '403':
//     description: ProblemDetail
//     schema:
//       "$ref": "#/definitions/ProblemDetail"
func (b *BookmarksAPI) GetAllPaths(user security.User, w http.ResponseWriter, r *http.Request) error {
	paths, err := b.Repository.GetAllPaths(user.Username)
	if err != nil {
		handler.LogFunction("api.GetAllPaths").Warnf("cannot get all bookmark paths: %v", err)
		return errors.ServerError{Err: fmt.Errorf("cannot get all bookmark paths"), Request: r}
	}

	return render.Render(w, r, BookmarksPathsResponse{
		Paths: paths,
		Count: len(paths),
	})
}

// swagger:operation GET /api/v1/bookmarks/byname bookmarks GetBookmarksByName
//
// get bookmarks by name
//
// search for bookmarks by name and return a list of search-results
//
// ---
// produces:
// - application/json
// parameters:
// - name: name
//   in: query
// responses:
//   '200':
//     description: BookmarkList
//     schema:
//       "$ref": "#/definitions/BookmarkList"
//   '400':
//     description: ProblemDetail
//     schema:
//       "$ref": "#/definitions/ProblemDetail"
//   '401':
//     description: ProblemDetail
//     schema:
//       "$ref": "#/definitions/ProblemDetail"
//   '403':
//     description: ProblemDetail
//     schema:
//       "$ref": "#/definitions/ProblemDetail"
func (b *BookmarksAPI) GetBookmarksByName(user security.User, w http.ResponseWriter, r *http.Request) error {
	name := r.URL.Query().Get("name")

	if name == "" {
		return errors.BadRequestError{Err: fmt.Errorf("missing name parameter"), Request: r}
	}

	handler.LogFunction("api.GetBookmarksByName").Debugf("get bookmarks by name: '%s' for user: '%s'", name, user.Username)

	bms, err := b.Repository.GetBookmarksByName(name, user.Username)
	var bookmarks []Bookmark
	if err != nil {
		handler.LogFunction("api.GetBookmarksByName").Warnf("cannot get bookmark by name: '%s', %v", name, err)
	}
	bookmarks = entityListToModel(bms)
	count := len(bookmarks)
	result := BookmarkList{
		Success: true,
		Count:   count,
		Message: fmt.Sprintf("Found %d items.", count),
		Value:   bookmarks,
	}

	return render.Render(w, r, BookmarkListResponse{BookmarkList: &result})
}

// swagger:operation GET /api/v1/bookmarks/mostvisited/{num} bookmarks GetMostVisited
//
// get recent accessed bookmarks
//
// return the most recently visited bookmarks
//
// ---
// produces:
// - application/json
// parameters:
// - name: num
//   in: path
// responses:
//   '200':
//     description: BookmarkList
//     schema:
//       "$ref": "#/definitions/BookmarkList"
//   '401':
//     description: ProblemDetail
//     schema:
//       "$ref": "#/definitions/ProblemDetail"
//   '403':
//     description: ProblemDetail
//     schema:
//       "$ref": "#/definitions/ProblemDetail"
func (b *BookmarksAPI) GetMostVisited(user security.User, w http.ResponseWriter, r *http.Request) error {
	n := chi.URLParam(r, "num")
	num, _ := strconv.Atoi(n)
	if num < 1 {
		num = 100
	}

	handler.LogFunction("api.GetMostVisited").Debugf("get the most recent, most often visited bookmarks for user: '%s'", user.Username)

	bms, err := b.Repository.GetMostRecentBookmarks(user.Username, num)
	var bookmarks []Bookmark
	if err != nil {
		handler.LogFunction("api.GetMostVisited").Warnf("cannot get most visited bookmarks: '%v'", err)
	}
	bookmarks = entityListToModel(bms)
	count := len(bookmarks)
	result := BookmarkList{
		Success: true,
		Count:   count,
		Message: fmt.Sprintf("Found %d items.", count),
		Value:   bookmarks,
	}

	return render.Render(w, r, BookmarkListResponse{BookmarkList: &result})
}

// swagger:operation POST /api/v1/bookmarks bookmarks CreateBookmark
//
// create a bookmark
//
// use the supplied payload to create a new bookmark
//
// ---
// consumes:
// - application/json
// produces:
// - application/json
// responses:
//   '200':
//     description: Result
//     schema:
//       "$ref": "#/definitions/Result"
//   '400':
//     description: ProblemDetail
//     schema:
//       "$ref": "#/definitions/ProblemDetail"
//   '401':
//     description: ProblemDetail
//     schema:
//       "$ref": "#/definitions/ProblemDetail"
//   '403':
//     description: ProblemDetail
//     schema:
//       "$ref": "#/definitions/ProblemDetail"
func (b *BookmarksAPI) Create(user security.User, w http.ResponseWriter, r *http.Request) error {
	var (
		id      string
		payload *BookmarkRequest
		t       store.NodeType
	)

	payload = &BookmarkRequest{}
	if err := render.Bind(r, payload); err != nil {
		handler.LogFunction("api.Create").Warnf("cannot bind payload: '%v'", err)
		return errors.BadRequestError{Err: fmt.Errorf("invalid request data supplied"), Request: r}
	}

	if payload.Path == "" || payload.DisplayName == "" {
		handler.LogFunction("api.Create").Warnf("required fields of bookmarks missing")
		return errors.BadRequestError{Err: fmt.Errorf("invalid request data supplied, missing Path or DisplayName"), Request: r}
	}

	handler.LogFunction("api.Create").Debugf("will try to create a new bookmark entry: '%s'", payload)

	t = store.Node
	if payload.Type == Folder {
		t = store.Folder
	}

	if err := b.Repository.InUnitOfWork(func(repo store.Repository) error {
		item, err := repo.Create(store.Bookmark{
			DisplayName: payload.DisplayName,
			Path:        payload.Path,
			Type:        t,
			URL:         payload.URL,
			UserName:    user.Username,
			Favicon:     payload.Favicon,
			SortOrder:   payload.SortOrder,
		})
		if err != nil {
			return err
		}
		id = item.ID
		return nil
	}); err != nil {
		handler.LogFunction("api.Create").Errorf("could not create a new bookmark: %v", err)
		return errors.ServerError{Err: fmt.Errorf("error creating a new bookmark: %v", err), Request: r}
	}

	handler.LogFunction("api.Create").Infof("bookmark created with ID: %s", id)
	return render.Render(w, r, ResultResponse{
		Result: &Result{
			Success: true,
			Message: fmt.Sprintf("Bookmark created with ID '%s'", id),
			Value:   id,
		},
		Status: http.StatusCreated,
	})
}

// swagger:operation PUT /api/v1/bookmarks bookmarks UpdateBookmark
//
// update a bookmark
//
// use the supplied payload to update a existing bookmark
//
// ---
// consumes:
// - application/json
// produces:
// - application/json
// responses:
//   '200':
//     description: Result
//     schema:
//       "$ref": "#/definitions/Result"
//   '400':
//     description: ProblemDetail
//     schema:
//       "$ref": "#/definitions/ProblemDetail"
//   '401':
//     description: ProblemDetail
//     schema:
//       "$ref": "#/definitions/ProblemDetail"
//   '403':
//     description: ProblemDetail
//     schema:
//       "$ref": "#/definitions/ProblemDetail"
func (b *BookmarksAPI) Update(user security.User, w http.ResponseWriter, r *http.Request) error {
	var (
		id      string
		payload *BookmarkRequest
	)

	payload = &BookmarkRequest{}
	if err := render.Bind(r, payload); err != nil {
		handler.LogFunction("api.Update").Warnf("cannot bind payload: '%v'", err)
		return errors.BadRequestError{Err: fmt.Errorf("invalid request data supplied"), Request: r}
	}

	if payload.Path == "" || payload.DisplayName == "" || payload.ID == "" {
		handler.LogFunction("api.Update").Warnf("required fields of bookmarks missing")
		return errors.BadRequestError{Err: fmt.Errorf("invalid request data supplied, missing ID, Path or DisplayName"), Request: r}

	}

	handler.LogFunction("api.Update").Debugf("will try to update existing bookmark entry: '%s'", payload)

	if err := b.Repository.InUnitOfWork(func(repo store.Repository) error {
		// 1) fetch the existing bookmark by id
		existing, err := repo.GetBookmarkById(payload.ID, user.Username)
		if err != nil {
			handler.LogFunction("api.Update").Warnf("could not find bookmark by id '%s': %v", payload.ID, err)
			return err
		}
		childCount := existing.ChildCount
		if existing.Type == store.Folder {
			// 2) ensure that the existing folder is not moved to itself
			folderPath := ensureFolderPath(existing.Path, existing.DisplayName)
			if payload.Path == folderPath {
				handler.LogFunction("api.Update").Warnf("a folder cannot be moved into itself: folder-path: '%s', destination: '%s'", folderPath, payload.Path)
				return errors.BadRequestError{Err: fmt.Errorf("cannot move folder into itself"), Request: r}
			}

			// 3) get the folder child-count
			// on save of a folder, update the child-count
			parentPath := existing.Path
			path := ensureFolderPath(parentPath, existing.DisplayName)
			nodeCount, err := repo.GetPathChildCount(path, user.Username)
			if err != nil {
				handler.LogFunction("api.Update").Warnf("could not get child-count of path '%s': %v", path, err)
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
		item, err := repo.Update(store.Bookmark{
			ID:          payload.ID,
			Created:     existing.Created,
			DisplayName: payload.DisplayName,
			Path:        payload.Path,
			Type:        existing.Type, // it does not make any sense to change the type of a bookmark!
			URL:         payload.URL,
			SortOrder:   payload.SortOrder,
			UserName:    user.Username,
			ChildCount:  childCount,
			Favicon:     payload.Favicon,
			AccessCount: payload.AccessCount,
		})
		if err != nil {
			handler.LogFunction("api.Update").Warnf("could not update bookmark: %v", err)
			return err
		}
		id = item.ID

		// also update the favicon if not available
		if item.Favicon == "" {
			// fire&forget, run this in background and do not wait for the result
			go b.fetchFavicon(item, user)
		}

		if existing.Type == store.Folder && (existingDisplayName != payload.DisplayName || existingPath != payload.Path) {
			// if we have a folder and change the displayname or the parent-path, this also affects ALL sub-elements
			// therefore all paths of sub-elements where this folder-path is present, need to be updated
			newPath := ensureFolderPath(payload.Path, payload.DisplayName)
			oldPath := ensureFolderPath(existingPath, existingDisplayName)

			handler.LogFunction("api.Update").Warnf("will update all old paths '%s' to new path '%s'", oldPath, newPath)

			bookmarks, err := repo.GetBookmarksByPathStart(oldPath, user.Username)
			if err != nil {
				handler.LogFunction("api.Update").Warnf("could not get bookmarks by path '%s': %v", oldPath, err)
				return err
			}

			for _, bm := range bookmarks {
				updatePath := strings.ReplaceAll(bm.Path, oldPath, newPath)
				if _, err := repo.Update(store.Bookmark{
					ID:          bm.ID,
					Created:     bm.Created,
					DisplayName: bm.DisplayName,
					Path:        updatePath,
					SortOrder:   bm.SortOrder,
					Type:        bm.Type,
					URL:         bm.URL,
					UserName:    user.Username,
					ChildCount:  bm.ChildCount,
					AccessCount: bm.AccessCount,
					Favicon:     bm.Favicon,
				}); err != nil {
					handler.LogFunction("api.Update").Warnf("cannot update bookmark path: %v", err)
					return err
				}
			}
		}

		// if the path has changed - update the childcount of affected paths
		if existingPath != payload.Path {
			// the affected paths are the origin-path and the destination-path
			if err := updateChildCountOfPath(existingPath, user.Username, repo); err != nil {
				handler.LogFunction("api.Update").Errorf("could not update child-count of path '%s': %v", existingPath, err)
				return err
			}
			// and the destination-path
			if err := updateChildCountOfPath(payload.Path, user.Username, repo); err != nil {
				handler.LogFunction("api.Update").Errorf("could not update child-count of path '%s': %v", payload.Path, err)
				return err
			}
		}

		return nil
	}); err != nil {
		handler.LogFunction("api.Update").Errorf("could not update bookmark because of error: %v", err)

		var badRequest errors.BadRequestError
		if er.As(err, &badRequest) {
			return badRequest
		} else {
			return errors.ServerError{Err: fmt.Errorf("error updating bookmark: %v", err), Request: r}
		}
	}

	handler.LogFunction("api.Update").Infof("updated bookmark with ID '%s'", id)

	return render.Render(w, r, ResultResponse{
		Result: &Result{
			Success: true,
			Message: fmt.Sprintf("Bookmark with ID '%s' was updated", id),
			Value:   id,
		},
	})
}

func updateChildCountOfPath(path, username string, repo store.Repository) error {
	if path == "/" {
		handler.LogFunction("api.updateChildCountOfPath").Debugf("skip the ROOT path '/'")
		return nil
	}

	folder, err := repo.GetFolderByPath(path, username)
	if err != nil {
		handler.LogFunction("api.updateChildCountOfPath").Errorf("cannot get the folder for path '%s': %v", path, err)
		return err
	}

	nodeCount, err := repo.GetPathChildCount(path, username)
	if err != nil {
		handler.LogFunction("api.updateChildCountOfPath").Errorf("could not get the child-count of the path '%s': %v", path, err)
		return err
	}

	var childCount int
	if len(nodeCount) > 0 {
		// use the firest element, because the count was queried for a specific path
		childCount = nodeCount[0].Count
	}

	folder.ChildCount = childCount

	if _, err := repo.Update(folder); err != nil {
		handler.LogFunction("api.updateChildCountOfPath").Errorf("could not update the child-count of folder '%s': %v", path, err)
		return err
	}
	return nil
}

// swagger:operation Delete /api/v1/bookmarks/{id} bookmarks DeleteBookmark
//
// delete a bookmark
//
// delete the bookmark identified by the supplied id
//
// ---
// produces:
// - application/json
// parameters:
// - name: id
//   in: path
// responses:
//   '200':
//     description: Result
//     schema:
//       "$ref": "#/definitions/Result"
//   '400':
//     description: ProblemDetail
//     schema:
//       "$ref": "#/definitions/ProblemDetail"
//   '401':
//     description: ProblemDetail
//     schema:
//       "$ref": "#/definitions/ProblemDetail"
//   '403':
//     description: ProblemDetail
//     schema:
//       "$ref": "#/definitions/ProblemDetail"
func (b *BookmarksAPI) Delete(user security.User, w http.ResponseWriter, r *http.Request) error {
	id := chi.URLParam(r, "id")

	if id == "" {
		return errors.BadRequestError{Err: fmt.Errorf("missing id parameter"), Request: r}
	}

	handler.LogFunction("api.Delete").Debugf("will try to delete bookmark with ID '%s'", id)

	if err := b.Repository.InUnitOfWork(func(repo store.Repository) error {
		// 1) fetch the existing bookmark by id
		existing, err := repo.GetBookmarkById(id, user.Username)
		if err != nil {
			handler.LogFunction("api.Update").Warnf("could not find bookmark by id '%s': %v", id, err)
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
		handler.LogFunction("api.Delete").Errorf("could not delete bookmark because of error: %v", err)
		return errors.ServerError{Err: fmt.Errorf("error deleting bookmark: %v", err), Request: r}
	}

	return render.Render(w, r, ResultResponse{
		Result: &Result{
			Success: true,
			Message: fmt.Sprintf("Bookmark with ID '%s' was deleted", id),
			Value:   id,
		},
	})
}

// swagger:operation PUT /api/v1/bookmarks/sortorder bookmarks UpdateSortOrder
//
// change the sortorder of bookmarks
//
// provide a new sortorder for a list of IDS
//
// ---
// consumes:
// - application/json
// produces:
// - application/json
// responses:
//   '200':
//     description: Result
//     schema:
//       "$ref": "#/definitions/Result"
//   '400':
//     description: ProblemDetail
//     schema:
//       "$ref": "#/definitions/ProblemDetail"
//   '401':
//     description: ProblemDetail
//     schema:
//       "$ref": "#/definitions/ProblemDetail"
//   '403':
//     description: ProblemDetail
//     schema:
//       "$ref": "#/definitions/ProblemDetail"
func (b *BookmarksAPI) UpdateSortOrder(user security.User, w http.ResponseWriter, r *http.Request) error {
	payload := &BookmarksSortOrderRequest{}
	if err := render.Bind(r, payload); err != nil {
		handler.LogFunction("api.UpdateSortOrder").Errorf("cannot bind payload: '%v'", err)
		return errors.BadRequestError{Err: fmt.Errorf("invalid request data supplied"), Request: r}
	}

	if len(payload.IDs) != len(payload.SortOrder) {
		handler.LogFunction("api.UpdateSortOrder").Errorf("ids and sortorders do not match!")
		return errors.BadRequestError{Err: fmt.Errorf("invalid request data supplied, IDs and SortOrder do not match"), Request: r}
	}

	var updates int
	if err := b.Repository.InUnitOfWork(func(repo store.Repository) error {
		for i, item := range payload.IDs {
			bm, err := repo.GetBookmarkById(item, user.Username)
			if err != nil {
				handler.LogFunction("api.UpdateSortOrder").Errorf("could not get bookmark by id '%s', %v", item, err)
				return err
			}

			handler.LogFunction("api.UpdateSortOrder").Debugf("will update sortOrder of bookmark '%s' with value %d", bm.DisplayName, payload.SortOrder[i])
			bm.SortOrder = payload.SortOrder[i]
			_, err = repo.Update(bm)
			if err != nil {
				handler.LogFunction("api.UpdateSortOrder").Errorf("could not update bookmark: %v", err)
				return err
			}
			updates += 1
		}
		return nil
	}); err != nil {
		handler.LogFunction("api.UpdateSortOrder").Errorf("could not update the sortorder for bookmark: %v", err)
		return errors.ServerError{Err: fmt.Errorf("error updating sortorder of bookmark: %v", err), Request: r}
	}

	return render.Render(w, r, ResultResponse{
		Result: &Result{
			Success: true,
			Message: fmt.Sprintf("Updated '%d' bookmark items.", updates),
			Value:   fmt.Sprintf("%d", updates),
		},
	})
}

// swagger:operation GET /api/v1/bookmarks/fetch bookmarks FetchAndForward
//
// forward to the bookmark
//
// fetch the URL of the specified bookmark and forward to the destination
//
// ---
// produces:
// - application/json
// parameters:
// - name: id
//   in: path
// responses:
//   '302':
//     description: Found, redirect to URL
//   '400':
//     description: ProblemDetail
//     schema:
//       "$ref": "#/definitions/ProblemDetail"
//   '404':
//     description: ProblemDetail
//     schema:
//       "$ref": "#/definitions/ProblemDetail"
//   '401':
//     description: ProblemDetail
//     schema:
//       "$ref": "#/definitions/ProblemDetail"
//   '403':
//     description: ProblemDetail
//     schema:
//       "$ref": "#/definitions/ProblemDetail"
func (b *BookmarksAPI) FetchAndForward(user security.User, w http.ResponseWriter, r *http.Request) error {
	id := chi.URLParam(r, "id")

	if id == "" {
		return errors.BadRequestError{Err: fmt.Errorf("missing id parameter"), Request: r}
	}

	handler.LogFunction("api.FetchAndForward").Debugf("try to fetch bookmark with ID '%s'", id)

	redirectURL := ""
	if err := b.Repository.InUnitOfWork(func(repo store.Repository) error {
		existing, err := repo.GetBookmarkById(id, user.Username)
		if err != nil {
			handler.LogFunction("api.FetchAndForward").Warnf("could not find bookmark by id '%s': %v", id, err)
			return err
		}

		if existing.Type == store.Folder {
			handler.LogFunction("api.FetchAndForward").Warnf("accessCount and redirect only valid for Nodes: ID '%s'", id)
			return errors.BadRequestError{Err: fmt.Errorf("cannot fetch and forward folder - ID '%s'", id), Request: r}
		}

		// update the access-count of nodes
		existing.AccessCount += 1
		if _, err := repo.Update(existing); err != nil {
			handler.LogFunction("api.FetchAndForward").Warnf("could not update bookmark '%s': %v", id, err)
			return err
		}

		redirectURL = existing.URL

		if existing.Favicon == "" {
			// fire&forget, run this in background and do not wait for the result
			go b.fetchFavicon(existing, user)
		}
		return nil
	}); err != nil {
		var badRequest errors.BadRequestError
		handler.LogFunction("api.FetchAndForward").Errorf("could not fetch and update bookmark by ID '%s': %v", id, err)

		if er.As(err, &badRequest) {
			return badRequest
		} else {
			return errors.ServerError{Err: fmt.Errorf("error fetching and updating bookmark: %v", err), Request: r}
		}
	}
	handler.LogFunction("api.FetchAndForward").Debugf("will redirect to bookmark URL '%s'", redirectURL)

	http.Redirect(w, r, redirectURL, http.StatusFound)
	return nil
}

// swagger:operation GET /api/v1/bookmarks/favicon bookmarks GetFavicon
//
// get the favicon from bookmark
//
// return the stored favicon for the given bookmark or return the default favicon
//
// ---
// produces:
// - application/json
// parameters:
// - name: id
//   in: path
// responses:
//   '200':
//     description: Favicon as a file
//   '400':
//     description: ProblemDetail
//     schema:
//       "$ref": "#/definitions/ProblemDetail"
//   '401':
//     description: ProblemDetail
//     schema:
//       "$ref": "#/definitions/ProblemDetail"
//   '403':
//     description: ProblemDetail
//     schema:
//       "$ref": "#/definitions/ProblemDetail"
func (b *BookmarksAPI) GetFavicon(user security.User, w http.ResponseWriter, r *http.Request) error {
	id := chi.URLParam(r, "id")

	if id == "" {
		return errors.BadRequestError{Err: fmt.Errorf("missing id parameter"), Request: r}
	}

	handler.LogFunction("api.GetFavicon").Debugf("try to fetch bookmark with ID '%s'", id)

	existing, err := b.Repository.GetBookmarkById(id, user.Username)
	if err != nil {
		handler.LogFunction("api.GetFavicon").Errorf("could not find bookmark by id '%s': %v", id, err)
		return errors.NotFoundError{Err: fmt.Errorf("could not find bookmark with ID '%s'", id), Request: r}
	}

	fullPath := path.Join(b.BasePath, b.FaviconPath, existing.Favicon)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		handler.LogFunction("api.GetFavicon").Errorf("the specified favicon '%s' is not available", fullPath)
		// not found - use default
		fullPath = path.Join(b.BasePath, b.DefaultFavicon)
	}

	http.ServeFile(w, r, fullPath)
	return nil
}

func (b *BookmarksAPI) fetchFavicon(bm store.Bookmark, user security.User) {
	favicon, payload, err := favicon.GetFaviconFromURL(bm.URL)
	if err != nil {
		handler.LogFunction("api.fetchFavicon").Errorf("cannot fetch favicon from URL '%s': %v", bm.URL, err)
		return
	}

	if len(payload) > 0 {

		hashPayload, err := hashInput(payload)
		if err != nil {
			handler.LogFunction("api.fetchFavicon").Errorf("could not hash payload: '%v'", err)
			return
		}
		hashFilename, err := hashInput([]byte(favicon))
		if err != nil {
			handler.LogFunction("api.fetchFavicon").Errorf("could not hash filename: '%v'", err)
			return
		}
		ext := filepath.Ext(favicon)
		filename := fmt.Sprintf("%s_%s%s", hashFilename, hashPayload, ext)
		fullPath := path.Join(b.BasePath, b.FaviconPath, filename)

		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			handler.LogFunction("api.fetchFavicon").Warnf("got favicon payload length of '%d' for URL '%s'", len(payload), bm.URL)

			if err := ioutil.WriteFile(fullPath, payload, 0644); err != nil {
				handler.LogFunction("api.fetchFavicon").Errorf("could not write favicon to file '%s': %v", fullPath, err)
				return
			}
		}

		// also update the favicon for the bookmark
		if err := b.Repository.InUnitOfWork(func(repo store.Repository) error {
			bm.Favicon = filename
			_, err := repo.Update(bm)
			return err
		}); err != nil {
			handler.LogFunction("api.fetchFavicon").Errorf("could not update bookmark with favicon '%s': %v", filename, err)
		}

	} else {
		handler.LogFunction("api.fetchFavicon").Warnf("not payload for favicon from URL '%s'", bm.URL)
	}
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
// path is valid, with all necessary delimitors
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
