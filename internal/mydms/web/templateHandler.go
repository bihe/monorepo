package web

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"golang.binggl.net/monorepo/internal/common"
	"golang.binggl.net/monorepo/internal/mydms/app/document"
	"golang.binggl.net/monorepo/internal/mydms/app/upload"
	"golang.binggl.net/monorepo/internal/mydms/web/templates"
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
	DocSvc        document.Service
	UploadSvc     upload.Service
	Version       string
	Build         string
	MaxUploadSize int64
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
			numDocs   int
			next      int
			documents document.PagedDocument
		)

		documents, search, numDocs, next = t.getDocuments(r)
		tmpl.Layout(
			t.layoutModel("Documents", search, "/public/folder.svg", *user),
			templates.DocumentsStyles(),
			templates.DocumentsNavigation(search),
			templates.DocumentsContent(
				templates.DocumentList(numDocs, next, search, documents)),
			searchURL,
		).Render(r.Context(), w)
	}
}

// DisplayDocumentsPartial returns the HTML list code for the documents
func (t *TemplateHandler) DisplayDocumentsPartial() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			next      int
			search    string
			numDocs   int
			documents document.PagedDocument
		)
		documents, search, numDocs, next = t.getDocuments(r)
		templates.DocumentList(numDocs, next, search, documents).Render(r.Context(), w)
	}
}

func (t *TemplateHandler) getDocuments(r *http.Request) (documents document.PagedDocument, search string, numDocs, next int) {
	user := ensureUser(r)
	var (
		skip int
		err  error
	)

	err = r.ParseForm()
	if err == nil {
		search = r.FormValue(searchParam)
		skipParam := r.FormValue(skipParam)

		if skipParam != "" {
			skip, err = strconv.Atoi(skipParam)
			if err != nil {
				t.Logger.Warn(fmt.Sprintf("could not parse skip param: '%s'; %v", skipParam, err))
				skip = 0
			}
			if skip < 0 {
				skip = 0
			}
		}
	} else {
		t.Logger.Error(fmt.Sprintf("could not parse provided form data; %v", err))
	}

	t.Logger.InfoRequest(fmt.Sprintf("display the documents for user: '%s'", user.Username), r)

	documents, err = t.DocSvc.SearchDocuments(search, "", "", time.Time{}, time.Time{}, defaultPageSize, skip)
	if err != nil {
		t.Logger.ErrorRequest(fmt.Sprintf("could not get documents for user '%s'; '%v'", user.Username, err), r)
		documents = document.PagedDocument{
			TotalEntries: 0,
			Documents:    make([]document.Document, 0),
		}
	}

	docSize := len(documents.Documents)
	numDocs = docSize + skip
	next = defaultPageSize + skip
	if next > documents.TotalEntries {
		next = 0
	}

	return documents, search, numDocs, next
}

// ShowEditDocumentDialog is used to display the document edit dialog
// either to create a new document or edit an existing one
func (t *TemplateHandler) ShowEditDocumentDialog() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			id  string
			doc document.Document
			err error
		)
		user := ensureUser(r)
		id = pathParam(r, "id")
		switch id {
		case "NEW":
			t.Logger.Debug("open dialog to create a new document")
		case "":
			t.Logger.ErrorRequest("empty param supplied for id", r)
		default:
			t.Logger.Debug(fmt.Sprintf("fetch document by id '%s' for user '%s'", id, *user), logging.LogV("id", id))
			doc, err = t.DocSvc.GetDocumentByID(id)
			if err != nil {
				t.Logger.ErrorRequest(fmt.Sprintf("could not get document for id '%s'; %v", id, err), r)
			}
		}
		templates.EditDocumentDialog(doc, templates.DisplayDocumentDownload(doc)).Render(r.Context(), w)
	}
}

// SearchListItems returns a JSON structure used by the frontend to search for tags/senders
// the structure is defined by the 3rd party frontend library used: https://github.com/lekoala/bootstrap5-tags
func (t *TemplateHandler) SearchListItems() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		listType := pathParam(r, "type")
		if listType == "" {
			t.Logger.ErrorRequest("empty type supplied unable to search list", r)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		q := queryParam(r, "query")

		var st document.SearchType
		switch listType {
		case "tags":
			st = document.TAGS
		case "senders":
			st = document.SENDERS
		}

		t.Logger.Debug(fmt.Sprintf("search list '%s' for item '%s'", listType, q))

		items, err := t.DocSvc.SearchList(q, st)
		if err != nil {
			t.Logger.ErrorRequest(fmt.Sprintf("could search items in list; %v", err), r)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// custom structure requuired by
		// https://github.com/lekoala/bootstrap5-tags
		type listItem struct {
			Value string `json:"value,omitempty"`
			Label string `json:"label,omitempty"`
		}
		listItems := make([]listItem, 0)
		for _, item := range items {
			listItems = append(listItems, listItem{
				Value: item,
				Label: item,
			})
		}

		json := tmpl.Json(listItems)
		w.Header().Add("content-type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(json))
	}
}

// DisplayDocumentUploadPartial is used to display the upload elements for documents
func (t *TemplateHandler) DisplayDocumentUploadPartial() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		templates.DisplayDocumentDownload(document.Document{}).Render(r.Context(), w)
	}
}

// UploadDocument uses the upload service to create a temp file which is later used
// when the overall document is saved
func (t *TemplateHandler) UploadDocument() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.ParseMultipartForm(t.MaxUploadSize)

		file, meta, err := r.FormFile("doc-fileupload")
		if err != nil {
			errMsg := strings.ReplaceAll(err.Error(), "\"", "'")
			t.Logger.ErrorRequest(fmt.Sprintf("could not upload document; '%v'", err), r)
			templates.DisplayTempDocumentUpload("", "", errMsg).Render(r.Context(), w)
			return
		}
		defer file.Close()

		initPass := r.Form.Get("doc-initPass")
		newPass := r.Form.Get("doc-newPass")
		var encReq upload.EncryptionRequest
		if initPass != "" && newPass != "" {
			// passwords are provided so we need to call the encryption service
			encReq = upload.EncryptionRequest{
				InitPassword: initPass,
				Password:     newPass,
			}
		}
		cType := meta.Header.Get("Content-Type")
		tempId, err := t.UploadSvc.Save(upload.File{
			File:     file,
			Name:     meta.Filename,
			Size:     meta.Size,
			MimeType: cType,
			Enc:      encReq,
		})
		if err != nil {
			errMsg := strings.ReplaceAll(err.Error(), "\"", "'")
			t.Logger.ErrorRequest(fmt.Sprintf("could not save document; '%v'", err), r)
			templates.DisplayTempDocumentUpload("", "", errMsg).Render(r.Context(), w)
			return
		}

		templates.DisplayTempDocumentUpload(meta.Filename, tempId, "").Render(r.Context(), w)
	}
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

func pathParam(r *http.Request, name string) string {
	return chi.URLParam(r, name)
}
