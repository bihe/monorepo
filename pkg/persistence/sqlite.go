package persistence

import (
	"database/sql"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/ncruces/go-sqlite3"
	_ "github.com/ncruces/go-sqlite3/embed"
	"github.com/ncruces/go-sqlite3/ext/unicode"
	"github.com/ncruces/go-sqlite3/gormlite"

	//"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// SqliteParam is used to provide additional parameters for the sqlite connection
type SqliteParam struct {
	Key string
	Val string
}

// MustCreateSqliteConn initializes a new connection to a sqlite database.
// The func validates the existence of the sqlite db file-path.
// If the file does not exist, the function will panic.
// If the connection cannot be established, the function will panic
func MustCreateSqliteConn(connURL string) *sql.DB {
	if connURL == "" {
		panic("empty connection URL supplied")
	}
	baseDSN := stripParamsDsn(connURL)

	// sqlite has "special" connection strings which are not files on the local filesystem
	// therefor the check if the file exists is not possible.
	// e.g. file::memory:
	if !strings.Contains(baseDSN, ":memory:") {
		_, err := os.Stat(connURL)
		if os.IsNotExist(err) {
			panic(fmt.Sprintf("the defined sqlite3 database file is not available '%s': %v", baseDSN, err))
		}
	}

	// register the unicode extension
	sqlite3.AutoExtension(unicode.Register)
	// use the configured sqlite3 driver available
	db, err := sql.Open("sqlite3", baseDSN)
	if err != nil {
		panic(fmt.Sprintf("unable to create a sqlite3 connection '%s': %v", baseDSN, err))
	}
	return db
}

// CreateGormSqliteCon uses best practices for sqlite and creates two connections
// One optimized for reading and the other optimized for writing
// found the information here: https://kerkour.com/sqlite-for-servers
func CreateGormSqliteCon(db *sql.DB) (con Connection, err error) {
	var (
		read  *gorm.DB
		write *gorm.DB
	)
	if db == nil {
		return Connection{}, fmt.Errorf("invalid db connection supplied")
	}

	gormConf := &gorm.Config{
		Logger:                 logger.Default.LogMode(logger.Silent),
		SkipDefaultTransaction: true,
	}

	read, err = gorm.Open(gormlite.OpenDB(db), gormConf)
	if err != nil {
		return Connection{}, fmt.Errorf("cannot create read database connection: %w", err)
	}
	readDB, err := read.DB()
	if err != nil {
		return Connection{}, fmt.Errorf("cannot access underlying read database connection: %w", err)
	}
	SetDefaultPragmas(readDB)
	readDB.SetMaxOpenConns(max(4, runtime.NumCPU())) // read in parallel with open connection per core

	write, err = gorm.Open(gormlite.OpenDB(db), gormConf)
	if err != nil {
		return Connection{}, fmt.Errorf("cannot create write database connection: %w", err)
	}
	writeDB, err := read.DB()
	if err != nil {
		return Connection{}, fmt.Errorf("cannot access underlying write database connection: %w", err)
	}
	SetDefaultPragmas(writeDB)
	writeDB.SetMaxOpenConns(1) // only use one active connection for writing

	return Connection{
		Read:  read,
		Write: write,
		Tx:    nil,
	}, nil
}

// stripParamsDsn removes any params which were added to the connStr in the form of ?key=val
func stripParamsDsn(connStr string) string {
	i := strings.Index(connStr, "?")
	if i == -1 {
		return connStr
	}
	basePath := connStr[0:i]
	return basePath
}

// SetDefaultPragmas defines some sqlite pragmas for good performance and litestream compatibility
// https://highperformancesqlite.com/articles/sqlite-recommended-pragmas
// https://litestream.io/tips/
func SetDefaultPragmas(db *sql.DB) error {
	var (
		stmt string
		val  string
	)
	defaultPragmas := map[string]string{
		"journal_mode": "wal",   // https://www.sqlite.org/pragma.html#pragma_journal_mode
		"busy_timeout": "5000",  // https://www.sqlite.org/pragma.html#pragma_busy_timeout
		"synchronous":  "1",     // NORMAL --> https://www.sqlite.org/pragma.html#pragma_synchronous
		"cache_size":   "10000", // 10000 pages = 40MB --> https://www.sqlite.org/pragma.html#pragma_cache_size
		"foreign_keys": "1",     // 1(bool) --> https://www.sqlite.org/pragma.html#pragma_foreign_keys
	}

	// set the pragmas
	for k := range defaultPragmas {
		stmt = fmt.Sprintf("pragma %s = %s", k, defaultPragmas[k])
		if _, err := db.Exec(stmt); err != nil {
			return err
		}
	}

	// validate the pragmas
	for k := range defaultPragmas {
		row := db.QueryRow(fmt.Sprintf("pragma %s", k))
		err := row.Scan(&val)
		if err != nil {
			return err
		}
		if val != defaultPragmas[k] {
			return fmt.Errorf("could not set pragma %s to %s", k, defaultPragmas[k])
		}
	}

	return nil
}
