package web

import (
	"embed"
	"html/template"
	"net/http"

	"golang.binggl.net/monorepo/pkg/config"
	"golang.binggl.net/monorepo/pkg/logging"
)

//go:embed templates/*
var templateFS embed.FS

var parsedTemplates *template.Template

func init() {
	parsedTemplates = template.Must(template.ParseFS(templateFS, "templates/*.html"))
}

// TemplateHandler takes care of providing HTML templates.
// This is the new approach with a template + htmx based UI to replace the angular frontend
// and have a more go-oriented approach towards UI and user-interaction. This reduces the
// cognitive load because less technology mix is needed. Via template + htmx approach only
// a limited amount of javascript is needed to achieve the frontend.
// As additional benefit the build should be faster, because the nodejs build can be removed
type TemplateHandler struct {
	Logger logging.Logger
	Env    config.Environment
}

type pageContext struct {
	Development bool
	Integration bool
	User        userData
}

type userData struct {
	Username    string
	Email       string
	UserID      string
	DisplayName string
	ProfileURL  string
}

func (t TemplateHandler) Index() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := ensureUser(r)
		t.Logger.Debug("retrieve the site with name for user", logging.LogV("username", user.DisplayName))
		data := pageContext{}
		data.User = userData{
			Username:    user.Username,
			Email:       user.Email,
			UserID:      user.UserID,
			DisplayName: user.DisplayName,
			ProfileURL:  user.ProfileURL,
		}
		parsedTemplates.ExecuteTemplate(w, "index.html", data)
	}
}

// Show403 displays a page which indicates that the given user has no access to the system
func (t TemplateHandler) Show403() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var data pageContext
		switch t.Env {
		case config.Development:
			data.Development = true
		case config.Integration:
			data.Integration = true
		}
		parsedTemplates.ExecuteTemplate(w, "403.html", data)
	}
}
