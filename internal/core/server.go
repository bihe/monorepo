package core

import (
	"fmt"

	"golang.binggl.net/monorepo/internal/core/app/conf"
	"golang.binggl.net/monorepo/internal/core/app/oidc"
	"golang.binggl.net/monorepo/internal/core/app/sites"
	"golang.binggl.net/monorepo/internal/core/app/store"
	"golang.binggl.net/monorepo/internal/core/app/upload"
	"golang.binggl.net/monorepo/internal/crypter"
	"golang.binggl.net/monorepo/pkg/config"
	"golang.binggl.net/monorepo/pkg/develop"
	"golang.binggl.net/monorepo/pkg/logging"
	"golang.binggl.net/monorepo/pkg/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"gorm.io/gorm"

	"gorm.io/driver/sqlite"
)

// run is the entry-point for the core/auth service
// where initialization, setup and execution is done
func Run(version, build, appName string) error {
	hostname, port, basePath, appCfg := server.ReadConfig[conf.AppConfig]("CO")

	// use the new pkg logger implementation
	logger := logConfig(appCfg)
	// ensure closing of logfile on exit
	defer logger.Close()

	// persistence store && application version
	con, err := gorm.Open(sqlite.Open(appCfg.Database.ConnectionString), &gorm.Config{})
	if err != nil {
		panic(fmt.Sprintf("cannot create database connection: %v", err))
	}

	// Build the layers of the gateway-server "onion" from the inside out. First, the
	// business logic store.Repository; oidc.Service; sites.Service; ...
	// and finally, a series of concrete transport adapters.
	var (
		encService crypter.EncryptionService
		client     *grpc.ClientConn
	)
	if appCfg.Upload.EncGrpcConn != "" {
		client, err = grpc.Dial(appCfg.Upload.EncGrpcConn, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			panic(fmt.Sprintf("could not connect: %v", err))
		}
		encService = crypter.NewClient(client)
	}

	var (
		repo                     = store.NewDBStore(con)
		oidcConfig, oidcVerifier = oidc.NewConfigAndVerifier(appCfg.OIDC)
		oidcSvc                  = oidc.New(oidcConfig, oidcVerifier, appCfg.Security, repo)
		siteSvc                  = sites.New(appCfg.Security.Claim.Roles[0], repo)
		uploadSvc                = upload.NewService(upload.ServiceOptions{
			Logger:           logger,
			Store:            upload.NewStore(appCfg.Upload.UploadPath),
			MaxUploadSize:    appCfg.Upload.MaxUploadSize,
			AllowedFileTypes: appCfg.Upload.AllowedFileTypes,
			Crypter:          encService,
			TimeOut:          "30s",
		},
		)
		handler = MakeHTTPHandler(oidcSvc, siteSvc, uploadSvc, logger, HTTPHandlerOptions{
			BasePath:  basePath,
			ErrorPath: appCfg.ErrorPath,
			Config:    appCfg,
			Version:   version,
			Build:     build,
		})
	)

	// only run the reload-server in development
	if appCfg.Environment == config.Development {
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

func logConfig(cfg conf.AppConfig) logging.Logger {
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
