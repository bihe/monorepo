package web

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"

	"golang.binggl.net/monorepo/internal/bookmarks/app/bookmarks"
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
	Logger  logging.Logger
	Env     config.Environment
	App     *bookmarks.Application
	Version string
	Build   string
}

// SearchBookmarks performs a search for bookmarks and displays the result using server-side rendering
func (t TemplateHandler) SearchBookmarks() http.HandlerFunc {
	tmpl := template.Must(template.ParseFS(TemplateFS, "templates/_layout.html", "templates/search.html"))
	return func(w http.ResponseWriter, r *http.Request) {
		name := queryParam(r, "q")
		user := ensureUser(r)
		data := t.getPageModel(r)
		model := BookmarksSearchModel{
			PageModel: data,
			Search:    name,
		}

		t.Logger.InfoRequest(fmt.Sprintf("get bookmarks by name: '%s' for user: '%s'", name, user.Username), r)

		bms, err := t.App.GetBookmarksByName(name, *user)
		if err != nil {
			t.Logger.ErrorRequest(fmt.Sprintf("could not get bookmarks for search '%s'; '%v'", name, err), r)
		} else {
			model.Bookmarks = bms
		}
		if err := tmpl.Execute(w, model); err != nil {
			t.Logger.Error("template error", logging.ErrV(err))
		}
	}
}

// GetBookmarksForPath retrieves and renders the bookmarks for a defined path
func (t TemplateHandler) GetBookmarksForPath() http.HandlerFunc {
	tmpl := template.Must(template.ParseFS(TemplateFS, "templates/_layout.html", "templates/bookmarks_path.html"))
	return func(w http.ResponseWriter, r *http.Request) {
		path := queryParam(r, "p")
		user := ensureUser(r)
		data := t.getPageModel(r)
		model := BookmarkPathModel{
			PageModel: data,
			Path:      path,
		}

		t.Logger.InfoRequest(fmt.Sprintf("get bookmarks for path: '%s' for user: '%s'", path, user.Username), r)

		bms, err := t.App.GetBookmarksByPath(path, *user)
		if err != nil {
			t.Logger.ErrorRequest(fmt.Sprintf("could not get bookmarks for path '%s'; '%v'", path, err), r)
		} else {
			model.Bookmarks = bms
		}
		if err := tmpl.Execute(w, model); err != nil {
			t.Logger.Error("template error", logging.ErrV(err))
		}
	}
}

// --------------------------------------------------------------------------
//  Errors
// --------------------------------------------------------------------------

// Show403 displays a page which indicates that the given user has no access to the system
func (t TemplateHandler) Show403() http.HandlerFunc {
	tmpl := template.Must(template.ParseFS(TemplateFS, "templates/403.html"))
	return func(w http.ResponseWriter, r *http.Request) {
		tmpl.Execute(w, t.getPageModel(r))
	}
}

// --------------------------------------------------------------------------

func (t TemplateHandler) getPageModel(r *http.Request) (data PageModel) {
	switch t.Env {
	case config.Development:
		data.Development = true
	}

	user, ok := security.UserFromContext(r.Context())
	if !ok || user == nil {
		t.Logger.Error("could not retrieve user from context; user not logged on")
		return
	}
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

func queryParam(r *http.Request, name string) string {
	keys, ok := r.URL.Query()[name]
	if !ok || len(keys[0]) < 1 {
		return ""
	}
	return keys[0]
}

func ensureUser(r *http.Request) *security.User {
	user, ok := security.UserFromContext(r.Context())
	if !ok || user == nil {
		panic("the sucurity context user is not available!")
	}
	return user
}
