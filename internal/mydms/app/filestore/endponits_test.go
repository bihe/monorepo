package filestore_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.binggl.net/monorepo/internal/mydms/app/filestore"
)

type mockFileService struct {
	fail bool
}

func (m *mockFileService) InitClient() (err error) {
	return nil
}

func (m *mockFileService) SaveFile(file filestore.FileItem) (err error) {
	return nil
}

func (m *mockFileService) GetFile(filePath string) (item filestore.FileItem, err error) {
	if m.fail == true {
		return filestore.FileItem{}, fmt.Errorf("error")
	}
	return filestore.FileItem{
		MimeType: "application/pdf",
		Payload:  []byte{0},
	}, nil
}

func (m *mockFileService) DeleteFile(filePath string) (err error) {
	return nil
}

func Test_Endpoint(t *testing.T) {
	e := filestore.MakeGetFileEndpoint(&mockFileService{})
	if e == nil {
		t.Errorf("could not crate endpoint")
	}

	r, err := e(context.TODO(), filestore.GetFilePathRequest{Path: "/path"})
	if err != nil {
		t.Errorf("could not get response from endpoint; %v", err)
	}
	resp, ok := r.(filestore.GetFilePathResponse)
	if !ok {
		t.Error("response is not of type 'filestore.GetFilePathResponse'")
	}
	assert.Equal(t, "application/pdf", resp.MimeType)

	// failed
	e = filestore.MakeGetFileEndpoint(&mockFileService{fail: true})
	if e == nil {
		t.Errorf("could not crate endpoint")
	}

	r, err = e(context.TODO(), filestore.GetFilePathRequest{Path: "/path"})
	if err != nil {
		t.Errorf("could not get response from endpoint; %v", err)
	}
	resp, ok = r.(filestore.GetFilePathResponse)
	if !ok {
		t.Error("response is not of type 'filestore.GetFilePathResponse'")
	}
	if resp.Failed() == nil {
		t.Errorf("error expected")
	}

}
