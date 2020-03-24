package main

import (
	"github.com/bihe/mydms/features/appinfo"
	"github.com/bihe/mydms/features/documents"
	"github.com/bihe/mydms/features/filestore"
	"github.com/bihe/mydms/features/upload"
	"github.com/bihe/mydms/internal"
	"github.com/bihe/mydms/internal/config"
	"github.com/bihe/mydms/internal/persistence"
	"github.com/labstack/echo/v4"
)

// registerRoutes defines the routes of the available handlers
func registerRoutes(e *echo.Echo, con persistence.Connection, config config.AppConfig, version internal.VersionInfo) (err error) {
	var (
		ur upload.Repository
		dr documents.Repository
	)

	ur, err = upload.NewRepository(con)
	if err != nil {
		return
	}
	dr, err = documents.NewRepository(con)
	if err != nil {
		return
	}

	// global API path
	api := e.Group("/api/v1")

	// appinfo
	ai := api.Group("/appinfo")
	aih := &appinfo.Handler{VersionInfo: version}
	ai.GET("", aih.GetAppInfo)

	// upload
	u := api.Group("/upload")
	uploadConfig := upload.Config{
		AllowedFileTypes: config.Upload.AllowedFileTypes,
		MaxUploadSize:    config.Upload.MaxUploadSize,
		UploadPath:       config.Upload.UploadPath,
	}
	uh := upload.NewHandler(ur, uploadConfig)
	u.POST("/file", uh.UploadFile)

	// file
	storeSvc := filestore.NewService(filestore.S3Config{
		Region: config.Filestore.Region,
		Bucket: config.Filestore.Bucket,
		Key:    config.Filestore.Key,
		Secret: config.Filestore.Secret,
	})
	f := api.Group("/file")
	fh := filestore.NewHandler(storeSvc)
	f.GET("", fh.GetFile)
	f.GET("/", fh.GetFile)

	// documents
	d := api.Group("/documents")
	dh := documents.NewHandler(documents.Repositories{
		DocRepo:    dr,
		UploadRepo: ur,
	}, storeSvc, uploadConfig)

	d.GET("/:type/search", dh.SearchList)
	d.GET("/:id", dh.GetDocumentByID)
	d.DELETE("/:id", dh.DeleteDocumentByID)
	d.GET("/search", dh.SearchDocuments)
	d.POST("", dh.SaveDocument)
	d.POST("/", dh.SaveDocument)

	return
}
