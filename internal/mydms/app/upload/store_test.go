package upload_test

import (
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"golang.binggl.net/monorepo/internal/mydms/app/upload"
)

func TestStore_Write_Read_Delete(t *testing.T) {
	store := upload.NewStore(t.TempDir())
	payload, err := os.ReadFile(unencryptedPDF)
	if err != nil {
		t.Fatalf("could not read testdata: %v", err)
	}
	item := upload.Upload{
		FileName: "test.pdf",
		Created:  time.Now().UTC(),
		MimeType: "application/pdf",
		ID:       uuid.New().String(),
		Payload:  payload,
	}

	// Write
	if err := store.Write(item); err != nil {
		t.Errorf("could not write item to store: %v", err)
	}

	// Read
	readItem, err := store.Read(item.ID)
	if err != nil {
		t.Errorf("could not read item from store: %v", err)
	}
	assert.Equal(t, item.ID, readItem.ID)
	assert.Equal(t, "test.pdf", readItem.FileName)
	assert.Equal(t, "application/pdf", readItem.MimeType)
	assert.Equal(t, len(payload), len(readItem.Payload))

	// Delete
	if err = store.Delete(readItem.ID); err != nil {
		t.Errorf("could not delete item from store: %v", err)
	}
}

func Test_Store_Validation(t *testing.T) {
	store := upload.NewStore(t.TempDir())

	err := store.Write(upload.Upload{})
	if err == nil {
		t.Error("error expected")
	}

	_, err = store.Read("")
	if err == nil {
		t.Error("error expected")
	}

	err = store.Delete("")
	if err == nil {
		t.Error("error expected")
	}
}

func Test_Store_Invalid_BasePath(t *testing.T) {
	store := upload.NewStore("/__")

	err := store.Write(upload.Upload{
		ID: "1",
	})
	if err == nil {
		t.Error("error expected")
	}

	_, err = store.Read("1")
	if err == nil {
		t.Error("error expected")
	}

	err = store.Delete("1")
	if err == nil {
		t.Error("error expected")
	}
}
