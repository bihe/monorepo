package core

import (
	"fmt"

	"golang.binggl.net/monorepo/internal/core/app/conf"
	"golang.binggl.net/monorepo/internal/core/app/oidc"
	"golang.binggl.net/monorepo/internal/core/app/sites"
	"golang.binggl.net/monorepo/internal/core/app/store"
	"golang.binggl.net/monorepo/pkg/config"
	"golang.binggl.net/monorepo/pkg/develop"
	"golang.binggl.net/monorepo/pkg/logging"
	"golang.binggl.net/monorepo/pkg/server"
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

	var (
		repo                     = store.NewDBStore(con)
		oidcConfig, oidcVerifier = oidc.NewConfigAndVerifier(appCfg.OIDC)
		oidcSvc                  = oidc.New(oidcConfig, oidcVerifier, appCfg.Security, repo)
		siteSvc                  = sites.New(appCfg.Security.Claim.Roles[0], repo)
		handler                  = MakeHTTPHandler(oidcSvc, siteSvc, logger, HTTPHandlerOptions{
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
