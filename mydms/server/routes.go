package server

import (
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
	"golang.binggl.net/monorepo/mydms"
	"golang.binggl.net/monorepo/mydms/config"
	"golang.binggl.net/monorepo/mydms/features/appinfo"
	"golang.binggl.net/monorepo/mydms/features/documents"
	"golang.binggl.net/monorepo/mydms/features/filestore"
	"golang.binggl.net/monorepo/mydms/features/upload"
	"golang.binggl.net/monorepo/mydms/persistence"
)

// registerRoutes defines the routes of the available handlers
func registerRoutes(e *echo.Echo, con persistence.Connection, config config.AppConfig, version mydms.VersionInfo, logger *log.Entry) (err error) {
	var (
		dr documents.Repository
	)

	dr, err = documents.NewRepository(con)
	if err != nil {
		return
	}

	// global API path
	api := e.Group("/api/v1")

	// appinfo
	ai := api.Group("/appinfo")
	aih := &appinfo.Handler{VersionInfo: version, Log: logger}
	ai.GET("", aih.GetAppInfo)

	// file
	storeSvc := filestore.NewService(filestore.S3Config{
		Region: config.Filestore.Region,
		Bucket: config.Filestore.Bucket,
		Key:    config.Filestore.Key,
		Secret: config.Filestore.Secret,
	})
	f := api.Group("/file")
	fh := filestore.NewHandler(storeSvc, logger)
	f.GET("", fh.GetFile)
	f.GET("/", fh.GetFile)

	// documents
	uClient := upload.NewClient(config.Upload.EndpointURL)
	d := api.Group("/documents")
	dh := documents.NewHandler(documents.Repositories{
		DocRepo: dr,
	}, storeSvc, uClient, logger)

	d.GET("/:type/search", dh.SearchList)
	d.GET("/:id", dh.GetDocumentByID)
	d.DELETE("/:id", dh.DeleteDocumentByID)
	d.GET("/search", dh.SearchDocuments)
	d.POST("", dh.SaveDocument)
	d.POST("/", dh.SaveDocument)

	return
}
