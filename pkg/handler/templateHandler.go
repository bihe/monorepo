package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"golang.binggl.net/monorepo/pkg/config"
	"golang.binggl.net/monorepo/pkg/handler/html"
	"golang.binggl.net/monorepo/pkg/logging"
)

// TemplateHandler provides some basics for HTML template based handlers
type TemplateHandler struct {
	Logger    logging.Logger
	Env       config.Environment
	Commit    string
	BasePath  string
	StartPage string
}

// Show403 displays a page which indicates that the given user has no access to the system
func (t *TemplateHandler) Show403() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		html.ErrorPage403(t.BasePath, t.Env, t.Commit).Render(w)
	}
}

// Show404 is used for the http not-found error
func (t *TemplateHandler) Show404() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		html.ErrorPage404(t.BasePath, t.Env, t.Commit).Render(w)
	}
}

// RenderErr uses the application template to render the error-page
func (t *TemplateHandler) RenderErr(r *http.Request, w http.ResponseWriter, message string) {
	html.ErrorApplication(t.BasePath, t.Env, t.Commit, t.StartPage, r, message).Render(w)
}

// Json serialized the given data
func Json[T any](data T) string {
	payload, err := json.Marshal(data)
	if err != nil {
		panic(fmt.Sprintf("could not marshall data; %v", err))
	}
	return string(payload)
}

// JsonIndent serialized the given data and used pretty formatting
func JsonIndent[T any](data T) string {
	payload, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		panic(fmt.Sprintf("could not marshall data; %v", err))
	}
	return string(payload)
}
