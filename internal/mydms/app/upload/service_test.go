package upload_test

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"testing"

	"golang.binggl.net/monorepo/internal/mydms/app/crypter"
	"golang.binggl.net/monorepo/internal/mydms/app/upload"
	"golang.binggl.net/monorepo/pkg/logging"

	"github.com/stretchr/testify/assert"
)

var logger = logging.NewNop()

const unencryptedPDF = "../../../../testdata/unencrypted.pdf"

// --------------------------------------------------------------------------

type mockStore struct {
	fail   bool
	upload upload.Upload
}

func (m *mockStore) Write(item upload.Upload) (err error) {
	if m.fail {
		return fmt.Errorf("error")
	}
	m.upload = item
	return nil
}

func (m *mockStore) Read(id string) (upload.Upload, error) {
	if m.fail {
		return upload.Upload{}, fmt.Errorf("error")
	}
	if id == m.upload.ID {
		return m.upload, nil
	}
	return upload.Upload{}, fmt.Errorf("error")
}

func (m *mockStore) Delete(id string) (err error) {
	if m.fail {
		return fmt.Errorf("error")
	}
	m.upload = upload.Upload{}
	return nil
}

var _ upload.Store = &mockStore{}

// --------------------------------------------------------------------------

type mockEncService struct {
	fail bool
}

func (m *mockEncService) Encrypt(ctx context.Context, req crypter.Request) ([]byte, error) {
	if m.fail {
		return nil, fmt.Errorf("error")
	}
	return nil, nil
}

var _ crypter.EncryptionService = &mockEncService{}

// --------------------------------------------------------------------------

const maxUploadSize = 5000000

func TestService_FileType_CaseInsensitive(t *testing.T) {
	svc := upload.NewService(upload.ServiceOptions{
		Logger:           logger,
		Store:            &mockStore{},
		MaxUploadSize:    maxUploadSize,
		AllowedFileTypes: []string{"pdf"},
	})
	payload, err := os.ReadFile(unencryptedPDF)
	if err != nil {
		t.Fatalf("could not read testfile: %v", err)
	}
	var b bytes.Buffer // A Buffer needs no initialization.
	if _, err := b.Write(payload); err != nil {
		t.Fatalf("could not write payload to buffer: %v", err)
	}

	// check that the supported file-types is case insensitive
	var id string
	if id, err = svc.Save(upload.File{
		File:     &b,
		MimeType: "application/pdf",
		Name:     "unencrypted.PDF",
		Size:     int64(len(payload)),
	}); err != nil {
		t.Fatalf("could not write file: %v", err)
	}
	assert.True(t, id != "")
}

func TestService_Write_Read_Delete(t *testing.T) {
	svc := upload.NewService(upload.ServiceOptions{
		Logger:           logger,
		Store:            &mockStore{},
		MaxUploadSize:    maxUploadSize,
		AllowedFileTypes: []string{"pdf", "png"},
	})

	payload, err := os.ReadFile(unencryptedPDF)
	if err != nil {
		t.Fatalf("could not read testfile: %v", err)
	}
	var b bytes.Buffer // A Buffer needs no initialization.
	if _, err := b.Write(payload); err != nil {
		t.Fatalf("could not write payload to buffer: %v", err)
	}

	// standard Save
	var id string
	if id, err = svc.Save(upload.File{
		File:     &b,
		MimeType: "application/pdf",
		Name:     "unencrypted.pdf",
		Size:     int64(len(payload)),
	}); err != nil {
		t.Errorf("could not write file: %v", err)
	}
	assert.True(t, id != "")

	// read the entry
	u, err := svc.Read(id)
	if err != nil {
		t.Errorf("could not read file: %v", err)
	}
	assert.Equal(t, "unencrypted.pdf", u.FileName)
	assert.Equal(t, "application/pdf", u.MimeType)
	assert.Equal(t, id, u.ID)
	assert.Equal(t, len(payload), len(u.Payload))

	// finally - delete the entry
	err = svc.Delete(id)
	if err != nil {
		t.Errorf("could not delete file: %v", err)
	}

	// some error-situations

	err = svc.Delete("")
	if err == nil {
		t.Error("error expected")
	}

	_, err = svc.Read("")
	if err == nil {
		t.Error("error expected")
	}

	_, err = svc.Read("unknown")
	if err == nil {
		t.Error("error expected")
	}

	// payload to big
	_, err = svc.Save(upload.File{
		File:     &b,
		MimeType: "application/pdf",
		Name:     "unencrypted.pdf",
		Size:     100000000,
	})
	if err == nil {
		t.Error("error expected")
	}

	// invalid payload - filetype
	_, err = svc.Save(upload.File{
		File:     &b,
		MimeType: "application/noidea",
		Name:     "unencrypted.noidea",
		Size:     100,
	})
	if err == nil {
		t.Error("error expected")
	}
}

func TestService_Write_Encrypt(t *testing.T) {
	svc := upload.NewService(upload.ServiceOptions{
		Logger:           logger,
		Store:            &mockStore{},
		MaxUploadSize:    maxUploadSize,
		AllowedFileTypes: []string{"pdf", "png"},
		Crypter:          &mockEncService{},
		TimeOut:          "10s",
	})

	payload, err := os.ReadFile(unencryptedPDF)
	if err != nil {
		t.Fatalf("could not read testfile: %v", err)
	}
	var b bytes.Buffer // A Buffer needs no initialization.
	if _, err := b.Write(payload); err != nil {
		t.Fatalf("could not write payload to buffer: %v", err)
	}

	// encrypted Save
	var id string
	if id, err = svc.Save(upload.File{
		File:     &b,
		MimeType: "application/pdf",
		Name:     "unencrypted.pdf",
		Size:     int64(len(payload)),
		Enc: upload.EncryptionRequest{
			Password: "12345",
		},
	}); err != nil {
		t.Errorf("could not write file: %v", err)
	}
	assert.True(t, id != "")

	// application-error
	svc = upload.NewService(upload.ServiceOptions{
		Logger:           logger,
		Store:            &mockStore{},
		MaxUploadSize:    10000,
		AllowedFileTypes: []string{"pdf", "png"},
		Crypter: &mockEncService{
			fail: true,
		},
		TimeOut: "10s",
	})

	if _, err = svc.Save(upload.File{
		File:     &b,
		MimeType: "application/pdf",
		Name:     "unencrypted.pdf",
		Size:     int64(len(payload)),
		Enc: upload.EncryptionRequest{
			Password: "12345",
		},
	}); err == nil {
		t.Error("error expected")
	}
}

func Test_Service_Errors_Store(t *testing.T) {
	svc := upload.NewService(upload.ServiceOptions{
		Logger:           logger,
		Store:            &mockStore{fail: true},
		MaxUploadSize:    maxUploadSize,
		AllowedFileTypes: []string{"pdf", "png"},
	})

	payload, err := os.ReadFile(unencryptedPDF)
	if err != nil {
		t.Fatalf("could not read testfile: %v", err)
	}
	var b bytes.Buffer // A Buffer needs no initialization.
	if _, err := b.Write(payload); err != nil {
		t.Fatalf("could not write payload to buffer: %v", err)
	}

	_, err = svc.Save(upload.File{
		File:     &b,
		MimeType: "application/pdf",
		Name:     "unencrypted.pdf",
		Size:     10000,
	})
	if err == nil {
		t.Error("error expected")
	}

	_, err = svc.Read("id")
	if err == nil {
		t.Error("error expected")
	}

	err = svc.Delete("id")
	if err == nil {
		t.Error("error expected")
	}
}
