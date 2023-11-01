package web

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"golang.binggl.net/monorepo/internal/bookmarks/app/bookmarks"
	"golang.binggl.net/monorepo/internal/bookmarks/web/templates"
	"golang.binggl.net/monorepo/pkg/config"
	"golang.binggl.net/monorepo/pkg/logging"
	"golang.binggl.net/monorepo/pkg/security"
)

//go:embed templates/*
var TemplateFS embed.FS

// TemplateHandler takes care of providing HTML templates.
// This is the new approach with a template + htmx based UI to replace the angular frontend
// and have a more go-oriented approach towards UI and user-interaction. This reduces the
// cognitive load because less technology mix is needed. Via template + htmx approach only
// a limited amount of javascript is needed to achieve the frontend.
// As additional benefit the build should be faster, because the nodejs build can be removed
type TemplateHandler struct {
	Logger    logging.Logger
	Env       config.Environment
	App       *bookmarks.Application
	Version   string
	Build     string
	templates map[string]*template.Template
}

// NewTemplateHandler performs some internal setup to prepare templates for later use
func NewTemplateHandler(app *bookmarks.Application, logger logging.Logger, environment config.Environment, version, build string) *TemplateHandler {
	templates := make(map[string]*template.Template)

	return &TemplateHandler{
		Logger:    logger,
		Env:       environment,
		App:       app,
		Version:   version,
		Build:     build,
		templates: templates,
	}
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

		templates.Layout(
			t.layoutModel("Search Bookmarks!", search, *user),
			templates.SearchStyles(),
			templates.SearchNavigation(search),
			templates.SearchContent(bms),
		).Render(r.Context(), w)
	}
}

// GetBookmarksForPath retrieves and renders the bookmarks for a defined path
func (t *TemplateHandler) GetBookmarksForPath() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := pathParam(r, "*")
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
		}

		templates.Layout(
			t.layoutModel("Bookmarks!", "", *user),
			templates.BookmarksByPathStyles(),
			templates.BookmarksByPathNavigation(pathHierarchy),
			templates.BookmarksByPathContent(templates.BookmarkList(bms, templates.Toast(templates.MessageModel{}))),
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

			templates.BookmarkList(
				bms,
				templates.Toast(
					templates.ShowErrorToast("Bookmark delete error", fmt.Sprintf("Error: '%s'", err)),
				),
			).Render(r.Context(), w)
			return
		}

		bms, err := t.App.GetBookmarksByPath(bm.Path, *user)
		if err != nil {
			t.Logger.ErrorRequest(fmt.Sprintf("could not get bookmarks for path '%s'; '%v'", bm.Path, err), r)
		}

		templates.BookmarkList(
			bms,
			templates.Toast(
				templates.ShowSuccessToast("Bookmark deleted", fmt.Sprintf("The bookmark '%s' was deleted.", bm.DisplayName)),
			),
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
			bm.Favicon = templates.ValidatorInput{Valid: true}
			bm.Type = bookmarks.Node
		} else {
			// fetch an existing bookmark
			b, err = t.App.GetBookmarkByID(id, *user)
			if err != nil {
				t.Logger.ErrorRequest(fmt.Sprintf("could not get bookmarks for id '%s'; '%v'", id, err), r)
			}
			bm.ID = templates.ValidatorInput{Val: b.ID, Valid: true}
			bm.Path = templates.ValidatorInput{Val: b.Path, Valid: true}
			bm.DisplayName = templates.ValidatorInput{Val: b.DisplayName, Valid: true}
			bm.URL = templates.ValidatorInput{Val: b.URL, Valid: true}
			bm.Type = b.Type
			bm.Favicon = templates.ValidatorInput{Val: b.Favicon, Valid: true}
			bm.InvertFaviconColor = (b.InvertFaviconColor == 1)

			paths, err = t.App.GetAllPaths(*user)
			if err != nil {
				t.Logger.ErrorRequest(fmt.Sprintf("could not get all paths for bookmarks; '%v'", err), r)
			}
		}

		templates.EditBookmarks(bm, paths).Render(r.Context(), w)
	}
}

// SaveBookmark receives the values from the edit form, validates and persists the data
func (t *TemplateHandler) SaveBookmark() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			recv       bookmarks.Bookmark
			bm         templates.Bookmark
			paths      []string
			err        error
			formPrefix string = "bookmark_"
		)

		err = r.ParseForm()
		if err != nil {
			t.Logger.ErrorRequest(fmt.Sprintf("could not parse supplied form data; '%v'", err), r)
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
		if r.FormValue(formPrefix+"Type") == "Folder" {
			recv.Type = bookmarks.Folder
		}
		if getIntFromString(r.FormValue(formPrefix+"UseCustomFavicon")) == 1 {
			customFavicon = true
			recv.Favicon = r.FormValue(formPrefix + "CustomFavicon")
		}

		// validation
		bm.ID = templates.ValidatorInput{Val: recv.ID, Valid: true}
		bm.DisplayName = templates.ValidatorInput{Val: recv.DisplayName, Valid: true}
		if recv.DisplayName == "" {
			bm.DisplayName.Valid = false
			bm.DisplayName.Message = "missing value!"
		}
		bm.Path = templates.ValidatorInput{Val: recv.Path, Valid: true}
		if recv.Path == "" {
			bm.Path.Valid = false
			bm.Path.Message = "missing value!"
		}
		bm.Type = recv.Type
		if bm.Type == bookmarks.Node {
			bm.URL = templates.ValidatorInput{Val: recv.URL, Valid: true}
			if recv.URL == "" {
				bm.URL.Valid = false
				bm.URL.Message = "missing value!"
			}
		}
		if customFavicon {
			bm.UseCustomFavicon = customFavicon
			bm.Favicon = templates.ValidatorInput{Val: recv.Favicon, Valid: true}
			if recv.Favicon == "" {
				bm.Favicon.Valid = false
				bm.Favicon.Message = "missing value!"
			}
		}
		bm.InvertFaviconColor = (recv.InvertFaviconColor == 1)

		templates.EditBookmarks(bm, paths).Render(r.Context(), w)

		// hint: Update with htmx Events: https://htmx.org/examples/update-other-content/#events
	}
}

// --------------------------------------------------------------------------
//  Errors
// --------------------------------------------------------------------------

// Show403 displays a page which indicates that the given user has no access to the system
func (t *TemplateHandler) Show403() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		templates.Page403(t.Env).Render(r.Context(), w)
	}
}

// --------------------------------------------------------------------------
//  Internals
// --------------------------------------------------------------------------

func (t *TemplateHandler) versionString() string {
	return fmt.Sprintf("%s-%s", t.Version, t.Build)
}

func (t *TemplateHandler) layoutModel(title, search string, user security.User) templates.LayoutModel {
	return templates.LayoutModel{
		Title:              title,
		Version:            t.versionString(),
		User:               user,
		Search:             search,
		PageReloadClientJS: templates.PageReloadClientJS(),
	}
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
		panic("the sucurity context user is not available!")
	}
	return user
}

func getIntFromString(val string) int {
	if val == "1" {
		return 1
	}
	return 0
}
