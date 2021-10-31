package main

import (
	"fmt"
	"os"

	"golang.binggl.net/monorepo/internal/core/api"
	"golang.binggl.net/monorepo/internal/core/app/conf"
	"golang.binggl.net/monorepo/internal/core/app/oidc"
	"golang.binggl.net/monorepo/internal/core/app/sites"
	"golang.binggl.net/monorepo/internal/core/app/store"
	"golang.binggl.net/monorepo/internal/core/app/upload"
	"golang.binggl.net/monorepo/internal/crypter"
	"golang.binggl.net/monorepo/pkg/logging"
	"golang.binggl.net/monorepo/pkg/server"
	"google.golang.org/grpc"
	"gorm.io/gorm"

	"gorm.io/driver/mysql"
)

var (
	// Version exports the application version
	Version = "1.0.0"
	// Build provides information about the application build
	Build = "localbuild"
	// AppName specifies the application itself
	AppName = "core"
	// ApplicationNameKey identifies the application in structured logging
	ApplicationNameKey = "appName"
	// HostIDKey identifies the host in structured logging
	HostIDKey = "localhost"
)

func main() {
	if err := run(Version, Build); err != nil {
		fmt.Fprintf(os.Stderr, "<< ERROR-RESULT >> '%s'\n", err)
		os.Exit(1)
	}
}

// run is the entry-point for the mydms service
// where initialization, setup and execution is done
func run(version, build string) error {
	//hostname, port, _, config := readConfig()
	hostname, port, basePath, config := server.ReadConfig("CO", func() interface{} {
		return &conf.AppConfig{} // use the correct object to deserialize the configuration
	})
	var appCfg = config.(*conf.AppConfig)

	// use the new pkg logger implementation
	logger := logConfig(*appCfg)
	// ensure closing of logfile on exit
	defer logger.Close()

	// persistence store && application version
	con, err := gorm.Open(mysql.Open(appCfg.Database.ConnectionString), &gorm.Config{})
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
		client, err = grpc.Dial(appCfg.Upload.EncGrpcConn, grpc.WithInsecure())
		if err != nil {
			panic(fmt.Sprintf("could not connect: %v", err))
		}
		encService = crypter.NewClient(client)
	}

	var (
		repo                     = store.Create(con, logger)
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
		handler = api.MakeHTTPHandler(oidcSvc, siteSvc, uploadSvc, logger, api.HTTPHandlerOptions{
			BasePath:  basePath,
			ErrorPath: appCfg.ErrorPath,
			Config:    *appCfg,
			Version:   Version,
			Build:     Build,
		})
	)

	return server.Run(server.RunOptions{
		AppName:       AppName,
		Version:       Version,
		Build:         Build,
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
