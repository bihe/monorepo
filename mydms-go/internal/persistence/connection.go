package persistence

import (
	"fmt"

	"github.com/jmoiron/sqlx"

	log "github.com/sirupsen/logrus"
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
			log.Errorf("could not complete the transaction: %v", err)
			if e := atomic.Rollback(); e != nil {
				return fmt.Errorf("%v; could not rollback transaction: %v", err, e)
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
