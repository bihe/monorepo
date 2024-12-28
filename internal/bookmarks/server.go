package bookmarks

import (
	"fmt"
	"path"

	"golang.binggl.net/monorepo/internal/bookmarks/app/bookmarks"
	"golang.binggl.net/monorepo/internal/bookmarks/app/conf"
	"golang.binggl.net/monorepo/internal/bookmarks/app/store"
	"golang.binggl.net/monorepo/pkg/config"
	"golang.binggl.net/monorepo/pkg/develop"
	"golang.binggl.net/monorepo/pkg/logging"
	"golang.binggl.net/monorepo/pkg/persistence"
	"golang.binggl.net/monorepo/pkg/server"
)

// run is the entry-point for the bookmarks service
// where initialization, setup and execution is done
func Run(version, build, appName string) error {
	//hostname, port, _, config := readConfig()
	hostname, port, basePath, appCfg := server.ReadConfig[conf.AppConfig]("BM")
	// use the new pkg logger implementation
	logger := logConfig(appCfg)
	// ensure closing of log-file on exit
	defer logger.Close()

	var (
		bRepo, fRepo = setupRepositories(&appCfg, logger)
		app          = &bookmarks.Application{
			Logger:        logger,
			BookmarkStore: bRepo,
			FavStore:      fRepo,
			FaviconPath:   path.Join(basePath, appCfg.FaviconUploadPath),
		}
		handler = MakeHTTPHandler(app, logger, HTTPHandlerOptions{
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

// setupRepositories enables the SQLITE repositories for the application
func setupRepositories(config *conf.AppConfig, logger logging.Logger) (store.BookmarkRepository, store.FaviconRepository) {
	con, err := persistence.CreateGormSqliteCon(config.Database.ConnectionString, make([]persistence.SqliteParam, 0))
	if err != nil {
		panic(fmt.Sprintf("cannot create database connection: %v", err))
	}
	return store.CreateBookmarkRepo(con, logger), store.CreateFaviconRepo(con, logger)
}
