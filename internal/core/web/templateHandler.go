package web

import (
	"embed"
	"html/template"
	"net/http"

	"golang.binggl.net/monorepo/pkg/config"
	"golang.binggl.net/monorepo/pkg/logging"
	"golang.binggl.net/monorepo/pkg/security"
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
	Logger     logging.Logger
	Env        config.Environment
	SiteApiURL string
}

// CreateTemplateHandler provides a new instance
func CreateTemplateHandler(log logging.Logger, env config.Environment, siteApi string) TemplateHandler {
	return TemplateHandler{
		Logger:     log,
		Env:        env,
		SiteApiURL: siteApi,
	}
}

func (t TemplateHandler) Index() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := t.getPageModel(r)
		parsedTemplates.ExecuteTemplate(w, "index.html", data)
	}
}

// --------------------------------------------------------------------------
//  Sites
// --------------------------------------------------------------------------

func (t TemplateHandler) GetSites() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := t.getPageModel(r)
		parsedTemplates.ExecuteTemplate(w, "sites_index.html", data)
	}
}

// --------------------------------------------------------------------------
//  Errors
// --------------------------------------------------------------------------

// Show403 displays a page which indicates that the given user has no access to the system
func (t TemplateHandler) Show403() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		parsedTemplates.ExecuteTemplate(w, "403.html", t.getPageModel(r))
	}
}

// --------------------------------------------------------------------------

func (t TemplateHandler) getPageModel(r *http.Request) (data PageModel) {
	switch t.Env {
	case config.Development:
		data.Development = true
	case config.Integration:
		data.Integration = true
	}
	data.CurrPage = r.URL.Path

	user, ok := security.UserFromContext(r.Context())
	if !ok || user == nil {
		t.Logger.Error("could not retrieve user from context; user not logged on")
		return
	}

	data.Authenticated = true
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
