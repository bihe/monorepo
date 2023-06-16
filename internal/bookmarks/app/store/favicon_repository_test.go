package store_test

import (
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.binggl.net/monorepo/internal/bookmarks/app"
	"golang.binggl.net/monorepo/internal/bookmarks/app/store"
	"golang.binggl.net/monorepo/pkg/logging"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var logger = logging.NewNop()

func favRepo(t *testing.T) (store.FaviconRepository, *sql.DB) {
	var (
		DB  *gorm.DB
		err error
	)
	if DB, err = gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{}); err != nil {
		t.Fatalf("cannot create database connection: %v", err)
	}
	// Migrate the schema
	DB.AutoMigrate(&store.Favicon{})
	db, err := DB.DB()
	if err != nil {
		t.Fatalf("could not get DB handle; %v", err)
	}
	return store.CreateFaviconRepo(DB, logger), db
}

func Test_CRUD_Favicon(t *testing.T) {
	repo, db := favRepo(t)
	defer db.Close()

	// create

	fav, err := repo.Save(store.Favicon{
		ID:           "favicon_id",
		Payload:      app.DefaultFavicon,
		LastModified: time.Now(),
	})
	if err != nil {
		t.Errorf("could not save favicon; %v", err)
	}

	if fav.ID != "favicon_id" || len(fav.Payload) != len(app.DefaultFavicon) {
		t.Errorf("the returned item is not valid")
	}

	// get

	fav, err = repo.Get("favicon_id")
	if err != nil {
		t.Errorf("could not get favicon by id: %v", err)
	}

	assert.Equal(t, "favicon_id", fav.ID)
	assert.Equal(t, len(app.DefaultFavicon), len(fav.Payload))
	assert.True(t, fav.LastModified.Before(time.Now()))

	_, err = repo.Get("favicon_id_not_found")
	if err == nil {
		t.Errorf("expected error for unknown id")
	}

	// update
	fav.Payload = make([]byte, 1)
	fav, err = repo.Save(fav)
	if err != nil {
		t.Errorf("could not update the favicon; %v", err)
	}

	assert.Equal(t, 1, len(fav.Payload))

	// delete
	err = repo.Delete(fav)
	if err != nil {
		t.Errorf("could not delete the favicon; %v", err)
	}

	_, err = repo.Get(fav.ID)
	if err == nil {
		t.Error("expected an error because of deleted favicon")
	}

	// validation

	_, err = repo.Save(store.Favicon{
		ID:           "id",
		Payload:      make([]byte, 0),
		LastModified: time.Now(),
	})
	if err == nil {
		t.Errorf("expected error for missing payload")
	}

	_, err = repo.Save(store.Favicon{
		ID:           "",
		Payload:      make([]byte, 0),
		LastModified: time.Now(),
	})
	if err == nil {
		t.Errorf("expected error for missing id")
	}

	err = repo.Delete(store.Favicon{
		ID:           "",
		Payload:      make([]byte, 1),
		LastModified: time.Now(),
	})
	if err == nil {
		t.Errorf("expected error for missing id")
	}
}

func Test_InUnitOfWork(t *testing.T) {
	repo, db := favRepo(t)
	defer db.Close()

	// create
	err := repo.InUnitOfWork(func(repo store.FaviconRepository) error {

		fav, err := repo.Save(store.Favicon{
			ID:           "favicon_id",
			Payload:      app.DefaultFavicon,
			LastModified: time.Now(),
		})
		if err != nil {
			return err
		}

		assert.Equal(t, "favicon_id", fav.ID)
		assert.Equal(t, len(app.DefaultFavicon), len(fav.Payload))
		assert.True(t, fav.LastModified.Before(time.Now()))

		return nil
	})
	if err != nil {
		t.Errorf("could not execute in UnitOfWork; %v", err)
	}

}
