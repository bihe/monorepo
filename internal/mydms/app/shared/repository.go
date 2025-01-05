package shared

import (
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"

	"golang.binggl.net/monorepo/pkg/persistence"

	//_ "github.com/mattn/go-sqlite3"
	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
)

// BaseRepository defines the common methods of all implementations
type BaseRepository interface {
	// CreateAtomic returns a new atomic object
	CreateAtomic() (Atomic, error)
}

// Connection defines a storage/database/... connection
type Connection struct {
	*sqlx.DB
	Active bool
}

// NewConnForSqlite creates a connection to a store/repository
func NewConnForSqlite(connstr string) Connection {
	dbType := "sqlite3"
	db := persistence.MustCreateSqliteConn(connstr)
	persistence.SetDefaultPragmas(db)
	dbx := sqlx.NewDb(db, dbType)
	return Connection{DB: dbx, Active: true}
}

// NewFromDB creates a new connection based on existing DB handle
func NewFromDB(db *sqlx.DB) Connection {
	return Connection{DB: db, Active: true}
}

// Atomic defines a transactional operation - like the A in ACID https://en.wikipedia.org/wiki/ACID
type Atomic struct {
	*sqlx.Tx
	Active bool
}

// CreateAtomic starts a new transaction
func (c Connection) CreateAtomic() (Atomic, error) {
	tx, err := c.Beginx()
	if err != nil {
		return Atomic{}, err
	}
	return Atomic{Tx: tx, Active: true}, nil
}

// HandleTX completes the given transaction
// if an error is supplied a rollback is performed
func HandleTX(complete bool, atomic *Atomic, err error) error {
	if complete {
		switch err {
		case nil:
			return atomic.Commit()
		default:
			log.Printf("could not complete the tx: %v", err)
			if atomic != nil {
				log.Printf("will try to rollback the tx")
				if e := atomic.Rollback(); e != nil {
					return fmt.Errorf("%v; could not rollback tx: %v", err, e)
				}
			} else {
				log.Printf("cannot rollback the tx, no atomic object available")
			}
		}
	}
	return err
}

// CheckTX returns a new Atomic object if the supplied object is inactive
// if the supplied object is active - the same object is returned
func CheckTX(c Connection, a *Atomic) (*Atomic, error) {
	if !a.Active {
		atomic, err := c.CreateAtomic()
		if err != nil {
			return nil, err
		}
		return &atomic, nil
	}
	return a, nil
}
