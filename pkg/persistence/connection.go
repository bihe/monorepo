// Package persistence simplifies the database interaction by providing helper functions
// it uses sqlx internally
package persistence

import (
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
)

// Connection defines a storage/database/... connection
type Connection struct {
	*sqlx.DB
	Active bool
}

// NewConn creates a connection to a store/repository
func NewConn(connstr string) Connection {
	// dialect is specifically set to "mysql"
	db := sqlx.MustConnect("mysql", connstr)
	return Connection{DB: db, Active: true}
}

// NewConnForDb creates a connection to a store/repository
func NewConnForDb(dbType, connstr string) Connection {
	db := sqlx.MustConnect(dbType, connstr)
	return Connection{DB: db, Active: true}
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
