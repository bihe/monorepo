package main

import (
	"fmt"
	"os"

	"golang.binggl.net/monorepo/internal/mydms"
	"golang.binggl.net/monorepo/internal/mydms/app/appinfo"
	"golang.binggl.net/monorepo/internal/mydms/app/document"
	"golang.binggl.net/monorepo/internal/mydms/app/filestore"
	"golang.binggl.net/monorepo/internal/mydms/app/upload"
	"golang.binggl.net/monorepo/pkg/logging"
	"golang.binggl.net/monorepo/pkg/persistence"
	"golang.binggl.net/monorepo/pkg/server"

	// include the sqlite driver for runtime
	_ "github.com/mattn/go-sqlite3"
)

var (
	// Version exports the application version
	Version = "3.0.0"
	// Build provides information about the application build
	Build = "localbuild"
	// AppName specifies the application itself
	AppName = "mydms-gokit"
	// ApplicationNameKey identifies the application in structured logging
	ApplicationNameKey = "appName"
	// HostIDKey identifies the host in structured logging
	HostIDKey = "hostID"
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
	hostname, port, basePath, appCfg := server.ReadConfig[mydms.AppConfig]("my")
	// use the new pkg logger implementation
	logger := logConfig(appCfg)
	// ensure closing of logfile on exit
	defer logger.Close()

	// persistence store && application version
	con := persistence.NewConnForDb("sqlite3", appCfg.Database.ConnectionString)
	repo, err := document.NewRepository(con)
	if err != nil {
		panic(fmt.Sprintf("cannot establish database connection: %v", err))
	}

	// Build the layers of the appInfoSvc "onion" from the inside out. First, the
	// business logic appInfoSvc; then, the set of endpoints that wrap the appInfoSvc;
	// and finally, a series of concrete transport adapters. The adapters, like
	// the HTTP handler or the gRPC server, are the bridge between Go kit and
	// the interfaces that the transports expect. Note that we're not binding
	// them to ports or anything yet; we'll do that next.
	var (
		fileSvc = filestore.NewService(logger, filestore.S3Config{
			Bucket: appCfg.Filestore.Bucket,
			Region: appCfg.Filestore.Region,
			Key:    appCfg.Filestore.Key,
			Secret: appCfg.Filestore.Secret,
		})
		uploadClient = upload.NewClient(appCfg.Upload.EndpointURL)
		appInfoSvc   = appinfo.NewService(logger, Version, Build)
		docSvc       = document.NewService(logger, repo, fileSvc, uploadClient)
		endpoints    = mydms.MakeServerEndpoints(appInfoSvc, docSvc, fileSvc, logger)
		apiSrv       = mydms.MakeHTTPHandler(endpoints, logger, mydms.HTTPHandlerOptions{
			BasePath:     basePath,
			ErrorPath:    "/error",
			AssetConfig:  appCfg.Assets,
			CookieConfig: appCfg.Cookies,
			CorsConfig:   appCfg.Cors,
			JWTConfig:    appCfg.Security,
		})
	)

	return server.Run(server.RunOptions{
		AppName:       AppName,
		Version:       Version,
		Build:         Build,
		HostName:      hostname,
		Port:          port,
		Environment:   string(appCfg.Environment),
		ServerHandler: apiSrv,
		Logger:        logger,
	})
}

func logConfig(cfg mydms.AppConfig) logging.Logger {
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
