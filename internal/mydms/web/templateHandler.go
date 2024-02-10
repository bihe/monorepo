package web

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"golang.binggl.net/monorepo/internal/common"
	"golang.binggl.net/monorepo/internal/mydms/app/document"
	"golang.binggl.net/monorepo/internal/mydms/web/templates"
	"golang.binggl.net/monorepo/pkg/config"
	"golang.binggl.net/monorepo/pkg/develop"
	"golang.binggl.net/monorepo/pkg/handler"
	tmpl "golang.binggl.net/monorepo/pkg/handler/templates"
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
	DocSvc  document.Service
	Version string
	Build   string
}

const searchURL = "/mydms"
const defaultPageSize = 20
const searchParam = "q"
const skipParam = "skip"

// DisplayDocuments shows the available documents for the given user
func (t *TemplateHandler) DisplayDocuments() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := ensureUser(r)
		var (
			search    string
			docNum    int
			documents document.PagedDocument
		)

		documents, search, _ = t.getDocuments(r)
		docNum = len(documents.Documents)
		tmpl.Layout(
			t.layoutModel("Documents", search, "/public/folder.svg", *user),
			templates.DocumentsStyles(),
			templates.DocumentsNavigation(search),
			templates.DocumentsContent(
				templates.DocumentList(docNum, defaultPageSize, search, documents)),
			searchURL,
		).Render(r.Context(), w)
	}
}

// DisplayDocumentsPartial returns the HTML list code for the documents
func (t *TemplateHandler) DisplayDocumentsPartial() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := ensureUser(r)
		var (
			skip      int
			err       error
			search    string
			docNum    int
			documents document.PagedDocument
		)
		documents, search, skip = t.getDocuments(r)
		if err != nil {
			t.Logger.ErrorRequest(fmt.Sprintf("could not get documents for user '%s'; '%v'", user.Username, err), r)
		}
		docSize := len(documents.Documents)
		docNum = docSize + skip
		if skip == 0 {
			skip = defaultPageSize
		}
		next := defaultPageSize + skip
		if next > documents.TotalEntries {
			next = 0
		}
		templates.DocumentList(docNum, next, search, documents).Render(r.Context(), w)
	}
}

func (t *TemplateHandler) getDocuments(r *http.Request) (document.PagedDocument, string, int) {
	user := ensureUser(r)
	var (
		search string
		skip   int
		err    error
	)

	err = r.ParseForm()
	if err != nil {
		t.Logger.Error(fmt.Sprintf("could not parse provided form data; %v", err))
	}
	search = r.FormValue(searchParam)
	skipParam := r.FormValue(skipParam)

	if skipParam != "" {
		skip, err = strconv.Atoi(skipParam)
		if err != nil {
			t.Logger.Warn(fmt.Sprintf("could not parse skip param: '%s'; %v", skipParam, err))
			skip = 0
		}
	}

	t.Logger.InfoRequest(fmt.Sprintf("display the documents for user: '%s'", user.Username), r)

	documents, err := t.DocSvc.SearchDocuments(search, "", "", time.Time{}, time.Time{}, defaultPageSize, skip)
	if err != nil {
		t.Logger.ErrorRequest(fmt.Sprintf("could not get documents for user '%s'; '%v'", user.Username, err), r)
		documents = document.PagedDocument{
			TotalEntries: 0,
			Documents:    make([]document.Document, 0),
		}
	}
	return documents, search, skip
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
		if a.URL == "/mydms" {
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

func ensureUser(r *http.Request) *security.User {
	user, ok := security.UserFromContext(r.Context())
	if !ok || user == nil {
		panic("the security context user is not available!")
	}
	return user
}

func queryParam(r *http.Request, name string) string {
	keys, ok := r.URL.Query()[name]
	if !ok || len(keys[0]) < 1 {
		return ""
	}
	return keys[0]
}
