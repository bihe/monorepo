package filestore

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"golang.binggl.net/monorepo/pkg/logging"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
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

func (f FileItem) String() string {
	return fmt.Sprintf("%s/%s", f.FolderName, f.FileName)
}

// FileService defines an interface for backend file services
type FileService interface {
	InitClient() (err error)
	SaveFile(file FileItem) (err error)
	GetFile(filePath string) (item FileItem, err error)
	DeleteFile(filePath string) (err error)
}

// S3Config defines the parameters to interact with S3 storage
type S3Config struct {
	Region   string
	Bucket   string
	Key      string
	Secret   string
	EndPoint string
}

// NewService returns a new instance of the fileservice
func NewService(ctx context.Context, logger logging.Logger, config S3Config) FileService {
	var svc FileService
	{
		svc = &s3service{config: config, logger: logger, ctx: ctx}
		svc = ServiceLoggingMiddleware(logger)(svc)
	}
	return svc
}

// --------------------------------------------------------------------------
// interface implementation
// --------------------------------------------------------------------------

// define an interface to use it for mocking
// https://docs.aws.amazon.com/sdk-for-go/v2/developer-guide/migrate-gosdk.html#mocking-and-iface
type s3ServiceClient interface {
	GetObject(context.Context, *s3.GetObjectInput, ...func(*s3.Options)) (*s3.GetObjectOutput, error)
	PutObject(context.Context, *s3.PutObjectInput, ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	DeleteObject(context.Context, *s3.DeleteObjectInput, ...func(*s3.Options)) (*s3.DeleteObjectOutput, error)
}

type s3service struct {
	ctx      context.Context
	logger   logging.Logger
	config   S3Config
	s3client s3ServiceClient
}

// InitClient determines if the backend store client was already initialized
// if it is not initialized it creates a new client using the supplied config
func (s *s3service) InitClient() (err error) {
	if s.s3client == nil {
		cfg, err := config.LoadDefaultConfig(s.ctx)
		if err != nil {
			return fmt.Errorf("could not start a new S3 session. %v", err)
		}
		var (
			client       *s3.Client
			endpoint     string
			usePathStyle bool
		)
		if s.config.EndPoint != "" && strings.HasPrefix(s.config.EndPoint, "http") {
			endpoint = s.config.EndPoint
			usePathStyle = true
			s.logger.Debug(fmt.Sprintf("s3 config: endpoint=%s (forcePathStyle=%t)", s.config.EndPoint, usePathStyle))
		}

		client = s3.NewFromConfig(cfg, func(o *s3.Options) {
			o.Region = s.config.Region
			if endpoint != "" {
				o.UsePathStyle = usePathStyle
				o.BaseEndpoint = &endpoint
			}
			o.Credentials = credentials.NewStaticCredentialsProvider(s.config.Key, s.config.Secret, "")
		})

		s.s3client = client
	}
	return nil
}

// GetFile retrieves a file defined by a given path from the backend store
func (s *s3service) GetFile(filePath string) (item FileItem, err error) {
	err = s.InitClient()
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

	s3obj, err := s.s3client.GetObject(s.ctx,
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
func (s *s3service) SaveFile(file FileItem) (err error) {
	err = s.InitClient()
	if err != nil {
		return err
	}

	fileSize := len(file.Payload)
	storagePath := fmt.Sprintf("%s/%s", file.FolderName, file.FileName)
	_, err = s.s3client.PutObject(s.ctx, &s3.PutObjectInput{
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
func (s *s3service) DeleteFile(filePath string) (err error) {
	err = s.InitClient()
	if err != nil {
		return err
	}

	_, err = s.s3client.DeleteObject(s.ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.config.Bucket),
		Key:    aws.String(filePath),
	})
	if err != nil {
		return fmt.Errorf("could not delete the file item '%s'. %v", filePath, err)
	}
	return nil
}
