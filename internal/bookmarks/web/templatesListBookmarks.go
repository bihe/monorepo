package web

import (
	"fmt"
	"net/http"
	"strings"

	"golang.binggl.net/monorepo/internal/bookmarks/web/html"
	"golang.binggl.net/monorepo/internal/common"
	base "golang.binggl.net/monorepo/pkg/handler/html"
)

const searchURL = "/bm/search"

// SearchBookmarks performs a search for bookmarks and displays the result using server-side rendering
func (t *TemplateHandler) SearchBookmarks() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		search := queryParam(r, "q")
		user := common.EnsureUser(r)

		t.Logger.InfoRequest(fmt.Sprintf("get bookmarks by name: '%s' for user: '%s'", search, user.Username), r)

		bms, err := t.App.GetBookmarksByName(search, *user)
		if err != nil {
			t.Logger.ErrorRequest(fmt.Sprintf("could not get bookmarks for search '%s'; '%v'", search, err), r)
		}

		ell := html.GetEllipsisValues(r)

		base.Layout(
			t.pageModel("Bookmark Search", search, "/public/bookmarks.svg", *user),
			html.SearchStyles(),
			html.SearchNavigation(search),
			html.SearchContent(search, bms, ell),
			searchURL,
		).Render(w)
	}
}

// SearchBookmarksPartial is used to return only the bookmark list for a search
func (t *TemplateHandler) SearchBookmarksPartial() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		search := queryParam(r, "q")
		user := common.EnsureUser(r)

		t.Logger.InfoRequest(fmt.Sprintf("get search bookmark-list partial by name: '%s' for user: '%s'", search, user.Username), r)

		bms, err := t.App.GetBookmarksByName(search, *user)
		if err != nil {
			t.Logger.ErrorRequest(fmt.Sprintf("could not get bookmarks for search '%s'; '%v'", search, err), r)
		}
		ell := html.GetEllipsisValues(r)
		html.SearchContent(search, bms, ell).Render(w)
	}
}

// GetBookmarksForPath retrieves and renders the bookmarks for a defined path
func (t *TemplateHandler) GetBookmarksForPath() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := pathParam(r, "*")
		if path == "" {
			// start with the root path
			path = "/"
		}
		user := common.EnsureUser(r)

		pathHierarchy := make([]html.BookmarkPathEntry, 1)
		// always start with the root item
		pathHierarchy[0] = html.BookmarkPathEntry{
			UrlPath:     "/",
			DisplayName: "/root",
		}

		// split up the path into it's sub-paths
		pathParts := strings.Split(path, "/")
		// now "grow" the paths so that the preceding part is added to the next part
		var lastPath string
		for _, p := range pathParts {
			if p == "" {
				continue
			}
			pathHierarchy = append(pathHierarchy, html.BookmarkPathEntry{
				UrlPath:     lastPath + "/" + p,
				DisplayName: p,
			})
			lastPath = lastPath + "/" + p
		}
		// highlight the last item
		pathHierarchy[len(pathHierarchy)-1].LastItem = true
		t.Logger.InfoRequest(fmt.Sprintf("get bookmarks for path: '%s' for user: '%s'", path, user.Username), r)

		bms, err := t.App.GetBookmarksByPath(path, *user)
		if err != nil {
			t.Logger.ErrorRequest(fmt.Sprintf("could not get bookmarks for path '%s'; '%v'", path, err), r)
			t.RenderErr(r, w, fmt.Sprintf("could not get bookmarks for path '%s'; '%v'", path, err))
			return
		}

		curFolder := ""
		favicon := ""
		if len(pathParts) > 0 {
			curFolder = pathParts[len(pathParts)-1]
			folder, err := t.App.GetBookmarksFolderByPath(path, *user)
			if err != nil {
				t.Logger.ErrorRequest(fmt.Sprintf("could not get bookmark folder for path '%s'; '%v'", path, err), r)
				t.RenderErr(r, w, fmt.Sprintf("could not get bookmark folder for path '%s'; '%v'", path, err))
				return
			}
			favicon = fmt.Sprintf("/bm/favicon/%s?t=%s", folder.ID, folder.TStamp())

			// we need to treat the root folder in a "special" way
			if path == "/" && strings.HasSuffix(folder.ID, "ROOT") {
				// use a "default" favicon for the root folder
				favicon = "/public/bookmarks.svg"
			}
		}

		ell := html.GetEllipsisValues(r)
		base.Layout(
			t.pageModel(curFolder, "", favicon, *user),
			html.BookmarksByPathStyles(),
			html.BookmarksByPathNavigation(pathHierarchy),
			html.BookmarkList(path, bms, ell),
			searchURL,
		).Render(w)
	}
}

// GetBookmarksForPathPartial only returns the bookmark list of the whole page
func (t *TemplateHandler) GetBookmarksForPathPartial() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := pathParam(r, "*")
		user := common.EnsureUser(r)

		t.Logger.InfoRequest(fmt.Sprintf("get bookmark-list partial for path: '%s' for user: '%s'", path, user.Username), r)

		bms, err := t.App.GetBookmarksByPath(path, *user)
		if err != nil {
			t.Logger.ErrorRequest(fmt.Sprintf("could not get bookmarks for path '%s'; '%v'", path, err), r)
		}
		html.BookmarkList(path, bms, html.GetEllipsisValues(r)).Render(w)
	}
}
