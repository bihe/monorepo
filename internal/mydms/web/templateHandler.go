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
	"golang.binggl.net/monorepo/internal/mydms/html"
	"golang.binggl.net/monorepo/pkg/handler"
	base "golang.binggl.net/monorepo/pkg/handler/html"
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

const defaultPageSize = 20
const searchParam = "q"
const skipParam = "skip"
const searchURL = "mydms"

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

		base.Layout(
			t.pageModel("Documents", search, "/public/mydms.svg", *user),
			html.DocumentsStyles(),
			html.DocumentsNavigation(search),
			html.DocumentsContent(
				html.DocumentList(numDocs, next, documents), search,
			), searchURL,
		).Render(w)
	}
}

// DisplayDocumentsPartial returns the HTML list code for the documents
func (t *TemplateHandler) DisplayDocumentsPartial() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			next      int
			numDocs   int
			documents document.PagedDocument
		)
		documents, _, numDocs, next = t.getDocuments(r)
		html.DocumentList(numDocs, next, documents).Render(w)
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
		validDoc := prepValidDoc(doc)
		html.EditDocumentDialog(validDoc, html.DisplayDocumentDownload(validDoc)).Render(w)
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

		// custom structure required by
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

		json := handler.Json(listItems)
		w.Header().Add("content-type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(json))
	}
}

// DisplayDocumentUploadPartial is used to display the upload elements for documents
func (t *TemplateHandler) DisplayDocumentUploadPartial() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		html.DisplayDocumentDownload(html.Document{}).Render(w)
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
			html.DisplayTempDocumentUpload("", "", errMsg).Render(w)
			return
		}
		defer file.Close()

		initPass := r.Form.Get("doc-initPass")
		newPass := r.Form.Get("doc-newPass")
		encReq := upload.EncryptionRequest{
			InitPassword: initPass,
			Password:     newPass,
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
			html.DisplayTempDocumentUpload("", "", errMsg).Render(w)
			return
		}

		html.DisplayTempDocumentUpload(meta.Filename, tempId, "").Render(w)
	}
}

// we define a JSON structure which is used to trigger actions on the frontend via htmx
type triggerDef struct {
	base.ToastMessage
	Refresh string `json:"refreshDocumentList,omitempty"`
}

// SaveDocument receives the data form the edit form, validates it
// and if the data is valid, persists the document
func (t *TemplateHandler) SaveDocument() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			err        error
			formPrefix string = "doc-"
			rcvDoc     document.Document
			validDoc   html.Document
			validData  bool
		)
		user := ensureUser(r)
		err = r.ParseForm()
		if err != nil {
			t.Logger.ErrorRequest(fmt.Sprintf("could not parse supplied form data; '%v'", err), r)
			t.RenderErr(r, w, fmt.Sprintf("could not parse supplied form data; '%v'", err))
			return
		}

		t.Logger.InfoRequest("save the document data", r)

		rcvDoc.ID = r.FormValue(formPrefix + "id")
		rcvDoc.Title = r.FormValue(formPrefix + "title")
		amount := r.FormValue(formPrefix + "amount")
		if amount != "" {
			// parse the supplied amount value
			f, err := strconv.ParseFloat(amount, 32)
			if err != nil {
				t.Logger.Warn(fmt.Sprintf("could not parse amount '%s'; %v", amount, err))
			} else {
				rcvDoc.Amount = float32(f)
			}
		}
		rcvDoc.InvoiceNumber = r.FormValue(formPrefix + "invoicenumber")
		rcvDoc.Tags = r.Form[formPrefix+"tags[]"]
		rcvDoc.Senders = r.Form[formPrefix+"senders[]"]
		rcvDoc.UploadToken = r.FormValue(formPrefix + "tempID")
		rcvDoc.FileName = r.FormValue(formPrefix + "filename")

		validData = true
		validDoc = prepValidDoc(rcvDoc)
		if validDoc.Title.Val == "" {
			validDoc.Title.Valid = false
			validDoc.Title.Message = "Title is required"
			validData = false
		}
		if len(validDoc.Tags.Val) == 0 {
			validDoc.Tags.Valid = false
			validDoc.Tags.Message = "Tags are required"
			validData = false
		}
		if len(validDoc.Senders.Val) == 0 {
			validDoc.Senders.Valid = false
			validDoc.Senders.Message = "Senders are required"
			validData = false
		}
		if validDoc.UploadToken.Val == "" {
			validDoc.UploadToken.Valid = false
			validDoc.UploadToken.Message = "Document upload is required"
			validData = false
		}

		if !validData {
			// show the same form again!
			t.Logger.ErrorRequest("the supplied data for saving a document is not valid!", r)
			html.EditDocumentDialog(validDoc, html.DisplayDocumentDownload(validDoc)).Render(w)
			return
		}

		savedDoc, err := t.DocSvc.SaveDocument(rcvDoc, *user)
		if err != nil {
			t.Logger.ErrorRequest(fmt.Sprintf("could not save the supplied document; %v", err), r)
			validDoc.Error = fmt.Sprintf("Cannot save the document; '%v'", err)
			html.EditDocumentDialog(validDoc, html.DisplayDocumentDownload(validDoc)).Render(w)
			return
		}

		t.Logger.Info("new document created", logging.LogV("ID", savedDoc.ID))

		// the trigger definitions is used for two actions
		// one) is the toast message to show a saved indicator
		// two) is the notification to reload the document list, because of the changes
		triggerEvent := triggerDef{
			ToastMessage: base.ToastMessage{
				Event: base.ToastMessageContent{
					Type:  base.MsgSuccess,
					Title: "Document saved!",
					Text:  fmt.Sprintf("The document with ID '%s' was saved.", savedDoc.ID),
				},
			},
			Refresh: "now",
		}
		// https://htmx.org/headers/hx-trigger/
		w.Header().Add("HX-Trigger", handler.Json(triggerEvent))
		validDoc.Close = true
		html.EditDocumentDialog(validDoc, html.DisplayDocumentDownload(validDoc)).Render(w)
	}
}

// ShowDeleteConfirmDialog shows a confirm dialog before an item is deleted
func (t *TemplateHandler) ShowDeleteConfirmDialog() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := pathParam(r, "id")
		user := ensureUser(r)
		t.Logger.InfoRequest(fmt.Sprintf("get document by id: '%s' for user: '%s'", id, user.Username), r)
		doc, err := t.DocSvc.GetDocumentByID(id)
		if err != nil {
			t.Logger.ErrorRequest(fmt.Sprintf("could not get document for id '%s'; '%v'", id, err), r)
		}
		base.DialogConfirmDeleteHx(doc.Title, "/mydms/"+doc.ID).Render(w)
	}
}

// DeleteDocument removed the document with the given id
func (t *TemplateHandler) DeleteDocument() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := pathParam(r, "id")
		user := ensureUser(r)
		t.Logger.InfoRequest(fmt.Sprintf("delete document by id: '%s' for user: '%s'", id, user.Username), r)

		err := t.DocSvc.DeleteDocumentByID(id)
		if err != nil {
			t.Logger.ErrorRequest(fmt.Sprintf("could not delete document by id '%s'; '%v'", id, err), r)
		}

		// the trigger definitions is used for two actions
		// one) is the toast message to show a saved indicator
		// two) is the notification to reload the document list, because of the changes
		triggerEvent := triggerDef{
			ToastMessage: base.ToastMessage{
				Event: base.ToastMessageContent{
					Type:  base.MsgSuccess,
					Title: "Document delete!",
					Text:  fmt.Sprintf("The document with ID '%s' was removed.", id),
				},
			},
			Refresh: "now",
		}
		// https://htmx.org/headers/hx-trigger/
		w.Header().Add("HX-Trigger", handler.Json(triggerEvent))
	}
}

// --------------------------------------------------------------------------
//  Internals
// --------------------------------------------------------------------------

func (t *TemplateHandler) versionString() string {
	return fmt.Sprintf("%s-%s", t.Version, t.Build)
}

func (t *TemplateHandler) pageModel(pageTitle, searchStr, favicon string, user security.User) base.LayoutModel {
	return common.CreatePageModel("/"+searchURL, pageTitle, searchStr, favicon, t.versionString(), t.Env, user)
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

func prepValidDoc(doc document.Document) html.Document {
	return html.Document{
		ID: doc.ID,
		Title: html.ValidStr{
			Val:   doc.Title,
			Valid: true,
		},
		Amount: html.ValidFloat{
			Val:   doc.Amount,
			Valid: true,
		},
		FileName: html.ValidStr{
			Val:   doc.FileName,
			Valid: true,
		},
		PreviewLink: html.ValidStr{
			Val:   doc.PreviewLink,
			Valid: true,
		},
		UploadToken: html.ValidStr{
			Val:   doc.UploadToken,
			Valid: true,
		},
		InvoiceNumber: html.ValidStr{
			Val:   doc.InvoiceNumber,
			Valid: true,
		},
		Tags: html.ValidStrSlice{
			Val:   doc.Tags,
			Valid: true,
		},
		Senders: html.ValidStrSlice{
			Val:   doc.Senders,
			Valid: true,
		},
	}
}
