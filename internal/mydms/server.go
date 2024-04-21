package mydms

import (
	"fmt"

	"golang.binggl.net/monorepo/internal/mydms/app/config"
	"golang.binggl.net/monorepo/internal/mydms/app/crypter"
	"golang.binggl.net/monorepo/internal/mydms/app/document"
	"golang.binggl.net/monorepo/internal/mydms/app/filestore"
	"golang.binggl.net/monorepo/internal/mydms/app/shared"
	"golang.binggl.net/monorepo/internal/mydms/app/upload"
	conf "golang.binggl.net/monorepo/pkg/config"
	"golang.binggl.net/monorepo/pkg/develop"
	"golang.binggl.net/monorepo/pkg/logging"
	"golang.binggl.net/monorepo/pkg/server"

	// include the sqlite driver for runtime
	_ "github.com/mattn/go-sqlite3"
)

// run is the entry-point for the core/auth service
// where initialization, setup and execution is done
func Run(version, build, appName string) error {
	hostname, port, basePath, appCfg := server.ReadConfig[config.AppConfig]("my")

	// use the new pkg logger implementation
	logger := logConfig(appCfg)
	// ensure closing of logfile on exit
	defer logger.Close()

	// shared.store && application version
	con := shared.NewConnForDb("sqlite3", appCfg.Database.ConnectionString)
	repo, err := document.NewRepository(con)
	if err != nil {
		panic(fmt.Sprintf("cannot establish database connection: %v", err))
	}

	var (
		fileSvc = filestore.NewService(logger, filestore.S3Config{
			Bucket: appCfg.Filestore.Bucket,
			Region: appCfg.Filestore.Region,
			Key:    appCfg.Filestore.Key,
			Secret: appCfg.Filestore.Secret,
		})
		crypterSvc = crypter.NewService(logger)
		uploadSvc  = upload.NewService(upload.ServiceOptions{
			Logger:           logger,
			Store:            upload.NewStore(appCfg.Upload.UploadPath),
			MaxUploadSize:    appCfg.Upload.MaxUploadSize,
			AllowedFileTypes: appCfg.Upload.AllowedFileTypes,
			Crypter:          crypterSvc,
			TimeOut:          "30s",
		})
		docSvc  = document.NewService(logger, repo, fileSvc, uploadSvc)
		handler = MakeHTTPHandler(docSvc, uploadSvc, fileSvc, logger, HTTPHandlerOptions{
			BasePath:  basePath,
			ErrorPath: appCfg.ErrorPath,
			Config:    appCfg,
			Version:   version,
			Build:     build,
		})
	)

	// only run the reload-server in development
	if appCfg.Environment == conf.Development {
		reload := develop.NewReloadServer()
		reload.Start()
	}

	return server.Run(server.RunOptions{
		AppName:       appName,
		Version:       version,
		Build:         build,
		HostName:      hostname,
		Port:          port,
		Environment:   string(appCfg.Environment),
		ServerHandler: handler,
		Logger:        logger,
	})
}

func logConfig(cfg config.AppConfig) logging.Logger {
	return logging.New(logging.LogConfig{
		FilePath:      cfg.Logging.FilePath,
		LogLevel:      cfg.Logging.LogLevel,
		GrayLogServer: cfg.Logging.GrayLogServer,
		Trace: logging.TraceConfig{
			AppName: cfg.AppName,
			HostID:  cfg.HostID,
		},
	}, cfg.Environment)
}
