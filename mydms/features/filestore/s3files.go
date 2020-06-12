package filestore

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
)

// --------------------------------------------------------------------------
// type definitions
// --------------------------------------------------------------------------

// FileItem represents the data-structure of a saved file
type FileItem struct {
	FileName   string
	FolderName string
	MimeType   string
	Payload    []byte
}

// FileService defines an interface for backend file services
type FileService interface {
	SaveFile(file FileItem) error
	GetFile(filePath string) (FileItem, error)
	DeleteFile(filePath string) error
}

// --------------------------------------------------------------------------
// interface implementation
// --------------------------------------------------------------------------

// S3Config defines the parameters to interact with S3 storage
type S3Config struct {
	Region string
	Bucket string
	Key    string
	Secret string
}

// NewService returns a new instance of the fileservice
func NewService(config S3Config) FileService {
	return &s3service{config: config}
}

type s3service struct {
	config S3Config
	client s3iface.S3API
}

// InitClient determines if the backend store client was already initialized
// if it is not initilized it creates a new client using the supplied config
func (s *s3service) InitClient() error {
	if s.client == nil {
		sess, err := session.NewSession(&aws.Config{
			Region:      aws.String(s.config.Region),
			Credentials: credentials.NewStaticCredentials(s.config.Key, s.config.Secret, ""),
		},
		)
		if err != nil {
			return fmt.Errorf("could not start a new S3 session. %v", err)
		}
		s.client = s3.New(sess)
	}
	return nil
}

// GetFile retrieves a file defined by a given path from the backend store
func (s *s3service) GetFile(filePath string) (FileItem, error) {
	err := s.InitClient()
	if err != nil {
		return FileItem{}, err
	}

	fileURLPath := filePath
	if strings.Index(fileURLPath, "/") == 0 {
		fileURLPath = fileURLPath[1:]
	}
	parts := strings.Split(fileURLPath, "/")
	if len(parts) != 2 {
		return FileItem{}, fmt.Errorf("invalid path supplied: %s", fileURLPath)
	}
	path := parts[0]
	fileName := parts[1]

	s3obj, err := s.client.GetObject(
		&s3.GetObjectInput{
			Bucket: aws.String(s.config.Bucket),
			Key:    aws.String(fileURLPath),
		})
	if err != nil {
		return FileItem{}, fmt.Errorf("could not get object %s/%s. %v", s.config.Bucket, fileURLPath, err)
	}
	ctype := s3obj.ContentType
	buf := new(bytes.Buffer)
	buf.ReadFrom(s3obj.Body)
	payload := buf.Bytes()
	s3obj.Body.Close()

	return FileItem{
		FileName:   fileName,
		FolderName: path,
		MimeType:   *ctype,
		Payload:    payload,
	}, nil
}

// SaveFile stores a file item using a given path to the backend store
func (s *s3service) SaveFile(file FileItem) error {
	err := s.InitClient()
	if err != nil {
		return err
	}

	fileSize := len(file.Payload)
	storagePath := fmt.Sprintf("%s/%s", file.FolderName, file.FileName)
	_, err = s.client.PutObject(&s3.PutObjectInput{
		Bucket:        aws.String(s.config.Bucket),
		Key:           aws.String(storagePath),
		Body:          bytes.NewReader(file.Payload),
		ContentLength: aws.Int64(int64(fileSize)),
		ContentType:   aws.String(file.MimeType),
	})
	if err != nil {
		return fmt.Errorf("could not upload file item '%s' to S3 storage. %v", storagePath, err)
	}
	return nil
}

// DeleteFile removes the item using the specified paht
func (s *s3service) DeleteFile(filePath string) error {
	err := s.InitClient()
	if err != nil {
		return err
	}

	_, err = s.client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(s.config.Bucket),
		Key:    aws.String(filePath),
	})
	if err != nil {
		return fmt.Errorf("could not delete the file item '%s'. %v", filePath, err)
	}
	return nil
}
