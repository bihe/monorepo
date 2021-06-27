package main

import (
	"fmt"
	"os"

	"golang.binggl.net/monorepo/internal/gway"
	"golang.binggl.net/monorepo/internal/gway/app/conf"
	"golang.binggl.net/monorepo/internal/gway/app/oidc"
	"golang.binggl.net/monorepo/internal/gway/app/store"
	"golang.binggl.net/monorepo/pkg/logging"
	"golang.binggl.net/monorepo/pkg/server"
	"gorm.io/gorm"

	"gorm.io/driver/mysql"
)

var (
	// Version exports the application version
	Version = "1.0.0"
	// Build provides information about the application build
	Build = "localbuild"
	// AppName specifies the application itself
	AppName = "gway"
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
	hostname, port, basePath, config := server.ReadConfig("GW", func() interface{} {
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
		repo                     = store.Create(con, logger)
		oidcConfig, oidcVerifier = oidc.NewConfigAndVerifier(appCfg.OIDC)
		oidcSvc                  = oidc.New(oidcConfig, oidcVerifier, appCfg.Security, repo)
		handler                  = gway.MakeHTTPHandler(oidcSvc, logger, gway.HTTPHandlerOptions{
			BasePath:  basePath,
			ErrorPath: appCfg.ErrorPath,
			Config:    *appCfg,
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
