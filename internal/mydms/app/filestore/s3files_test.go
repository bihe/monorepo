package filestore

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/stretchr/testify/assert"
	"golang.binggl.net/monorepo/pkg/logging"
)

// rather small PDF payload
// https://stackoverflow.com/questions/17279712/what-is-the-smallest-possible-valid-pdf
const pdfPayload = `%PDF-1.0
1 0 obj<</Type/Catalog/Pages 2 0 R>>endobj 2 0 obj<</Type/Pages/Kids[3 0 R]/Count 1>>endobj 3 0 obj<</Type/Page/MediaBox[0 0 3 3]>>endobj
xref
0 4
0000000000 65535 f
0000000010 00000 n
0000000053 00000 n
0000000102 00000 n
trailer<</Size 4/Root 1 0 R>>
startxref
149
%EOF
`
const mimeType = "application/pdf"

var logger = logging.NewNop()

// Define a mock struct to be used in your unit tests of myFunc.
// https://github.com/aws/aws-sdk-go/blob/master/service/s3/s3iface/interface.go
type mockS3Client struct {
	s3iface.S3API
}

func (m *mockS3Client) GetObject(input *s3.GetObjectInput) (*s3.GetObjectOutput, error) {
	if *input.Key == "" || *input.Key == "null/null" {
		return nil, fmt.Errorf("could not get object with Key %s", *input.Key)
	}
	return &s3.GetObjectOutput{
		ContentType: aws.String(mimeType),
		Body:        ioutil.NopCloser(bytes.NewReader([]byte(pdfPayload))),
	}, nil
}

func (m *mockS3Client) PutObject(input *s3.PutObjectInput) (*s3.PutObjectOutput, error) {
	if *input.Key == "/" {
		return nil, fmt.Errorf("could not upload object with Key %s", *input.Key)
	}
	return &s3.PutObjectOutput{}, nil
}

func (m *mockS3Client) DeleteObject(input *s3.DeleteObjectInput) (*s3.DeleteObjectOutput, error) {
	if *input.Key == "/" {
		return nil, fmt.Errorf("could not delete object with Key %s", *input.Key)
	}
	return &s3.DeleteObjectOutput{}, nil
}

func TestInitClient(t *testing.T) {
	svc := NewService(logger, S3Config{})
	err := svc.InitClient()
	if err != nil {
		t.Errorf("error initializing client")
	}
}

func TestGetS3Entry(t *testing.T) {
	service := s3service{
		config: S3Config{},
		client: &mockS3Client{},
	}

	cases := []struct {
		Name     string
		Path     string
		Expected FileItem
	}{
		{
			Name: "valid file",
			Path: "2009_08_06/20090806-invoice.pdf",
			Expected: FileItem{
				FolderName: "2009_08_06",
				FileName:   "20090806-invoice.pdf",
				MimeType:   mimeType,
				Payload:    []byte(pdfPayload),
			},
		},
		{
			Name: "valid file",
			Path: "/2009_08_06/20090806-invoice.pdf",
			Expected: FileItem{
				FolderName: "2009_08_06",
				FileName:   "20090806-invoice.pdf",
				MimeType:   mimeType,
				Payload:    []byte(pdfPayload),
			},
		},
		{
			Name:     "invalid path",
			Path:     "",
			Expected: FileItem{},
		},
		{
			Name:     "invalid file",
			Path:     "null/null",
			Expected: FileItem{},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			fileItem, err := service.GetFile(c.Path)
			if c.Name == "invalid path" || c.Name == "invalid file" {
				assert.Error(t, err, "error for invalid file expected!")
				return
			}
			assert.NoErrorf(t, err, "could not get file from s3 backend: %v", err)
			assert.True(t, fileItem.FolderName == c.Expected.FolderName && fileItem.FileName == c.Expected.FileName, "could not get the correct file-metadata from the backend!")
			assert.Equal(t, len(fileItem.Payload), len(c.Expected.Payload), "incorrect payload size returned!")
		})
	}

}

func TestSaveS3Entry(t *testing.T) {
	service := s3service{
		config: S3Config{},
		client: &mockS3Client{},
	}
	err := service.SaveFile(FileItem{
		FileName:   "test.pdf",
		FolderName: "__TEST",
		MimeType:   mimeType,
		Payload:    []byte(pdfPayload),
	})
	if err != nil {
		t.Errorf("could not get upload file to s3 backend: %v", err)
	}

	err = service.SaveFile(FileItem{
		FileName:   "",
		FolderName: "",
		MimeType:   mimeType,
		Payload:    []byte(pdfPayload),
	})
	if err == nil {
		t.Errorf("expected error invalid path")
	}
}

func TestDeleteS3Entry(t *testing.T) {
	service := s3service{
		config: S3Config{},
		client: &mockS3Client{},
	}
	err := service.DeleteFile("test.pdf")
	if err != nil {
		t.Errorf("could not get delete file from s3 backend: %v", err)
	}

	err = service.DeleteFile("/")
	if err == nil {
		t.Errorf("expected error invalid path")
	}
}
