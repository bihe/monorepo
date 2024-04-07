package persistence

import (
	"fmt"
	"net/url"
	"runtime"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// SqliteParam is used to provide addtional parameters for the sqlite connection
type SqliteParam struct {
	Key string
	Val string
}

// CreateGormSqliteRWCon uses best practises for sqlite and creates two connections
// One optimized for reading and the other optimized for writing
// found the information here: https://kerkour.com/sqlite-for-servers
func CreateGormSqliteRWCon(connURL string, params []SqliteParam) (read, write *gorm.DB, err error) {
	if connURL == "" {
		return nil, nil, fmt.Errorf("empty connection URL supplied")
	}

	gormConf := &gorm.Config{
		Logger:                 logger.Default.LogMode(logger.Silent),
		SkipDefaultTransaction: true,
	}

	dsName := fmt.Sprintf("%s?%s", connURL, commonSqliteParams(params))

	read, err = gorm.Open(sqlite.Open(dsName), gormConf)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot create read database connection: %w", err)
	}
	readDB, err := read.DB()
	if err != nil {
		return nil, nil, fmt.Errorf("cannot access underlying read database connection: %w", err)
	}
	readDB.SetMaxOpenConns(max(4, runtime.NumCPU())) // read in parallel with open connection per core

	write, err = gorm.Open(sqlite.Open(dsName), gormConf)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot create write database connection: %w", err)
	}
	writeDB, err := read.DB()
	if err != nil {
		return nil, nil, fmt.Errorf("cannot access underlying write database connection: %w", err)
	}
	writeDB.SetMaxOpenConns(1) // only use one active connection for writing

	return read, write, nil
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
