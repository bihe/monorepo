package handler

import (
	"net/http"

	"golang.binggl.net/monorepo/pkg/config"
	"golang.binggl.net/monorepo/pkg/handler/templates"
	"golang.binggl.net/monorepo/pkg/logging"
)

// TemplateHanlder provides some basics for HTML template based handlers
type TemplateHandler struct {
	Logger    logging.Logger
	Env       config.Environment
	BasePath  string
	StartPage string
}

// Show403 displays a page which indicates that the given user has no access to the system
func (t *TemplateHandler) Show403() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		templates.ErrorPageLayout(t.BasePath, templates.Error403(t.Env)).Render(r.Context(), w)
	}
}

// Show404 is used for the http not-found error
func (t *TemplateHandler) Show404() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		templates.ErrorPageLayout(t.BasePath, templates.Error404(t.Env)).Render(r.Context(), w)
	}
}

// RenderErr uses the application template to render the error-page
func (t *TemplateHandler) RenderErr(r *http.Request, w http.ResponseWriter, message string) {
	templates.ErrorPageLayout(t.BasePath, templates.ErrorApplication(
		t.StartPage,
		t.Env,
		r,
		message)).Render(r.Context(), w)
}
