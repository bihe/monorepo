package persistence

import (
	"fmt"

	"gorm.io/gorm"
)

// Unit wraps a transaction connection handle
type Unit struct {
	Tx *gorm.DB
}

// DBRepository defines the bare minimum logic to create repository for database access
type DBRepository interface {
	// Con returns a suiteable connection to work with
	Con() *gorm.DB

	// WriteCon returns the connection used for writing
	WriteCon() *gorm.DB

	// InUnitOfWork performs the operation within a transaction
	InUnitOfWork(handle func(u Unit) error) error
}

// --------------------------------------------------------------------------
//   Basic repository using the gorm package
// --------------------------------------------------------------------------

var _ DBRepository = &GormRepository{}

// GormRepository manages read/write/transaction connectes
// to interact with databases.
// The repository uses the gorm package to interact with databases
type GormRepository struct {
	Tx    *gorm.DB
	Read  *gorm.DB
	Write *gorm.DB
}

// Con returns a suitable connection. It is either read focused connection
// or a transaction.
// The func panics if no read connection is available
func (r *GormRepository) Con() *gorm.DB {
	if r.Tx != nil {
		return r.Tx
	}
	if r.Read == nil {
		panic("no read database connection is available")
	}
	return r.Read
}

// Retrieve a write connection. If this is a transaction use the tx connection
// The func panics if no write connection is available
func (r *GormRepository) WriteCon() *gorm.DB {
	if r.Tx != nil {
		return r.Tx
	}
	if r.Write == nil {
		panic("no write database connection is available")
	}
	return r.Read
}

// InUnitOfWork uses a transaction to execute the supplied function
func (r *GormRepository) InUnitOfWork(handle func(u Unit) error) error {
	return r.Write.Transaction(func(tx *gorm.DB) error {
		// be sure the stop recursion here
		if r.Tx != nil {
			return fmt.Errorf("a shared connection/transaction is already available, will not start a new one")
		}
		return handle(Unit{tx})
	})
}
