package persistence

import (
	"fmt"

	"gorm.io/gorm"
)

// A Connection struct abstracts the access to the underlying database connections
// It is used for reading from a database connection and writing to a database connection.
// The writing is assumed is always in the context of a database transaction.
type Connection struct {
	// Read is a read connection used for fast access to the underlying database transaction
	Read *gorm.DB
	// Write is a write connection which is used primarily to write in particular to create a transaction connection
	Write *gorm.DB
	// Tx is a transaction which enables atomic access with rollback
	Tx *gorm.DB
}

// R returns a suitable connection. It is either read focused connection
// or a transaction.
// The func panics if no read connection is available
func (c Connection) R() *gorm.DB {
	if c.Tx != nil {
		return c.Tx
	}
	if c.Read == nil {
		panic("no read database connection is available")
	}
	return c.Read
}

// W retrieves a write connection. If this is a transaction use the tx connection
// The func panics if no write connection is available
func (c Connection) W() *gorm.DB {
	if c.Tx != nil {
		return c.Tx
	}
	if c.Write == nil {
		panic("no write database connection is available")
	}
	return c.Write
}

// Begin starts a transaction and executes the provided handle function in a transaction context
func (c Connection) Begin(handle func(c Connection) error) error {
	if c.Tx != nil {
		return fmt.Errorf("a transaction is already available, will not start a new one")
	}
	return c.Write.Transaction(func(tx *gorm.DB) error {
		return handle(Connection{
			Read:  c.Read,
			Write: c.Write,
			Tx:    tx,
		})
	})
}
