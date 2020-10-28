package mydms

import (
	"context"
	"time"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"golang.binggl.net/monorepo/internal/mydms/app/appinfo"
	"golang.binggl.net/monorepo/internal/mydms/app/document"
	"golang.binggl.net/monorepo/internal/mydms/app/filestore"
)

// Endpoints collects all of the endpoints that compose a profile service. It's
// meant to be used as a helper struct, to collect all of the endpoints into a
// single parameter.
type Endpoints struct {
	GetAppInfoEndpoint         endpoint.Endpoint
	GetDocumentByIDEndpoint    endpoint.Endpoint
	DeleteDocumentByIDEndpoint endpoint.Endpoint
	SearchListEndpoint         endpoint.Endpoint
	SearchDocumentsEndpoint    endpoint.Endpoint
	SaveDocumentEndpoint       endpoint.Endpoint
	GetFileEndpoint            endpoint.Endpoint
}

// MakeServerEndpoints returns an Endpoints struct where each endpoint invokes
// the corresponding method on the provided service.
func MakeServerEndpoints(ai appinfo.Service, doc document.Service, f filestore.FileService, logger log.Logger) Endpoints {

	// Application endpoints

	var appInfoEndopint endpoint.Endpoint
	{
		appInfoEndopint = appinfo.MakeGetAppInfoEndpoint(ai)
		appInfoEndopint = EndpointLoggingMiddleware(log.With(logger, "method", "GetAppInfo"))(appInfoEndopint)
	}

	// Document endpoints

	var documentByIDEndpoint endpoint.Endpoint
	{
		documentByIDEndpoint = document.MakeGetDocumentByIDEndpoint(doc)
		documentByIDEndpoint = EndpointLoggingMiddleware(log.With(logger, "method", "GetDocumentByID"))(documentByIDEndpoint)
	}

	var deleteDocumentByIDEndpoint endpoint.Endpoint
	{
		deleteDocumentByIDEndpoint = document.MakeDeleteDocumentByIDEndpoint(doc)
		deleteDocumentByIDEndpoint = EndpointLoggingMiddleware(log.With(logger, "method", "DeleteDocumentByID"))(deleteDocumentByIDEndpoint)
	}

	var searchListEndpoint endpoint.Endpoint
	{
		searchListEndpoint = document.MakeSearchListEndpoint(doc)
		searchListEndpoint = EndpointLoggingMiddleware(log.With(logger, "method", "SearchList"))(searchListEndpoint)
	}

	var searchDocumentsEndpoint endpoint.Endpoint
	{
		searchDocumentsEndpoint = document.MakeSearchDocumentsEndpoint(doc)
		searchDocumentsEndpoint = EndpointLoggingMiddleware(log.With(logger, "method", "SearchDocuments"))(searchDocumentsEndpoint)
	}

	var saveDocumentEndpoint endpoint.Endpoint
	{
		saveDocumentEndpoint = document.MakeSaveDocumentEnpoint(doc)
		saveDocumentEndpoint = EndpointLoggingMiddleware(log.With(logger, "method", "SaveDocument"))(saveDocumentEndpoint)
	}

	// File endpoints

	var getFileEndpoint endpoint.Endpoint
	{
		getFileEndpoint = filestore.MakeGetFileEndpoint(f)
		getFileEndpoint = EndpointLoggingMiddleware(log.With(logger, "method", "GetFile"))(getFileEndpoint)
	}

	return Endpoints{
		GetAppInfoEndpoint:         appInfoEndopint,
		GetDocumentByIDEndpoint:    documentByIDEndpoint,
		DeleteDocumentByIDEndpoint: deleteDocumentByIDEndpoint,
		SearchListEndpoint:         searchListEndpoint,
		SearchDocumentsEndpoint:    searchDocumentsEndpoint,
		SaveDocumentEndpoint:       saveDocumentEndpoint,
		GetFileEndpoint:            getFileEndpoint,
	}
}

// --------------------------------------------------------------------------
// Logging middleware
// --------------------------------------------------------------------------

// EndpointLoggingMiddleware returns an endpoint middleware that logs the
// duration of each invocation, and the resulting error, if any.
func EndpointLoggingMiddleware(logger log.Logger) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			defer func(begin time.Time) {
				logger.Log("transport_error", err, "took", time.Since(begin))
			}(time.Now())
			return next(ctx, request)
		}
	}
}
