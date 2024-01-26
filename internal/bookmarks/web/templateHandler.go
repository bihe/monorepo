package web

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"golang.binggl.net/monorepo/internal/bookmarks/app/bookmarks"
	"golang.binggl.net/monorepo/internal/bookmarks/web/templates"
	"golang.binggl.net/monorepo/internal/common"
	"golang.binggl.net/monorepo/pkg/config"
	"golang.binggl.net/monorepo/pkg/develop"
	"golang.binggl.net/monorepo/pkg/handler"
	tmpl "golang.binggl.net/monorepo/pkg/handler/templates"
	"golang.binggl.net/monorepo/pkg/logging"
	"golang.binggl.net/monorepo/pkg/security"
)

// TemplateHandler takes care of providing HTML templates.
// This is the new approach with a template + htmx based UI to replace the angular frontend
// and have a more go-oriented approach towards UI and user-interaction. This reduces the
// cognitive load because less technology mix is needed. Via template + htmx approach only
// a limited amount of javascript is needed to achieve the frontend.
// As additional benefit the build should be faster, because the nodejs build can be removed
type TemplateHandler struct {
	*handler.TemplateHandler
	App     *bookmarks.Application
	Version string
	Build   string
}

// SearchBookmarks performs a search for bookmarks and displays the result using server-side rendering
func (t *TemplateHandler) SearchBookmarks() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		search := queryParam(r, "q")
		user := ensureUser(r)

		t.Logger.InfoRequest(fmt.Sprintf("get bookmarks by name: '%s' for user: '%s'", search, user.Username), r)

		bms, err := t.App.GetBookmarksByName(search, *user)
		if err != nil {
			t.Logger.ErrorRequest(fmt.Sprintf("could not get bookmarks for search '%s'; '%v'", search, err), r)
		}

		ell := GetEllipsisValues(r)
		tmpl.Layout(
			t.layoutModel("Bookmark Search", search, "/public/folder.svg", *user),
			templates.SearchStyles(),
			templates.SearchNavigation(search),
			templates.SearchContent(bms, ell),
		).Render(r.Context(), w)
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
		user := ensureUser(r)

		pathHierarchy := make([]templates.BookmarkPathEntry, 1)
		// always start with the root item
		pathHierarchy[0] = templates.BookmarkPathEntry{
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
			pathHierarchy = append(pathHierarchy, templates.BookmarkPathEntry{
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
			favicon = "/api/v1/bookmarks/favicon/" + folder.ID
		}

		ell := GetEllipsisValues(r)
		tmpl.Layout(
			t.layoutModel(curFolder, "", favicon, *user),
			templates.BookmarksByPathStyles(),
			templates.BookmarksByPathNavigation(pathHierarchy),
			templates.BookmarksByPathContent(templates.BookmarkList(
				path,
				bms,
				ell,
			)),
		).Render(r.Context(), w)
	}
}

// GetBookmarksForPathPartial only returns the bookmark list of the whole page
func (t *TemplateHandler) GetBookmarksForPathPartial() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := pathParam(r, "*")
		user := ensureUser(r)

		t.Logger.InfoRequest(fmt.Sprintf("get bookmark-list partial for path: '%s' for user: '%s'", path, user.Username), r)

		bms, err := t.App.GetBookmarksByPath(path, *user)
		if err != nil {
			t.Logger.ErrorRequest(fmt.Sprintf("could not get bookmarks for path '%s'; '%v'", path, err), r)
		}
		ell := GetEllipsisValues(r)
		templates.BookmarkList(
			path,
			bms,
			ell,
		).Render(r.Context(), w)
	}
}

// DeleteConfirm shows a confirm dialog before an item is deleted
func (t *TemplateHandler) DeleteConfirm() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := pathParam(r, "id")
		user := ensureUser(r)
		t.Logger.InfoRequest(fmt.Sprintf("get bookmarks by id: '%s' for user: '%s'", id, user.Username), r)

		bm, err := t.App.GetBookmarkByID(id, *user)
		if err != nil {
			t.Logger.ErrorRequest(fmt.Sprintf("could not get bookmarks for id '%s'; '%v'", id, err), r)
		}
		templates.DialogConfirmDelete(bm.DisplayName, bm.ID).Render(r.Context(), w)
	}
}

// DeleteBookmark deletes the given bookmark and replaces the bookmark-list with a new list without the deleted bookmark
func (t *TemplateHandler) DeleteBookmark() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := pathParam(r, "id")
		user := ensureUser(r)
		ell := GetEllipsisValues(r)

		t.Logger.InfoRequest(fmt.Sprintf("get bookmark by id: '%s' for user: '%s'", id, user.Username), r)
		bm, err := t.App.GetBookmarkByID(id, *user)
		if err != nil {
			t.Logger.ErrorRequest(fmt.Sprintf("could not get bookmark for id '%s'; '%v'", id, err), r)
		}
		err = t.App.Delete(bm.ID, *user)
		if err != nil {
			t.Logger.ErrorRequest(fmt.Sprintf("could not delete bookmark with id '%s'; '%v'", id, err), r)

			bms, bErr := t.App.GetBookmarksByPath(bm.Path, *user)
			if bErr != nil {
				t.Logger.ErrorRequest(fmt.Sprintf("could not get bookmarks for path '%s'; '%v'", bm.Path, bErr), r)
			}

			// show a notification-toast about the error!
			w.Header().Add("HX-Trigger", tmpl.ErrorToast("Bookmark delete error", fmt.Sprintf("Error: '%s'", err)))
			templates.BookmarkList(
				bm.Path,
				bms,
				ell,
			).Render(r.Context(), w)
			return
		}

		bms, err := t.App.GetBookmarksByPath(bm.Path, *user)
		if err != nil {
			t.Logger.ErrorRequest(fmt.Sprintf("could not get bookmarks for path '%s'; '%v'", bm.Path, err), r)
		}

		// show a notification-toast about the update!
		// https://htmx.org/headers/hx-trigger/
		w.Header().Add("HX-Trigger", tmpl.SuccessToast("Bookmark deleted", fmt.Sprintf("The bookmark '%s' was deleted.", bm.DisplayName)))
		templates.BookmarkList(
			bm.Path,
			bms,
			ell,
		).Render(r.Context(), w)

	}
}

// EditBookmarkDialog shows a dialog to edit a bookmark
func (t *TemplateHandler) EditBookmarkDialog() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := pathParam(r, "id")
		user := ensureUser(r)
		t.Logger.InfoRequest(fmt.Sprintf("get bookmarks by id: '%s' for user: '%s'", id, user.Username), r)

		var (
			bm    templates.Bookmark
			b     *bookmarks.Bookmark
			paths []string
			err   error
		)
		if id == "-1" {
			bm.ID = templates.ValidatorInput{Val: "-1", Valid: true}
			bm.Path = templates.ValidatorInput{Val: queryParam(r, "path"), Valid: true}
			bm.DisplayName = templates.ValidatorInput{Valid: true}
			bm.URL = templates.ValidatorInput{Valid: true}
			bm.CustomFavicon = templates.ValidatorInput{Valid: true}
			bm.Type = bookmarks.Node
		} else {
			// fetch an existing bookmark
			b, err = t.App.GetBookmarkByID(id, *user)
			if err != nil {
				t.Logger.ErrorRequest(fmt.Sprintf("could not get bookmarks for id '%s'; '%v'", id, err), r)
				t.RenderErr(r, w, fmt.Sprintf("could not get bookmarks for id '%s'; '%v'", id, err))
				return
			}
			bm.ID = templates.ValidatorInput{Val: b.ID, Valid: true}
			bm.Path = templates.ValidatorInput{Val: b.Path, Valid: true}
			bm.DisplayName = templates.ValidatorInput{Val: b.DisplayName, Valid: true}
			bm.URL = templates.ValidatorInput{Val: b.URL, Valid: true}
			bm.Type = b.Type
			bm.CustomFavicon = templates.ValidatorInput{Valid: true}
			bm.InvertFaviconColor = (b.InvertFaviconColor == 1)

			paths, err = t.App.GetAllPaths(*user)
			if err != nil {
				t.Logger.ErrorRequest(fmt.Sprintf("could not get all paths for bookmarks; '%v'", err), r)
			}
		}
		templates.EditBookmarks(bm, paths).Render(r.Context(), w)
	}
}

const errorFavicon = `<span id="bookmark_favicon_display" class="error_icon">
<i id="error_tooltip_favicon" class="position_error_icon bi bi-exclamation-square-fill" data-bs-toggle="tooltip" data-bs-title="%s"></i>
</span>
<script type="text/javascript">
[...document.querySelectorAll('[data-bs-toggle="tooltip"]')].map(tooltipTriggerEl => new bootstrap.Tooltip(tooltipTriggerEl))
</script>
`
const favIconImage = `<img id="bookmark_favicon_display" class="bookmark_favicon_preview" src="/api/v1/bookmarks/favicon/temp/%s">
<input type="hidden" name="bookmark_Favicon" value="%s"/>`

// FetchCustomFaviconURL fetches the given custom favicon URL and returns a new image
func (t *TemplateHandler) FetchCustomFaviconURL() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		favURL := r.FormValue("bookmark_CustomFavicon")
		fav, err := t.App.LocalFetchFaviconURL(favURL)
		if err != nil {
			errMsg := strings.ReplaceAll(err.Error(), "\"", "'")
			t.Logger.ErrorRequest(fmt.Sprintf("could not fetch the custom favicon; '%v'", err), r)
			w.Write([]byte(fmt.Sprintf(errorFavicon, errMsg)))
			return
		}
		w.Write([]byte(fmt.Sprintf(favIconImage, fav.Name, fav.Name)))
	}
}

// FetchCustomFaviconFromPage tries to fetch the favicon from the given Page-URL
func (t *TemplateHandler) FetchCustomFaviconFromPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		pageUrl := r.FormValue("bookmark_URL")
		fav, err := t.App.LocalExtractFaviconFromURL(pageUrl)
		if err != nil {
			errMsg := strings.ReplaceAll(err.Error(), "\"", "'")
			t.Logger.ErrorRequest(fmt.Sprintf("could not fetch the custom favicon; '%v'", err), r)
			w.Write([]byte(fmt.Sprintf(errorFavicon, errMsg)))
			return
		}
		w.Write([]byte(fmt.Sprintf(favIconImage, fav.Name, fav.Name)))
	}
}

type triggerDef struct {
	tmpl.ToastMessage
	Refresh string `json:"refreshBookmarkList,omitempty"`
}

// SaveBookmark receives the values from the edit form, validates and persists the data
func (t *TemplateHandler) SaveBookmark() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			recv       bookmarks.Bookmark
			formBm     templates.Bookmark
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
		user := ensureUser(r)
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
		formBm.ID = templates.ValidatorInput{Val: recv.ID, Valid: true}
		formBm.DisplayName = templates.ValidatorInput{Val: recv.DisplayName, Valid: true}
		if recv.DisplayName == "" {
			formBm.DisplayName.Valid = false
			formBm.DisplayName.Message = "missing value!"
			validData = false
		}
		formBm.Path = templates.ValidatorInput{Val: recv.Path, Valid: true}
		if recv.Path == "" {
			formBm.Path.Valid = false
			formBm.Path.Message = "missing value!"
			validData = false
		}
		formBm.Type = recv.Type
		if formBm.Type == bookmarks.Node {
			formBm.URL = templates.ValidatorInput{Val: recv.URL, Valid: true}
			if recv.URL == "" {
				formBm.URL.Valid = false
				formBm.URL.Message = "missing value!"
				validData = false
			}
		}
		if customFavicon {
			formBm.UseCustomFavicon = customFavicon
			formBm.CustomFavicon = templates.ValidatorInput{Val: recv.Favicon, Valid: true}
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
			templates.EditBookmarks(formBm, paths).Render(r.Context(), w)
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
				templates.EditBookmarks(formBm, paths).Render(r.Context(), w)
				return
			}
			t.Logger.Info("new bookmark created", logging.LogV("ID", created.ID))

			triggerEvent := triggerDef{
				ToastMessage: tmpl.ToastMessage{
					Event: tmpl.ToastMessageContent{
						Type:  tmpl.MsgSuccess,
						Title: "Bookmark saved!",
						Text:  fmt.Sprintf("The bookmark '%s' was created.", created.DisplayName),
					},
				},
				Refresh: "now",
			}
			// https://htmx.org/headers/hx-trigger/
			w.Header().Add("HX-Trigger", tmpl.Json(triggerEvent))
			formBm.Close = true
			templates.EditBookmarks(formBm, paths).Render(r.Context(), w)
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

			updated, err := t.App.UpdateBookmark(*existing, *user)
			if err != nil {
				t.Logger.ErrorRequest(fmt.Sprintf("could not update bookmark entry '%s'; '%v'", recv.ID, err), r)
				formBm.Error = "Error: " + err.Error()
				templates.EditBookmarks(formBm, paths).Render(r.Context(), w)
				return
			}

			t.Logger.Info("bookmark updated", logging.LogV("ID", updated.ID))

			triggerEvent := triggerDef{
				ToastMessage: tmpl.ToastMessage{
					Event: tmpl.ToastMessageContent{
						Type:  tmpl.MsgSuccess,
						Title: "Bookmark saved!",
						Text:  fmt.Sprintf("The bookmark '%s' (%s) was updated.", existing.DisplayName, existing.ID),
					},
				},
				Refresh: "now",
			}
			// https://htmx.org/headers/hx-trigger/
			w.Header().Add("HX-Trigger", tmpl.Json(triggerEvent))
			formBm.Close = true
			templates.EditBookmarks(formBm, paths).Render(r.Context(), w)
			return
		}
	}
}

// SortBookmarks performs a reordering of the bookmark list
func (t *TemplateHandler) SortBookmarks() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := ensureUser(r)
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
			triggerEvent := triggerDef{
				ToastMessage: tmpl.ToastMessage{
					Event: tmpl.ToastMessageContent{
						Type:  tmpl.MsgError,
						Title: "Error sorting!",
						Text:  fmt.Sprintf("Could not perform sorting: %v", err),
					},
				},
			}
			// https://htmx.org/headers/hx-trigger/
			w.Header().Add("HX-Trigger", tmpl.Json(triggerEvent))
			return
		}

		triggerEvent := triggerDef{
			ToastMessage: tmpl.ToastMessage{
				Event: tmpl.ToastMessageContent{
					Type:  tmpl.MsgSuccess,
					Title: "List sorted!",
					Text:  fmt.Sprintf("%d bookmarks were successfully sorted", updates),
				},
			},
			Refresh: "now",
		}
		// https://htmx.org/headers/hx-trigger/
		w.Header().Add("HX-Trigger", tmpl.Json(triggerEvent))

	}
}

// --------------------------------------------------------------------------
//  UI Ellipsis Handling
// --------------------------------------------------------------------------

// Desktop Browser
var StdEllipsis = templates.EllipsisValues{
	PathLen:   50,
	NodeLen:   60,
	FolderLen: 50,
}

// Mobile View
var MobileEllipsis = templates.EllipsisValues{
	PathLen:   5,
	NodeLen:   30,
	FolderLen: 20,
}

func GetEllipsisValues(r *http.Request) (ell templates.EllipsisValues) {
	vX, _ := getViewPort(r)
	ell = StdEllipsis
	if vX == 0 {
		// this looks odd - use the std
		return
	}
	// iPhone 12 Pro
	if vX <= 390 {
		ell = MobileEllipsis
	}
	return
}

const cookieViewPortName = "viewport"

func getViewPort(r *http.Request) (x, y int) {
	cookie, _ := r.Cookie(cookieViewPortName)
	if cookie != nil && cookie.Value != "" {
		dim := strings.Split(cookie.Value, ":")
		if len(dim) == 2 {
			if v, err := strconv.Atoi(dim[0]); err == nil {
				x = v
			}
			if v, err := strconv.Atoi(dim[1]); err == nil {
				y = v
			}
		}
	}
	return
}

// --------------------------------------------------------------------------
//  Internals
// --------------------------------------------------------------------------

func (t *TemplateHandler) versionString() string {
	return fmt.Sprintf("%s-%s", t.Version, t.Build)
}

func (t *TemplateHandler) layoutModel(pageTitle, search, favicon string, user security.User) tmpl.LayoutModel {
	appNav := make([]tmpl.NavItem, 0)
	var title string
	for _, a := range common.AvailableApps {
		if a.URL == "/bm" {
			a.Active = true
			title = a.DisplayName
		}
		appNav = append(appNav, a)
	}
	if pageTitle == "" {
		pageTitle = title
	}
	model := tmpl.LayoutModel{
		PageTitle:  pageTitle,
		Favicon:    favicon,
		Version:    t.versionString(),
		User:       user,
		Search:     search,
		Navigation: appNav,
	}

	if model.Favicon == "" {
		model.Favicon = "/public/folder.svg"
	}
	var jsReloadLogic string
	if t.Env == config.Development {
		jsReloadLogic = develop.PageReloadClientJS
	}
	model.PageReloadClientJS = tmpl.PageReloadClientJS(jsReloadLogic)
	return model
}

// --------------------------------------------------------------------------

func queryParam(r *http.Request, name string) string {
	keys, ok := r.URL.Query()[name]
	if !ok || len(keys[0]) < 1 {
		return ""
	}
	return keys[0]
}

func pathParam(r *http.Request, name string) string {
	return chi.URLParam(r, name)
}

func ensureUser(r *http.Request) *security.User {
	user, ok := security.UserFromContext(r.Context())
	if !ok || user == nil {
		panic("the security context user is not available!")
	}
	return user
}

func getIntFromString(val string) int {
	if val == "1" {
		return 1
	}
	return 0
}
