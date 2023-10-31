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
	"golang.binggl.net/monorepo/pkg/develop"
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

const (
	pathTemplate  string = "bookmarks_path.html"
	bookmarkList  string = "bookmarks_list.html"
	confirmDelete string = "dialog_delete_confirm.html"
	editBookmark  string = "dialog_edit_bookmark.html"
)

// NewTemplateHandler performs some internal setup to prepare templates for later use
func NewTemplateHandler(app *bookmarks.Application, logger logging.Logger, environment config.Environment, version, build string) *TemplateHandler {
	functions := template.FuncMap{
		"trailingSlash":   trailingSlash,
		"cssShowWhenTrue": cssShowWhenTrue,
		"valWhenNodeEq":   valWhenNodeEq,
		"valWhenTrue":     valWhenTrue,
	}
	templates := make(map[string]*template.Template)
	templates[pathTemplate] = parseTemplate(pathTemplate, functions, "templates/_layout.html", "templates/toast.html", "templates/bookmarks_path.html", "templates/bookmarks_list.html")
	templates[confirmDelete] = parseTemplate(confirmDelete, functions, "templates/dialog_delete_confirm.html")
	templates[bookmarkList] = parseTemplate(bookmarkList, functions, "templates/toast.html", "templates/bookmarks_list.html")
	templates[editBookmark] = parseTemplate(editBookmark, functions, "templates/dialog_edit_bookmark.html")

	return &TemplateHandler{
		Logger:    logger,
		Env:       environment,
		App:       app,
		Version:   version,
		Build:     build,
		templates: templates,
	}
}

func parseTemplate(name string, functions template.FuncMap, templates ...string) *template.Template {
	template, err := template.New(name).Funcs(functions).ParseFS(TemplateFS, templates...)
	if err != nil {
		panic(err)
	}
	return template
}

// SearchBookmarks performs a search for bookmarks and displays the result using server-side rendering
func (t *TemplateHandler) SearchBookmarks() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := queryParam(r, "q")
		user := ensureUser(r)

		t.Logger.InfoRequest(fmt.Sprintf("get bookmarks by name: '%s' for user: '%s'", name, user.Username), r)

		bms, err := t.App.GetBookmarksByName(name, *user)
		if err != nil {
			t.Logger.ErrorRequest(fmt.Sprintf("could not get bookmarks for search '%s'; '%v'", name, err), r)
		}

		templates.Layout(templates.LayoutModel{
			Title:              "Search Bookmarks!",
			Version:            t.versionString(),
			User:               *user,
			Search:             name,
			PageReloadClientJS: templates.PageReloadClientJS(),
		}, templates.SearchStyles(), templates.SearchNavigation(name), templates.SearchContent(bms)).Render(r.Context(), w)
	}
}

// GetBookmarksForPath retrieves and renders the bookmarks for a defined path
func (t *TemplateHandler) GetBookmarksForPath() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := pathParam(r, "*")
		user := ensureUser(r)
		data := t.getPageModel(r)

		pathHierarchy := make([]BookmarkPathEntry, 1)
		// always start with the root item
		pathHierarchy[0] = BookmarkPathEntry{
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
			pathHierarchy = append(pathHierarchy, BookmarkPathEntry{
				UrlPath:     lastPath + "/" + p,
				DisplayName: p,
			})
			lastPath = lastPath + "/" + p
		}
		// highlight the last item
		pathHierarchy[len(pathHierarchy)-1].LastItem = true

		model := BookmarkPathModel{
			PageModel: data,
			Path:      pathHierarchy,
		}

		t.Logger.InfoRequest(fmt.Sprintf("get bookmarks for path: '%s' for user: '%s'", path, user.Username), r)

		bms, err := t.App.GetBookmarksByPath(path, *user)
		if err != nil {
			t.Logger.ErrorRequest(fmt.Sprintf("could not get bookmarks for path '%s'; '%v'", path, err), r)
		} else {
			model.Bookmarks = bms
		}
		callTemplate[BookmarkPathModel](t, pathTemplate, model, w)
	}
}

// DeleteConfirm shows a confirm dialog before an item is deleted
func (t *TemplateHandler) DeleteConfirm() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := pathParam(r, "id")
		user := ensureUser(r)
		t.Logger.InfoRequest(fmt.Sprintf("get bookmarks by id: '%s' for user: '%s'", id, user.Username), r)
		model := ConfirmDeleteModel{}

		bm, err := t.App.GetBookmarkByID(id, *user)
		if err != nil {
			t.Logger.ErrorRequest(fmt.Sprintf("could not get bookmarks for id '%s'; '%v'", id, err), r)
		} else {
			model.Item = bm.DisplayName
			model.ID = bm.ID
		}
		callTemplate[ConfirmDeleteModel](t, confirmDelete, model, w)
	}
}

// DeleteBookmark deletes the given bookmark and replaces the bookmark-list with a new list without the deleted bookmark
func (t *TemplateHandler) DeleteBookmark() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := pathParam(r, "id")
		user := ensureUser(r)
		t.Logger.InfoRequest(fmt.Sprintf("get bookmark by id: '%s' for user: '%s'", id, user.Username), r)

		model := BookmarkResultModel{}

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
			} else {
				model.Bookmarks = bms
			}
			model.Message = showErrorToast("Bookmark delete error", fmt.Sprintf("Error: '%s'", err))
			callTemplate[BookmarkResultModel](t, bookmarkList, model, w)
			return
		}

		bms, err := t.App.GetBookmarksByPath(bm.Path, *user)
		if err != nil {
			t.Logger.ErrorRequest(fmt.Sprintf("could not get bookmarks for path '%s'; '%v'", bm.Path, err), r)
		} else {
			model.Bookmarks = bms
		}
		model.Message = showSuccessToast("Bookmark deleted", fmt.Sprintf("The bookmark '%s' was deleted.", bm.DisplayName))
		callTemplate[BookmarkResultModel](t, bookmarkList, model, w)
	}
}

// EditBookmarkDialog shows a dialog to edit a bookmark
func (t *TemplateHandler) EditBookmarkDialog() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := pathParam(r, "id")
		user := ensureUser(r)
		t.Logger.InfoRequest(fmt.Sprintf("get bookmarks by id: '%s' for user: '%s'", id, user.Username), r)
		model := EditBookmarkModel{}

		bm, err := t.App.GetBookmarkByID(id, *user)
		if err != nil {
			t.Logger.ErrorRequest(fmt.Sprintf("could not get bookmarks for id '%s'; '%v'", id, err), r)
		} else {
			model.Bookmark = *bm
		}
		callTemplate[EditBookmarkModel](t, editBookmark, model, w)
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

func (t *TemplateHandler) getPageModel(r *http.Request) (data PageModel) {
	switch t.Env {
	case config.Development:
		data.Development = true
	}

	user, ok := security.UserFromContext(r.Context())
	if !ok || user == nil {
		t.Logger.Error("could not retrieve user from context; user not logged on")
		return
	}

	data.PageReloadClientJS = template.HTML(develop.PageReloadClientJS)
	data.Authenticated = true
	data.VersionString = fmt.Sprintf("%s-%s", t.Version, t.Build)
	data.User = UserModel{
		Username:    user.Username,
		Email:       user.Email,
		UserID:      user.UserID,
		DisplayName: user.DisplayName,
		ProfileURL:  user.ProfileURL,
		Token:       user.Token,
	}
	return data
}

func callTemplate[M any](t *TemplateHandler, tmpl string, model M, w http.ResponseWriter) {
	if err := t.templates[tmpl].Execute(w, model); err != nil {
		t.Logger.Error("template error", logging.ErrV(err))
		panic(err)
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
