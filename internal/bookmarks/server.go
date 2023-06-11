package main

import (
	"fmt"
	"os"
	"path"
	"strings"

	"golang.binggl.net/monorepo/internal/bookmarks/api"
	"golang.binggl.net/monorepo/internal/bookmarks/app/bookmarks"
	"golang.binggl.net/monorepo/internal/bookmarks/app/conf"
	"golang.binggl.net/monorepo/internal/bookmarks/app/store"
	"golang.binggl.net/monorepo/pkg/logging"
	"golang.binggl.net/monorepo/pkg/server"
	"gorm.io/gorm"

	"gorm.io/driver/sqlite"
)

var (
	// Version exports the application version
	Version = "1.0.0"
	// Build provides information about the application build
	Build = "localbuild"
	// AppName specifies the application itself
	AppName = "bookmarks"
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

// run is the entry-point for the bookmarks service
// where initialization, setup and execution is done
func run(version, build string) error {
	//hostname, port, _, config := readConfig()
	hostname, port, basePath, config := server.ReadConfig("BM", func() interface{} {
		return &conf.AppConfig{} // use the correct object to deserialize the configuration
	})
	var appCfg = config.(*conf.AppConfig)

	// use the new pkg logger implementation
	logger := logConfig(*appCfg)
	// ensure closing of logfile on exit
	defer logger.Close()

	var (
		bRepo, fRepo = setupRepositories(appCfg, logger)
		app          = &bookmarks.Application{
			Logger:        logger,
			BookmarkStore: bRepo,
			FavStore:      fRepo,
			FaviconPath:   path.Join(basePath, appCfg.FaviconUploadPath),
		}
		handler = api.MakeHTTPHandler(app, logger, api.HTTPHandlerOptions{
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
