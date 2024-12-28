package persistence

import (
	"fmt"
	"net/url"
	"runtime"
	"strings"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// SqliteParam is used to provide additional parameters for the sqlite connection
type SqliteParam struct {
	Key string
	Val string
}

// CreateGormSqliteCon uses best practices for sqlite and creates two connections
// One optimized for reading and the other optimized for writing
// found the information here: https://kerkour.com/sqlite-for-servers
func CreateGormSqliteCon(connURL string, params []SqliteParam) (con Connection, err error) {
	var (
		read  *gorm.DB
		write *gorm.DB
	)
	if connURL == "" {
		return Connection{}, fmt.Errorf("empty connection URL supplied")
	}

	gormConf := &gorm.Config{
		Logger:                 logger.Default.LogMode(logger.Silent),
		SkipDefaultTransaction: true,
	}

	baseDS := stripParamsDsn(connURL)
	dsName := fmt.Sprintf("%s?%s", baseDS, commonSqliteParams(params))

	read, err = gorm.Open(sqlite.Open(dsName), gormConf)
	if err != nil {
		return Connection{}, fmt.Errorf("cannot create read database connection: %w", err)
	}
	readDB, err := read.DB()
	if err != nil {
		return Connection{}, fmt.Errorf("cannot access underlying read database connection: %w", err)
	}
	readDB.SetMaxOpenConns(max(4, runtime.NumCPU())) // read in parallel with open connection per core

	write, err = gorm.Open(sqlite.Open(dsName), gormConf)
	if err != nil {
		return Connection{}, fmt.Errorf("cannot create write database connection: %w", err)
	}
	writeDB, err := read.DB()
	if err != nil {
		return Connection{}, fmt.Errorf("cannot access underlying write database connection: %w", err)
	}
	writeDB.SetMaxOpenConns(1) // only use one active connection for writing

	return Connection{
		Read:  read,
		Write: write,
		Tx:    nil,
	}, nil
}

func stripParamsDsn(connStr string) string {
	i := strings.Index(connStr, "?")
	if i == -1 {
		return connStr
	}
	// strip any params which were added to the connStr
	basePath := connStr[0:i]
	return basePath
}

func commonSqliteParams(params []SqliteParam) string {
	conParams := make(url.Values)
	// default values
	conParams.Add("_txlock", "immediate")
	conParams.Add("_journal_mode", "WAL")
	conParams.Add("_busy_timeout", "5000")
	conParams.Add("_synchronous", "NORMAL")
	conParams.Add("_cache_size", "1000000000")
	conParams.Add("_foreign_keys", "true")

	for _, p := range params {
		if conParams.Get(p.Key) != "" {
			conParams.Set(p.Key, p.Val)
		} else {
			conParams.Add(p.Key, p.Val)
		}
	}
	return conParams.Encode()
}
