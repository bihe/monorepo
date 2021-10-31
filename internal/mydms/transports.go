// Package mydms mydms-API
//
// The API of the mydms-application.
//
// Terms Of Service:
//
//     Schemes: https
//     Host: mydms.binggl.net
//     BasePath: /api/v1
//     Version: 3.0.0
//     License: Apache 2.0 https://opensource.org/licenses/Apache-2.0
//
//     Consumes:
//     - application/json
//
//     Produces:
//     - application/json
//
// swagger:meta
//go:generate ../../tools/swagger_linux_amd64 generate spec -o ./assets/swagger/swagger.json -m -w ./ -x handler
package mydms

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	stdlog "log"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"golang.binggl.net/monorepo/internal/mydms/app/appinfo"
	"golang.binggl.net/monorepo/internal/mydms/app/document"
	"golang.binggl.net/monorepo/internal/mydms/app/filestore"
	"golang.binggl.net/monorepo/internal/mydms/app/shared"
	"golang.binggl.net/monorepo/pkg/config"
	"golang.binggl.net/monorepo/pkg/logging"
	"golang.binggl.net/monorepo/pkg/security"
	"golang.binggl.net/monorepo/pkg/server"

	"github.com/go-kit/kit/transport"
	httptransport "github.com/go-kit/kit/transport/http"
	pkgerr "golang.binggl.net/monorepo/pkg/errors"
)

// HTTPHandlerOptions containts configuration settings relevant for the HTTP handler implementation
type HTTPHandlerOptions struct {
	BasePath     string
	ErrorPath    string
	JWTConfig    config.Security
	CookieConfig config.ApplicationCookies
	CorsConfig   config.CorsSettings
	AssetConfig  config.AssetSettings
}

// MakeHTTPHandler mounts all service endpoints into an http.Handler.
func MakeHTTPHandler(e Endpoints, logger logging.Logger, opts HTTPHandlerOptions) http.Handler {
	r := server.SetupBasicRouter(opts.BasePath, opts.CookieConfig, opts.CorsConfig, opts.AssetConfig, logger)
	apiRouter := server.SetupSecureAPIRouter(opts.ErrorPath, opts.JWTConfig, opts.CookieConfig, logger)
	options := []httptransport.ServerOption{
		httptransport.ServerErrorHandler(transport.ErrorHandlerFunc(func(ctx context.Context, err error) {
			logger.Error("httptransport error", logging.ErrV(err))
		})),
		httptransport.ServerErrorEncoder(encodeError),
	}

	// GetAppInfo retrieves meta information about the application
	// swagger:operation GET /api/v1/apinfo appinfo GetAppInfo
	//
	// get metadata of the aplication
	//
	// ---
	// produces:
	// - application/json
	// responses:
	//   '200':
	//     description: GetAppInfoResponse
	//     schema:
	//       "$ref": "#/definitions/AppInfo"
	//   '401':
	//     description: ProblemDetail
	//     schema:
	//       "$ref": "#/definitions/ProblemDetail"
	//   '403':
	//     description: ProblemDetail
	//     schema:
	//       "$ref": "#/definitions/ProblemDetail"
	apiRouter.Mount("/appinfo", httptransport.NewServer(
		e.GetAppInfoEndpoint,
		decodeGetAppInfoRequest,
		encodeAppInfoResponse,
		options...,
	))

	// GetDocumentByID returns a document by the specified ID
	// swagger:operation GET /api/v1/documents/{id} documents GetDocumentByID
	//
	// use the supplied id to retrieve the document
	//
	// ---
	// produces:
	// - application/json
	// parameters:
	// - name: id
	//   type: string
	//   in: path
	//   required: true
	//
	// responses:
	//   '200':
	//     description: GetDocumentByIDResponse
	//     schema:
	//       "$ref": "#/definitions/Document"
	//   '400':
	//     description: ProblemDetail
	//     schema:
	//       "$ref": "#/definitions/ProblemDetail"
	//   '401':
	//     description: ProblemDetail
	//     schema:
	//       "$ref": "#/definitions/ProblemDetail"
	//   '403':
	//     description: ProblemDetail
	//     schema:
	//       "$ref": "#/definitions/ProblemDetail"
	getDocByID := httptransport.NewServer(
		e.GetDocumentByIDEndpoint,
		decodeGetDocumentByIDRequest,
		encodeGetDocumentByIDResponse,
		options...,
	)
	apiRouter.Mount("/documents/{id}", getDocByID)
	apiRouter.Mount("/documents/", getDocByID)

	// DeleteDocumentByID deletes a document by the specified ID
	// swagger:operation DELETE /api/v1/documents/{id} documents DeleteDocumentByID
	//
	// use the supplied id to delete the document from the store
	//
	// ---
	// produces:
	// - application/json
	// parameters:
	// - name: id
	//   type: string
	//   in: path
	//   required: true
	//
	// responses:
	//   '200':
	//     description: DeleteDocumentByResponse
	//     schema:
	//       "$ref": "#/definitions/IDResult"
	//   '400':
	//     description: ProblemDetail
	//     schema:
	//       "$ref": "#/definitions/ProblemDetail"
	//   '401':
	//     description: ProblemDetail
	//     schema:
	//       "$ref": "#/definitions/ProblemDetail"
	//   '403':
	//     description: ProblemDetail
	//     schema:
	//       "$ref": "#/definitions/ProblemDetail"
	deleteDocumentByID := httptransport.NewServer(
		e.DeleteDocumentByIDEndpoint,
		decodeDeleteDocumentByIDRequest,
		encodeDeleteDocumentByIDResponse,
		options...,
	)
	apiRouter.Method("DELETE", "/documents/{id}", deleteDocumentByID)
	apiRouter.Method("DELETE", "/documents/", deleteDocumentByID)

	// SearchList search for tags/senders
	// swagger:operation GET /api/v1/documents/{type}/search documents SearchList
	//
	// search either by tags or senders with the supplied search term
	//
	// ---
	// produces:
	// - application/json
	// parameters:
	// - name: type
	//   type: string
	//   in: path
	//   required: true
	//   description: search type tags || senders
	// - name: name
	//   type: string
	//   in: query
	//   required: false
	//   description: search term
	//
	// responses:
	//   '200':
	//     description: SearchListResponse
	//     schema:
	//       "$ref": "#/definitions/EntriesResult"
	//   '400':
	//     description: ProblemDetail
	//     schema:
	//       "$ref": "#/definitions/ProblemDetail"
	//   '401':
	//     description: ProblemDetail
	//     schema:
	//       "$ref": "#/definitions/ProblemDetail"
	//   '403':
	//     description: ProblemDetail
	//     schema:
	//       "$ref": "#/definitions/ProblemDetail"
	apiRouter.Mount("/documents/{type}/search", httptransport.NewServer(
		e.SearchListEndpoint,
		decodeSearchListRequest,
		encodeSearchListResponse,
		options...,
	))

	// SearchDocuments search for documents
	// swagger:operation GET /api/v1/documents/search documents SearchDocuments
	//
	// use filters to search for docments. the result is a paged set
	//
	// ---
	// produces:
	// - application/json
	// parameters:
	// - name: title
	//   type: string
	//   in: query
	//   description: title search
	// - name: tag
	//   type: string
	//   in: query
	//   description: tag search
	// - name: sender
	//   type: string
	//   in: query
	//   description: sender search
	// - name: from
	//   type: string
	//   in: query
	//   description: start date
	// - name: to
	//   type: string
	//   in: query
	//   description: end date
	// - name: limit
	//   type: integer
	//   in: query
	//   description: limit max results
	// - name: skip
	//   type: integer
	//   in: query
	//   description: skip N results (0-based)
	//
	// responses:
	//   '200':
	//     description: SearchDocumentsResponse
	//     schema:
	//       "$ref": "#/definitions/PagedDcoument"
	//   '400':
	//     description: ProblemDetail
	//     schema:
	//       "$ref": "#/definitions/ProblemDetail"
	//   '401':
	//     description: ProblemDetail
	//     schema:
	//       "$ref": "#/definitions/ProblemDetail"
	//   '403':
	//     description: ProblemDetail
	//     schema:
	//       "$ref": "#/definitions/ProblemDetail"
	apiRouter.Mount("/documents/search", httptransport.NewServer(
		e.SearchDocumentsEndpoint,
		decodeSearchDocumentsRequest,
		encodeSearchDocumentsResponse,
		options...,
	))

	// Create a new bookmark
	// swagger:operation POST /api/v1/documents documents SaveDocument
	//
	// create a document
	//
	// use the supplied document payload and store the data
	//
	// ---
	// consumes:
	// - application/json
	// produces:
	// - application/json
	// responses:
	//   '200':
	//     description: SaveDocumentResponse
	//     schema:
	//       "$ref": "#/definitions/IDResult"
	//   '400':
	//     description: ProblemDetail
	//     schema:
	//       "$ref": "#/definitions/ProblemDetail"
	//   '401':
	//     description: ProblemDetail
	//     schema:
	//       "$ref": "#/definitions/ProblemDetail"
	//   '403':
	//     description: ProblemDetail
	//     schema:
	//       "$ref": "#/definitions/ProblemDetail"
	apiRouter.Method("POST", "/documents", httptransport.NewServer(
		e.SaveDocumentEndpoint,
		decodeSaveDocumentRequest,
		encodeSaveDocumentResponse,
		options...,
	))

	// GetFile retrieves the file by the given path
	// swagger:operation GET /api/v1/file filestore GetFile
	//
	// get file by path
	//
	// returns the binary payload of the file
	//
	// ---
	// produces:
	// - application/octet-stream
	// - application/json
	// parameters:
	// - type: string
	//   name: path
	//   in: query
	//
	// responses:
	//   '200':
	//     description: Binary file payload
	//
	//   '400':
	//     description: ProblemDetail
	//     schema:
	//       "$ref": "#/definitions/ProblemDetail"
	//   '401':
	//     description: ProblemDetail
	//     schema:
	//       "$ref": "#/definitions/ProblemDetail"
	//   '403':
	//     description: ProblemDetail
	//     schema:
	//       "$ref": "#/definitions/ProblemDetail"
	apiRouter.Mount("/file", httptransport.NewServer(
		e.GetFileEndpoint,
		decodeGetFileRequest,
		encodeFilePathResponse,
		options...,
	))

	// we are mounting all APIs under /api/v1 path
	r.Mount("/api/v1", apiRouter)

	return r
}

// --------------------------------------------------------------------------
// encode / decode requests and responses
// --------------------------------------------------------------------------

// AppInfoRequest
// --------------------------------------------------------------------------
func decodeGetAppInfoRequest(_ context.Context, r *http.Request) (request interface{}, err error) {
	user := ensureUser(r)
	return appinfo.GetAppInfoRequest{User: user}, nil
}

func encodeAppInfoResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	e, ok := response.(appinfo.GetAppInfoResponse)
	if ok && e.Failed() != nil {
		encodeError(ctx, e.Failed(), w)
		return nil
	}
	return respondJSON(w, appinfo.AppInfo{
		UserInfo:    e.AppInfo.UserInfo,
		VersionInfo: e.AppInfo.VersionInfo,
	})
}

// DocumentByIDRequest
// --------------------------------------------------------------------------
func decodeGetDocumentByIDRequest(_ context.Context, r *http.Request) (request interface{}, err error) {
	id := chi.URLParam(r, "id")
	if id == "" {
		return nil, shared.ErrValidation("missing 'id' path-parameter")
	}
	return document.GetDocumentByIDRequest{
		ID: chi.URLParam(r, "id"),
	}, nil
}

func encodeGetDocumentByIDResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	e, ok := response.(document.GetDocumentByIDResponse)
	if ok && e.Failed() != nil {
		encodeError(ctx, e.Failed(), w)
		return nil
	}
	return respondJSON(w, e.Document)
}

// GetFileRequest
// --------------------------------------------------------------------------
func decodeGetFileRequest(_ context.Context, r *http.Request) (request interface{}, err error) {
	path := r.URL.Query().Get("path")
	if path == "" {
		return nil, shared.ErrValidation("missing 'path' query-parameter")
	}
	decodedPath, err := base64.StdEncoding.DecodeString(path)
	if err != nil {
		return nil, shared.ErrValidation("'path' is not in base64 format")
	}
	return filestore.GetFilePathRequest{Path: string(decodedPath)}, nil
}

func encodeFilePathResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	e, ok := response.(filestore.GetFilePathResponse)
	if ok && e.Failed() != nil {
		encodeError(ctx, e.Failed(), w)
		return nil
	}

	w.Header().Set("Content-Type", e.MimeType)
	_, err := w.Write(e.Payload)
	return err
}

// DeleteDocumentByIDRequest
// --------------------------------------------------------------------------
func decodeDeleteDocumentByIDRequest(_ context.Context, r *http.Request) (request interface{}, err error) {
	id := chi.URLParam(r, "id")
	if id == "" {
		return nil, shared.ErrValidation("missing 'id' path-parameter")
	}
	return document.DeleteDocumentByIDRequest{
		ID: chi.URLParam(r, "id"),
	}, nil
}

func encodeDeleteDocumentByIDResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	e, ok := response.(document.DeleteDocumentByIDResponse)
	if ok && e.Failed() != nil {
		encodeError(ctx, e.Failed(), w)
		return nil
	}
	return respondJSON(w, document.IDResult{
		Result: document.Result{
			ActionResult: document.Deleted,
			Message:      fmt.Sprintf("Document with id '%s' was deleted.", e.ID),
		},
		ID: e.ID,
	})
}

// SearchListRequest
// --------------------------------------------------------------------------
func decodeSearchListRequest(_ context.Context, r *http.Request) (request interface{}, err error) {
	t := chi.URLParam(r, "type")
	if t == "" {
		return nil, shared.ErrValidation("missing 'type' path-parameter")
	}
	name := r.URL.Query().Get("name")

	var st document.ListType
	switch strings.ToLower(t) {
	case "tags":
		st = document.Tags
	case "senders":
		st = document.Senders
	default:
		return nil, shared.ErrValidation("only {Tags, Senders} are allowed types")
	}

	return document.SearchListRequest{
		Name:       name,
		SearchType: st,
	}, nil
}

func encodeSearchListResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	e, ok := response.(document.SearchListResponse)
	if ok && e.Failed() != nil {
		encodeError(ctx, e.Failed(), w)
		return nil
	}

	return respondJSON(w, EntriesResult{
		Lenght:  len(e.Entries),
		Entries: e.Entries,
	})
}

// SearchDocumentRequest
// --------------------------------------------------------------------------
func decodeSearchDocumentsRequest(_ context.Context, r *http.Request) (request interface{}, err error) {
	var (
		title    string
		tag      string
		sender   string
		fromDate string
		toDate   string
		limit    int
		skip     int
		from     time.Time
		to       time.Time
	)

	title = r.URL.Query().Get("title")
	tag = r.URL.Query().Get("tag")
	sender = r.URL.Query().Get("sender")
	fromDate = r.URL.Query().Get("from")
	toDate = r.URL.Query().Get("to")

	from = parseDateTime(fromDate)
	to = parseDateTime(toDate)

	// defaults
	limit = parseIntVal(r.URL.Query().Get("limit"), 20)
	skip = parseIntVal(r.URL.Query().Get("skip"), 0)

	return document.SearchDocumentsRequest{
		Title:  title,
		Tag:    tag,
		Sender: sender,
		From:   from,
		Until:  to,
		Limit:  limit,
		Skip:   skip,
	}, nil
}

func encodeSearchDocumentsResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	e, ok := response.(document.SearchDocumentsResponse)
	if ok && e.Failed() != nil {
		encodeError(ctx, e.Failed(), w)
		return nil
	}

	return respondJSON(w, document.PagedDocument{
		TotalEntries: e.Result.TotalEntries,
		Documents:    e.Result.Documents,
	})
}

// SaveDocumentRequest
// --------------------------------------------------------------------------
func decodeSaveDocumentRequest(_ context.Context, r *http.Request) (request interface{}, err error) {
	payload := &DocumentRequest{}
	if err := render.Bind(r, payload); err != nil {
		return nil, shared.ErrValidation("invalid Document data supplied")
	}
	user := ensureUser(r)

	return document.SaveDocumentRequest{
		Document: *payload.Document,
		User:     *user,
	}, nil
}

func encodeSaveDocumentResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	e, ok := response.(document.SaveDocumentResponse)
	if ok && e.Failed() != nil {
		encodeError(ctx, e.Failed(), w)
		return nil
	}

	return respondJSON(w, document.IDResult{
		Result: document.Result{
			ActionResult: document.Saved,
			Message:      fmt.Sprintf("Document with id '%s' was saved.", e.ID),
		},
		ID: e.ID,
	})
}

// --------------------------------------------------------------------------------------------------------------------

const t = "about:blank"

func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	if err == nil {
		panic("encodeError with nil error")
	}
	var (
		pd            *pkgerr.ProblemDetail
		errNotFound   *shared.NotFoundError
		errValidation *shared.ValidationError
	)
	if errors.As(err, &errNotFound) {
		pd = &pkgerr.ProblemDetail{
			Type:   t,
			Title:  "object cannot be found",
			Status: http.StatusNotFound,
			Detail: errNotFound.Error(),
		}
	} else if errors.As(err, &errValidation) {
		pd = &pkgerr.ProblemDetail{
			Type:   t,
			Title:  "error in parameter-validaton",
			Status: http.StatusBadRequest,
			Detail: errValidation.Error(),
		}
	} else {
		pd = &pkgerr.ProblemDetail{
			Type:   t,
			Title:  "cannot service the request",
			Status: http.StatusInternalServerError,
			Detail: err.Error(),
		}
	}
	writeProblemJSON(w, pd)
}

func ensureUser(r *http.Request) *security.User {
	user, ok := security.UserFromContext(r.Context())
	if !ok || user == nil {
		panic("the sucurity context user is not available!")
	}
	return user
}

func respondJSON(w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}

func writeProblemJSON(w http.ResponseWriter, pd *pkgerr.ProblemDetail) {
	w.Header().Set("Content-Type", "application/problem+json; charset=utf-8")
	w.WriteHeader(pd.Status)
	b, err := json.Marshal(pd)
	if err != nil {
		stdlog.Printf("could not marshal json %v\n", err)
	}
	_, err = w.Write(b)
	if err != nil {
		stdlog.Printf("could not write bytes using http.ResponseWriter: %v\n", err)
	}
}

func parseIntVal(input string, def int) int {
	v, err := strconv.Atoi(input)
	if err != nil {
		return def
	}
	return v
}

const jsonTimeLayout = "2006-01-02T15:04:05+07:00"

func parseDateTime(input string) time.Time {
	t, err := time.Parse(jsonTimeLayout, input)
	if err != nil {
		return time.Time{}
	}
	return t
}
