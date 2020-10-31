package mydms

import (
	"context"
	"fmt"
	"time"

	"github.com/go-kit/kit/endpoint"
	"golang.binggl.net/monorepo/internal/mydms/app/appinfo"
	"golang.binggl.net/monorepo/internal/mydms/app/document"
	"golang.binggl.net/monorepo/internal/mydms/app/filestore"
	"golang.binggl.net/monorepo/pkg/logging"
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
func MakeServerEndpoints(ai appinfo.Service, doc document.Service, f filestore.FileService, logger logging.Logger) Endpoints {

	// Application endpoints

	var appInfoEndopint endpoint.Endpoint
	{
		appInfoEndopint = appinfo.MakeGetAppInfoEndpoint(ai)
		appInfoEndopint = EndpointLoggingMiddleware(logger, "GetAppInfo")(appInfoEndopint)
	}

	// Document endpoints

	var documentByIDEndpoint endpoint.Endpoint
	{
		documentByIDEndpoint = document.MakeGetDocumentByIDEndpoint(doc)
		documentByIDEndpoint = EndpointLoggingMiddleware(logger, "GetDocumentByID")(documentByIDEndpoint)
	}

	var deleteDocumentByIDEndpoint endpoint.Endpoint
	{
		deleteDocumentByIDEndpoint = document.MakeDeleteDocumentByIDEndpoint(doc)
		deleteDocumentByIDEndpoint = EndpointLoggingMiddleware(logger, "DeleteDocumentByID")(deleteDocumentByIDEndpoint)
	}

	var searchListEndpoint endpoint.Endpoint
	{
		searchListEndpoint = document.MakeSearchListEndpoint(doc)
		searchListEndpoint = EndpointLoggingMiddleware(logger, "SearchList")(searchListEndpoint)
	}

	var searchDocumentsEndpoint endpoint.Endpoint
	{
		searchDocumentsEndpoint = document.MakeSearchDocumentsEndpoint(doc)
		searchDocumentsEndpoint = EndpointLoggingMiddleware(logger, "SearchDocuments")(searchDocumentsEndpoint)
	}

	var saveDocumentEndpoint endpoint.Endpoint
	{
		saveDocumentEndpoint = document.MakeSaveDocumentEnpoint(doc)
		saveDocumentEndpoint = EndpointLoggingMiddleware(logger, "SaveDocument")(saveDocumentEndpoint)
	}

	// File endpoints

	var getFileEndpoint endpoint.Endpoint
	{
		getFileEndpoint = filestore.MakeGetFileEndpoint(f)
		getFileEndpoint = EndpointLoggingMiddleware(logger, "GetFile")(getFileEndpoint)
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
func EndpointLoggingMiddleware(logger logging.Logger, endpointName string) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			logger.Info(fmt.Sprintf("call endpoint: %s", endpointName))
			defer func(begin time.Time) {
				logger.Info(fmt.Sprintf("%s: endpoint stats", endpointName), logging.LogV("took", time.Since(begin).String()))
				if err != nil {
					logger.Error(fmt.Sprintf("%s: transport-error", endpointName), logging.ErrV(err))
				}
			}(time.Now())
			return next(ctx, request)
		}
	}
}
