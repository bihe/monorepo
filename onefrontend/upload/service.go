package upload

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.binggl.net/monorepo/crypter"
	"golang.binggl.net/monorepo/pkg/logging"

	log "github.com/sirupsen/logrus"
)

// --------------------------------------------------------------------------
// Type definition
// --------------------------------------------------------------------------

// File provides a reader for the payload and meta-data
type File struct {
	File     io.Reader
	Name     string
	Size     int64
	MimeType string
	Enc      EncryptionRequest
}

// EncryptionRequest contains parammeters used for encryption
type EncryptionRequest struct {
	InitPassword string
	Password     string
	Token        string
}

// --------------------------------------------------------------------------
// Interface definition
// --------------------------------------------------------------------------

// Service contains the main logic of the upload package
// the Service takes care of saving, reading and deleting uploaded files
type Service interface {
	Save(file File) (string, error)
	Read(id string) (Upload, error)
	Delete(id string) error
}

// ServiceOptions defines parameters used to initialize a new Service
type ServiceOptions struct {
	Logger           *log.Entry
	Store            Store
	MaxUploadSize    int64
	AllowedFileTypes []string
	// the Encryptionservice which is used to optionally encrypt the payload / change password of encrypted payload
	Crypter crypter.EncryptionService
	TimeOut string
}

// NewService creates a new Service instance
func NewService(init ServiceOptions) Service {
	duration := 0 * time.Second
	if init.TimeOut != "" {
		duration = parseDuration(init.TimeOut)
	}
	return &uploadService{
		logger:           init.Logger,
		store:            init.Store,
		maxUploadSize:    init.MaxUploadSize,
		allowedFileTypes: init.AllowedFileTypes,
		crypter:          init.Crypter,
		timeOut:          duration,
	}
}

// --------------------------------------------------------------------------
// Errors
// --------------------------------------------------------------------------

var (
	// ErrInvalidParameters tells the caller that incorrect, invalid parameters were supplied
	ErrInvalidParameters = errors.New("invalid parameters supplied")
	// ErrValidation tells the caller that the process cannot proceed because of validation errors
	ErrValidation = errors.New("validation error")
	// ErrService tells the caller that an application error happened
	ErrService = errors.New("application error occurred")
)

// --------------------------------------------------------------------------
// Implementation
// --------------------------------------------------------------------------

type uploadService struct {
	store            Store
	logger           *log.Entry
	maxUploadSize    int64
	allowedFileTypes []string
	crypter          crypter.EncryptionService
	timeOut          time.Duration
}

// compile time check if all methods of Service are implemented in the uploadService
var _ Service = &uploadService{}

func (s *uploadService) Save(file File) (string, error) {
	var (
		id      string
		payload []byte
		err     error
	)
	if file.Size > s.maxUploadSize {
		return id, fmt.Errorf("the upload exceeds the maximum size of %d - filesize is: %d; %w", s.maxUploadSize, file.Size, ErrValidation)
	}

	logging.LogWith(s.logger, "upload.Save").Debugf("trying to upload file: '%s'", file.Name)

	ext := filepath.Ext(file.Name)
	ext = strings.Replace(ext, ".", "", 1)
	var typeAllowed = false
	for _, t := range s.allowedFileTypes {
		if t == ext {
			typeAllowed = true
			break
		}
	}
	if !typeAllowed {
		return id, fmt.Errorf("the uploaded file-type '%s' is not allowed, only use: '%s'; %w", ext, strings.Join(s.allowedFileTypes, ","), ErrValidation)
	}

	// Copy
	b := &bytes.Buffer{}
	if _, err = io.Copy(b, file.File); err != nil {
		logging.LogWith(s.logger, "upload.Save").Errorf("could not copy file: %v", err)
		return id, ErrService
	}
	payload = b.Bytes()

	// optional encryption
	// we only try to encrypt something, if the crypter is initialized and a password is supplied
	if s.crypter != nil && file.Enc.Password != "" {
		ctxt, cancel := context.WithTimeout(context.Background(), s.timeOut)
		defer cancel()

		payload, err = s.crypter.Encrypt(ctxt, crypter.Request{
			AuthToken: file.Enc.Token,
			InitPass:  file.Enc.InitPassword,
			NewPass:   file.Enc.Password,
			Type:      crypter.PDF, // only encrypt PDFs for now
			Payload:   payload,
		})
		if err != nil {
			logging.LogWith(s.logger, "upload.Save").Errorf("could not encrypt file: %v", err)
			return id, fmt.Errorf("could not encrypt payload, %w", ErrService)
		}
	}

	id = uuid.New().String()
	u := Upload{
		ID:       id,
		FileName: file.Name,
		MimeType: file.MimeType,
		Payload:  payload,
		Created:  time.Now().UTC(),
	}
	if err = s.store.Write(u); err != nil {
		logging.LogWith(s.logger, "upload.Save").Errorf("could not save upload file: %v", err)
		return id, ErrService
	}
	return id, nil
}

func (s *uploadService) Read(id string) (Upload, error) {
	var item Upload
	if id == "" {
		return item, fmt.Errorf("invalid or empty id supplied '%v'; %w", id, ErrInvalidParameters)
	}

	logging.LogWith(s.logger, "upload.Read").Debugf("get file by ID: '%s'", id)

	item, err := s.store.Read(id)
	if err != nil {
		logging.LogWith(s.logger, "upload.Read").Errorf("cannot get item by id '%s': %v", id, err)
		return item, fmt.Errorf("cannot get item by id '%s'", id)
	}
	return item, nil
}

func (s *uploadService) Delete(id string) error {
	if id == "" {
		return fmt.Errorf("invalid or empty id supplied '%v'; %w", id, ErrInvalidParameters)
	}

	logging.LogWith(s.logger, "upload.Delete").Debugf("delete file by ID: '%s'", id)

	err := s.store.Delete(id)
	if err != nil {
		logging.LogWith(s.logger, "upload.Delete").Errorf("cannot delete item by id '%s': %v", id, err)
		return fmt.Errorf("cannot delete item by id '%s'", id)
	}
	return nil
}

func parseDuration(duration string) time.Duration {
	d, err := time.ParseDuration(duration)
	if err != nil {
		panic(fmt.Sprintf("wrong value, cannot parse duration: %v", err))
	}
	return d
}
