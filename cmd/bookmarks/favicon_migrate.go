package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"golang.binggl.net/monorepo/internal/bookmarks/app/store"
	"golang.binggl.net/monorepo/pkg/logging"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	glog "gorm.io/gorm/logger"

	flag "github.com/spf13/pflag"
)

const journalMode = "_journal_mode"
const journalModeValue = "WAL"

var logger = logging.NewNop()

// setupRepository enables the SQLITE repository for bookmark storage
func setupRepositories(dbConnStr string) (store.BookmarkRepository, store.FaviconRepository, *sql.DB) {
	// Documentation for DNS based SQLITE PRAGMAS: https://github.com/mattn/go-sqlite3
	// add the WAL mode for SQLITE
	if !strings.Contains(dbConnStr, journalMode) {
		paramDelim := "?"
		if strings.Contains(dbConnStr, "?") {
			paramDelim = "&"
		}
		dbConnStr += fmt.Sprintf("%s%s=%s", paramDelim, journalMode, journalModeValue)
	}

	con, err := gorm.Open(sqlite.Open(dbConnStr), &gorm.Config{
		Logger: glog.Default.LogMode(glog.Silent),
	})
	if err != nil {
		panic(fmt.Sprintf("cannot create database connection: %v", err))
	}
	con.AutoMigrate(&store.Favicon{})
	db, err := con.DB()
	if err != nil {
		panic(fmt.Sprintf("cannot access underlying database: %v", err))
	}
	// this is done because of SQLITE WAL
	// found here: https://stackoverflow.com/questions/35804884/sqlite-concurrent-writing-performance
	db.SetMaxOpenConns(1)

	return store.CreateBookmarkRepo(con, logger), store.CreateFaviconRepo(con, logger), db
}

func main() {
	var (
		con   string
		usr   string
		fPath string
	)

	flag.StringVar(&con, "connstr", "", "connection-string to the sqlite DB")
	flag.StringVar(&usr, "user", "", "the name of the user of the bookmarks")
	flag.StringVar(&fPath, "path", "", "path to the favicons on disk")
	flag.Parse()

	if con == "" || usr == "" || fPath == "" {
		flag.Usage()
		return
	}

	bRepo, fRepo, db := setupRepositories(con)
	defer db.Close()

	bookmarks, err := bRepo.GetAllBookmarks(usr)
	if err != nil {
		log.Fatalf("cannot get bookmarks for user '%s'; %v", usr, err)
	}

	for _, b := range bookmarks {
		if b.Favicon != "" {
			fullPath := path.Join(fPath, b.Favicon)

			if strings.HasPrefix(fullPath, "~/") {
				dirname, _ := os.UserHomeDir()
				fullPath = filepath.Join(dirname, fullPath[2:])
			}

			if err = writeFavicon(b.Favicon, fullPath, fRepo); err == nil {
				log.Printf("### the favicon %s was migrated", b.Favicon)
			}
		}
	}
}

func writeFavicon(id string, fullPath string, repo store.FaviconRepository) error {
	meta, err := os.Stat(fullPath)
	if err != nil {
		log.Printf(">>>>> the specified favicon '%s' is not available\n", fullPath)
		return err
	}
	payload, err := os.ReadFile(fullPath)
	if err != nil {
		log.Printf(">>>>> the specified favicon '%s' could not be read, %v", fullPath, err)
		return err
	}

	_, err = repo.Save(store.Favicon{
		ID:           id,
		Payload:      payload,
		LastModified: meta.ModTime(),
	})
	if err != nil {
		log.Printf(">>>>> could not store the favicon; %v", err)
		return err
	}
	return nil
}
