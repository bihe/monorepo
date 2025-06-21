package web

import (
	"fmt"
	"net/http"
	"time"

	"golang.binggl.net/monorepo/internal/bookmarks/app/bookmarks"
	"golang.binggl.net/monorepo/internal/bookmarks/web/html"
	"golang.binggl.net/monorepo/internal/common"
	base "golang.binggl.net/monorepo/pkg/handler/html"
	"golang.binggl.net/monorepo/pkg/logging"
)

// DeleteConfirm shows a confirm dialog before an item is deleted
func (t *TemplateHandler) DeleteConfirm() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := pathParam(r, "id")
		user := common.EnsureUser(r)
		t.Logger.InfoRequest(fmt.Sprintf("get bookmarks by id: '%s' for user: '%s'", id, user.Username), r)

		bm, err := t.App.GetBookmarkByID(id, *user)
		if err != nil {
			t.Logger.ErrorRequest(fmt.Sprintf("could not get bookmarks for id '%s'; '%v'", id, err), r)
		}
		base.DialogConfirmDeleteHx(bm.DisplayName, "/bm/delete/"+bm.ID).Render(w)
	}
}

// DeleteBookmark deletes the given bookmark and replaces the bookmark-list with a new list without the deleted bookmark
func (t *TemplateHandler) DeleteBookmark() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := pathParam(r, "id")
		user := common.EnsureUser(r)

		t.Logger.InfoRequest(fmt.Sprintf("get bookmark by id: '%s' for user: '%s'", id, user.Username), r)
		bm, err := t.App.GetBookmarkByID(id, *user)
		if err != nil {
			t.Logger.ErrorRequest(fmt.Sprintf("could not get bookmark for id '%s'; '%v'", id, err), r)
		}
		err = t.App.Delete(bm.ID, *user)
		if err != nil {
			t.Logger.ErrorRequest(fmt.Sprintf("could not delete bookmark with id '%s'; '%v'", id, err), r)
			triggerRefreshWithToast(w,
				base.MsgError,
				"Bookmark delete error!",
				fmt.Sprintf("Error: '%s'", err))
			return
		}
		triggerRefreshWithToast(w,
			base.MsgSuccess,
			"Bookmark deleted!",
			fmt.Sprintf("The bookmark '%s' was deleted.", bm.DisplayName))

	}
}

// EditBookmarkDialog shows a dialog to edit a bookmark
func (t *TemplateHandler) EditBookmarkDialog() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := pathParam(r, "id")
		user := common.EnsureUser(r)
		t.Logger.InfoRequest(fmt.Sprintf("get bookmarks by id: '%s' for user: '%s'", id, user.Username), r)

		var (
			bm    html.Bookmark
			b     *bookmarks.Bookmark
			paths []string
			err   error
		)
		if id == "-1" {
			bm.ID = html.ValidatorInput{Val: "-1", Valid: true}
			bm.Path = html.ValidatorInput{Val: queryParam(r, "path"), Valid: true}
			bm.DisplayName = html.ValidatorInput{Valid: true}
			bm.URL = html.ValidatorInput{Valid: true}
			bm.CustomFavicon = html.ValidatorInput{Valid: true}
			bm.Type = bookmarks.Node
			bm.TStamp = fmt.Sprintf("%d", time.Now().Unix())
		} else {
			// fetch an existing bookmark
			b, err = t.App.GetBookmarkByID(id, *user)
			if err != nil {
				t.Logger.ErrorRequest(fmt.Sprintf("could not get bookmarks for id '%s'; '%v'", id, err), r)
				t.RenderErr(r, w, fmt.Sprintf("could not get bookmarks for id '%s'; '%v'", id, err))
				return
			}
			bm.ID = html.ValidatorInput{Val: b.ID, Valid: true}
			bm.Path = html.ValidatorInput{Val: b.Path, Valid: true}
			bm.DisplayName = html.ValidatorInput{Val: b.DisplayName, Valid: true}
			bm.URL = html.ValidatorInput{Val: b.URL, Valid: true}
			bm.Type = b.Type
			bm.CustomFavicon = html.ValidatorInput{Valid: true}
			bm.InvertFaviconColor = (b.InvertFaviconColor == 1)
			bm.TStamp = b.TStamp()
			bm.CurrentFavicon = b.Favicon
			paths, err = t.App.GetAllPaths(*user)
			if err != nil {
				t.Logger.ErrorRequest(fmt.Sprintf("could not get all paths for bookmarks; '%v'", err), r)
			}
		}
		html.EditBookmarks(bm, paths).Render(w)
	}
}

// SaveBookmark receives the values from the edit form, validates and persists the data
func (t *TemplateHandler) SaveBookmark() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			recv       bookmarks.Bookmark
			formBm     html.Bookmark
			paths      []string
			err        error
			formPrefix string = "bookmark_"
		)

		err = r.ParseForm()
		if err != nil {
			t.Logger.ErrorRequest(fmt.Sprintf("could not parse supplied form data; '%v'", err), r)
			t.RenderErr(r, w, fmt.Sprintf("could not parse supplied form data; '%v'", err))
			return
		}
		user := common.EnsureUser(r)
		t.Logger.InfoRequest(fmt.Sprintf("save the bookmark data for user: '%s'", user.Username), r)

		customFavicon := false
		recv.ID = r.FormValue(formPrefix + "ID")
		recv.Path = r.FormValue(formPrefix + "Path")
		recv.DisplayName = r.FormValue(formPrefix + "DisplayName")
		recv.URL = r.FormValue(formPrefix + "URL")
		recv.InvertFaviconColor = getIntFromString(r.FormValue(formPrefix + "InvertFaviconColor"))
		recv.Type = bookmarks.Node
		recv.Favicon = r.FormValue(formPrefix + "Favicon")
		if r.FormValue(formPrefix+"Type") == "Folder" {
			recv.Type = bookmarks.Folder
		}

		// validation
		validData := true
		formBm.TStamp = fmt.Sprintf("%d", time.Now().Unix())
		formBm.ID = html.ValidatorInput{Val: recv.ID, Valid: true}
		formBm.DisplayName = html.ValidatorInput{Val: recv.DisplayName, Valid: true}
		if recv.DisplayName == "" {
			formBm.DisplayName.Valid = false
			formBm.DisplayName.Message = "missing value!"
			validData = false
		}
		formBm.Path = html.ValidatorInput{Val: recv.Path, Valid: true}
		if recv.Path == "" {
			formBm.Path.Valid = false
			formBm.Path.Message = "missing value!"
			validData = false
		}
		formBm.Type = recv.Type
		if formBm.Type == bookmarks.Node {
			formBm.URL = html.ValidatorInput{Val: recv.URL, Valid: true}
			if recv.URL == "" {
				formBm.URL.Valid = false
				formBm.URL.Message = "missing value!"
				validData = false
			}
		}
		if customFavicon {
			formBm.UseCustomFavicon = customFavicon
			formBm.CustomFavicon = html.ValidatorInput{Val: recv.Favicon, Valid: true}
			if recv.Favicon == "" {
				formBm.CustomFavicon.Valid = false
				formBm.CustomFavicon.Message = "missing value!"
				validData = false
			}
		}
		formBm.InvertFaviconColor = (recv.InvertFaviconColor == 1)

		if !validData {
			// show the same form again!
			t.Logger.ErrorRequest("the supplied data for creating bookmark entry is not valid!", r)
			html.EditBookmarks(formBm, paths).Render(w)
			return
		}

		if recv.ID == "-1" {
			// create a new bookmark entry
			bm := bookmarks.Bookmark{
				Path:               recv.Path,
				DisplayName:        recv.DisplayName,
				Type:               recv.Type,
				URL:                recv.URL,
				InvertFaviconColor: recv.InvertFaviconColor,
				Favicon:            recv.Favicon,
			}
			if customFavicon {
				// the provided favicon needs to be an ID
				bm.Favicon = recv.Favicon
			}
			created, err := t.App.CreateBookmark(bm, *user)
			if err != nil {
				t.Logger.ErrorRequest(fmt.Sprintf("could not create a new bookmark entry; '%v'", err), r)
				formBm.Error = "Error: " + err.Error()
				html.EditBookmarks(formBm, paths).Render(w)
				return
			}
			t.Logger.Info("new bookmark created", logging.LogV("ID", created.ID))

			triggerRefreshWithToast(w,
				base.MsgSuccess,
				"Bookmark saved!",
				fmt.Sprintf("The bookmark '%s' was created.", created.DisplayName))

			formBm.Close = true
			html.EditBookmarks(formBm, paths).Render(w)
			return

		} else {
			paths, err = t.App.GetAllPaths(*user)
			if err != nil {
				t.Logger.ErrorRequest(fmt.Sprintf("could not get all paths for bookmarks; '%v'", err), r)
			}

			// update an exiting entry
			existing, err := t.App.GetBookmarkByID(recv.ID, *user)
			if err != nil {
				t.Logger.Error("the given bookmark ID is not available", logging.ErrV(err), logging.LogV("ID", recv.ID))
				t.RenderErr(r, w, fmt.Sprintf("the given bookmark '%s' is not available; %v", recv.ID, err))
				return
			}

			// update the existing bookmark with the supplied data
			existing.DisplayName = recv.DisplayName
			existing.Path = recv.Path
			existing.InvertFaviconColor = recv.InvertFaviconColor
			existing.URL = recv.URL
			existing.Favicon = recv.Favicon
			formBm.TStamp = existing.TStamp()
			updated, err := t.App.UpdateBookmark(*existing, *user)
			if err != nil {
				t.Logger.ErrorRequest(fmt.Sprintf("could not update bookmark entry '%s'; '%v'", recv.ID, err), r)
				formBm.Error = "Error: " + err.Error()
				html.EditBookmarks(formBm, paths).Render(w)
				return
			}

			t.Logger.Info("bookmark updated", logging.LogV("ID", updated.ID))

			triggerRefreshWithToast(w,
				base.MsgSuccess,
				"Bookmark saved!",
				fmt.Sprintf("The bookmark '%s' (%s) was updated.", existing.DisplayName, existing.ID))

			formBm.Close = true
			html.EditBookmarks(formBm, paths).Render(w)
			return
		}
	}
}

// SortBookmarks performs a reordering of the bookmark list
func (t *TemplateHandler) SortBookmarks() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := common.EnsureUser(r)
		err := r.ParseForm()
		if err != nil {
			t.Logger.ErrorRequest(fmt.Sprintf("could not parse supplied form data; '%v'", err), r)
			t.RenderErr(r, w, fmt.Sprintf("could not parse supplied form data; '%v'", err))
			return
		}

		idList := r.Form["ID"]
		indexList := make([]int, len(idList))
		t.Logger.Info("will save the new bookmark list", logging.LogV("ID_List", fmt.Sprintf("%v", idList)))
		// we jus ust the order of supplied IDs as the "natural" sort-order
		for i := range idList {
			indexList[i] = i
		}

		sortOrder := bookmarks.BookmarksSortOrder{
			IDs:       idList,
			SortOrder: indexList,
		}

		updates, err := t.App.UpdateSortOrder(sortOrder, *user)
		if err != nil {

			triggerToast(w,
				base.MsgError,
				"Error sorting!",
				fmt.Sprintf("Could not perform sorting: %v", err))
			return
		}

		triggerRefreshWithToast(w,
			base.MsgSuccess,
			"List sorted!",
			fmt.Sprintf("%d bookmarks were successfully sorted", updates))
	}
}
