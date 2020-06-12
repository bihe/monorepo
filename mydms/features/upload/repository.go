package upload

import (
	"fmt"
	"time"

	"golang.binggl.net/monorepo/mydms/persistence"
)

// Upload defines an entity within the persistence store
type Upload struct {
	ID       string    `db:"id"`
	FileName string    `db:"filename"`
	MimeType string    `db:"mimetype"`
	Created  time.Time `db:"created"`
}

// Repository provides CRUD methods for uploads
type Repository interface {
	Write(item Upload, a persistence.Atomic) (err error)
	Read(id string) (Upload, error)
	Delete(id string, a persistence.Atomic) (err error)
}

type dbRepository struct {
	c persistence.Connection
}

// NewRepository creates a new instance using an existing connection
func NewRepository(c persistence.Connection) (Repository, error) {
	if !c.Active {
		return nil, fmt.Errorf("no repository connection available")
	}
	return &dbRepository{c}, nil
}

// Write saves an upload item
func (rw *dbRepository) Write(item Upload, a persistence.Atomic) (err error) {
	var atomic *persistence.Atomic

	defer func() {
		err = persistence.HandleTX(!a.Active, atomic, err)
	}()

	if atomic, err = persistence.CheckTX(rw.c, &a); err != nil {
		return
	}

	_, err = atomic.NamedExec("INSERT INTO UPLOADS (id,filename,mimetype,created) VALUES (:id, :filename, :mimetype, :created)", &item)
	if err != nil {
		err = fmt.Errorf("cannot write upload item: %v", err)
		return
	}
	return nil
}

// Read gets an item by it's ID
func (rw *dbRepository) Read(id string) (Upload, error) {
	u := Upload{}

	err := rw.c.Get(&u, "SELECT id, filename, mimetype, created FROM UPLOADS WHERE id=?", id)
	if err != nil {
		return Upload{}, fmt.Errorf("cannot get upload-item by id '%s': %v", id, err)
	}
	return u, nil
}

// Delete removes the item with the specified id from the store
func (rw *dbRepository) Delete(id string, a persistence.Atomic) (err error) {
	var atomic *persistence.Atomic

	defer func() {
		err = persistence.HandleTX(!a.Active, atomic, err)
	}()

	if atomic, err = persistence.CheckTX(rw.c, &a); err != nil {
		return
	}
	_, err = atomic.Exec("DELETE FROM UPLOADS WHERE id = ?", id)
	if err != nil {
		err = fmt.Errorf("cannot delete upload item: %v", err)
		return
	}
	return nil
}
