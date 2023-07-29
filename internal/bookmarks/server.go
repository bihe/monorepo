package bookmarks

import (
	"fmt"
	"path"
	"strings"

	"golang.binggl.net/monorepo/internal/bookmarks/app/bookmarks"
	"golang.binggl.net/monorepo/internal/bookmarks/app/conf"
	"golang.binggl.net/monorepo/internal/bookmarks/app/store"
	"golang.binggl.net/monorepo/pkg/logging"
	"golang.binggl.net/monorepo/pkg/server"
	"gorm.io/gorm"

	"gorm.io/driver/sqlite"
)

// run is the entry-point for the bookmarks service
// where initialization, setup and execution is done
func Run(version, build, appName string) error {
	//hostname, port, _, config := readConfig()
	hostname, port, basePath, appCfg := server.ReadConfig[conf.AppConfig]("BM")
	// use the new pkg logger implementation
	logger := logConfig(appCfg)
	// ensure closing of logfile on exit
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

const journalMode = "_journal_mode"
const journalModeValue = "WAL"

// setupRepositories enables the SQLITE repository for bookmark storage
func setupRepositories(config *conf.AppConfig, logger logging.Logger) (store.BookmarkRepository, store.FaviconRepository) {
	// Documentation for DNS based SQLITE PRAGMAS: https://github.com/mattn/go-sqlite3
	// add the WAL mode for SQLITE
	dbConnStr := config.Database.ConnectionString
	if !strings.Contains(dbConnStr, journalMode) {
		paramDelim := "?"
		if strings.Contains(dbConnStr, "?") {
			paramDelim = "&"
		}
		dbConnStr += fmt.Sprintf("%s%s=%s", paramDelim, journalMode, journalModeValue)
	}

	con, err := gorm.Open(sqlite.Open(dbConnStr), &gorm.Config{})
	if err != nil {
		panic(fmt.Sprintf("cannot create database connection: %v", err))
	}

	db, err := con.DB()
	if err != nil {
		panic(fmt.Sprintf("cannot access underlying database: %v", err))
	}
	// this is done because of SQLITE WAL
	// found here: https://stackoverflow.com/questions/35804884/sqlite-concurrent-writing-performance
	db.SetMaxOpenConns(1)

	return store.CreateBookmarkRepo(con, logger), store.CreateFaviconRepo(con, logger)
}
