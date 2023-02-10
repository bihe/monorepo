package api

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"golang.binggl.net/monorepo/internal/bookmarks/app"
	"golang.binggl.net/monorepo/internal/bookmarks/app/bookmarks"
)

// BookmarksHandlers implements the API used for bookmarks
type BookmarksHandler struct {
	App *bookmarks.Application
}

// GetBookmarkByID retrieves a bookmark by id
func (b *BookmarksHandler) GetBookmarkByID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		user := ensureUser(r)

		bm, err := b.App.GetBookmarkByID(id, *user)
		if err != nil {
			encodeError(err, b.App.Logger, w, r)
			return
		}
		respondJSON(w, bm)
	}
}

// GetBookmarksByPath retrieves bookmarks by path
func (b *BookmarksHandler) GetBookmarksByPath() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := queryParam(r, "path")
		user := ensureUser(r)

		b.App.Logger.InfoRequest(fmt.Sprintf("get bookmarks by path: '%s' for user: '%s'", path, user.Username), r)

		bms, err := b.App.GetBookmarksByPath(path, *user)
		if err != nil {
			encodeError(err, b.App.Logger, w, r)
			return
		}
		count := len(bms)
		respondJSON(w, BookmarkList{
			Success: true,
			Count:   count,
			Message: fmt.Sprintf("Found %d items.", count),
			Value:   bms,
		})
	}
}

// GetBookmarksFolderByPath retrieve bookmark folder by path
func (b *BookmarksHandler) GetBookmarksFolderByPath() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := queryParam(r, "path")
		user := ensureUser(r)

		b.App.Logger.InfoRequest(fmt.Sprintf("get bookmarks-folder by path: '%s' for user: '%s'", path, user.Username), r)

		bm, err := b.App.GetBookmarksFolderByPath(path, *user)
		if err != nil {
			encodeError(err, b.App.Logger, w, r)
			return
		}
		respondJSON(w, BookmarkFolderResult{
			Success: true,
			Message: fmt.Sprintf("Found bookmark folder for path %s.", path),
			Value:   *bm,
		})
	}
}

// GetAllPaths returns all paths
func (b *BookmarksHandler) GetAllPaths() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		user := ensureUser(r)

		paths, err := b.App.GetAllPaths(*user)
		if err != nil {
			encodeError(err, b.App.Logger, w, r)
			return
		}
		respondJSON(w, BookmarksPaths{
			Paths: paths,
			Count: len(paths),
		})
	}
}

// GetBookmarksByName retrieve bookmarks by name
func (b *BookmarksHandler) GetBookmarksByName() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := queryParam(r, "name")
		user := ensureUser(r)

		b.App.Logger.InfoRequest(fmt.Sprintf("get bookmarks by name: '%s' for user: '%s'", name, user.Username), r)

		bms, err := b.App.GetBookmarksByName(name, *user)
		if err != nil {
			encodeError(err, b.App.Logger, w, r)
			return
		}
		count := len(bms)
		respondJSON(w, BookmarkList{
			Success: true,
			Count:   count,
			Message: fmt.Sprintf("Found %d items.", count),
			Value:   bms,
		})
	}
}

// GetFavicon for the specified bookmark
func (b *BookmarksHandler) GetFavicon() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := ensureUser(r)
		id := chi.URLParam(r, "id")

		b.App.Logger.InfoRequest(fmt.Sprintf("try to fetch bookmark with ID '%s'", id), r)
		favicon, err := b.App.GetFavicon(id, *user)
		if err != nil {
			encodeError(err, b.App.Logger, w, r)
			return
		}
		http.ServeContent(w, r, favicon.Name, favicon.Modified, bytes.NewReader(favicon.Payload))
	}
}

// ---- CRUD ----

// Create a new bookmark
func (b *BookmarksHandler) Create() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := ensureUser(r)

		payload := &BookmarkRequest{}
		if err := render.Bind(r, payload); err != nil {
			b.App.Logger.ErrorRequest(fmt.Sprintf("cannot bind payload: '%v'", err), r)
			encodeError(app.ErrValidation(err.Error()), b.App.Logger, w, r)
			return
		}

		b.App.Logger.InfoRequest(fmt.Sprintf("will try to create a new bookmark entry: '%s'", payload), r)

		bm, err := b.App.CreateBookmark(*payload.Bookmark, *user)
		if err != nil {
			encodeError(err, b.App.Logger, w, r)
			return
		}
		w.WriteHeader(http.StatusCreated)
		respondJSON(w, Result{
			Success: true,
			Message: fmt.Sprintf("Bookmark created with ID '%s'", bm.ID),
			Value:   bm.ID,
		})

	}
}

// Update a bookmark
func (b *BookmarksHandler) Update() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := ensureUser(r)

		payload := &BookmarkRequest{}
		if err := render.Bind(r, payload); err != nil {
			b.App.Logger.ErrorRequest(fmt.Sprintf("cannot bind payload: '%v'", err), r)
			encodeError(app.ErrValidation(err.Error()), b.App.Logger, w, r)
			return
		}

		b.App.Logger.InfoRequest(fmt.Sprintf("will try to update existing bookmark entry: '%s'", payload), r)

		bm, err := b.App.Update(*payload.Bookmark, *user)
		if err != nil {
			encodeError(err, b.App.Logger, w, r)
			return
		}
		respondJSON(w, Result{
			Success: true,
			Message: fmt.Sprintf("Bookmark with ID '%s' was updated", bm.ID),
			Value:   bm.ID,
		})

	}
}

// Delete a bookmark by id
func (b *BookmarksHandler) Delete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := ensureUser(r)
		id := chi.URLParam(r, "id")

		b.App.Logger.InfoRequest(fmt.Sprintf("will try to delete bookmark with ID '%s'", id), r)

		err := b.App.Delete(id, *user)
		if err != nil {
			encodeError(err, b.App.Logger, w, r)
			return
		}
		respondJSON(w, Result{
			Success: true,
			Message: fmt.Sprintf("Bookmark with ID '%s' was deleted", id),
			Value:   id,
		})
	}
}

// UpdateSortOrder modifies the display sort-order
func (b *BookmarksHandler) UpdateSortOrder() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := ensureUser(r)
		payload := &BookmarksSortOrderRequest{}

		if err := render.Bind(r, payload); err != nil {
			b.App.Logger.ErrorRequest(fmt.Sprintf("cannot bind payload: '%v'", err), r)
			encodeError(app.ErrValidation(err.Error()), b.App.Logger, w, r)
			return
		}

		updates, err := b.App.UpdateSortOrder(*payload.BookmarksSortOrder, *user)
		if err != nil {
			encodeError(err, b.App.Logger, w, r)
			return
		}

		respondJSON(w, Result{
			Success: true,
			Message: fmt.Sprintf("Updated '%d' bookmark items.", updates),
			Value:   fmt.Sprintf("%d", updates),
		})
	}
}

// FetchAndForward retrievs a bookmark by id and forwards to the url of the bookmark
func (b *BookmarksHandler) FetchAndForward() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := ensureUser(r)
		id := chi.URLParam(r, "id")

		b.App.Logger.InfoRequest(fmt.Sprintf("try to fetch bookmark with ID '%s'", id), r)

		redirectURL, err := b.App.FetchAndForward(id, *user)
		if err != nil {
			encodeError(err, b.App.Logger, w, r)
			return
		}

		b.App.Logger.InfoRequest(fmt.Sprintf("will redirect to bookmark URL '%s'", redirectURL), r)
		http.Redirect(w, r, redirectURL, http.StatusFound)
	}
}
