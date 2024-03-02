// Package crypter implements an encryption services.
// the main purpose is to encrypt files / binary payloads.
// initial implementation focuses on PDF files
package crypter

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	pdfApi "github.com/pdfcpu/pdfcpu/pkg/api"
	pdfModel "github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"golang.binggl.net/monorepo/pkg/logging"
)

// --------------------------------------------------------------------------
// Types definition
// --------------------------------------------------------------------------

// PayloadType enumerates the available payload types
type PayloadType string

const (
	// PDF is a payload of application/pdf
	PDF PayloadType = "PDF"
)

// Request combines the parameters passed on to the encryption service
type Request struct {
	Payload  []byte
	Type     PayloadType
	InitPass string
	NewPass  string
}

// --------------------------------------------------------------------------
// EncryptionService definition
// --------------------------------------------------------------------------

// EncryptionService describes a service that encrypts payloads
type EncryptionService interface {
	// Encrypt uses the supplied token to validate the user and encrypts the provided payload
	// The encryption is performend using the supplied passwords.
	Encrypt(ctx context.Context, req Request) ([]byte, error)
}

var (
	// ErrInvlaidPayloadType tells the caller that encryption is not possible fo the supplied payload-type
	ErrInvlaidPayloadType = errors.New("invalid payload type supplied")
)

// NewService returns a Service with all of the expected middlewares wired in.
func NewService(logger logging.Logger) EncryptionService {
	var svc EncryptionService
	{
		svc = &encryptionSvc{
			logger: logger,
		}
		svc = ServiceLoggingMiddleware(logger)(svc)
	}
	return svc
}

// --------------------------------------------------------------------------
// Service Middleware
// --------------------------------------------------------------------------

// ServiceMiddleware describes a service middleware.
// it is used to intercept the method execution and perform actions before/after the
// serivce method execution
type ServiceMiddleware func(EncryptionService) EncryptionService

// ServiceLoggingMiddleware takes a logger as a dependency
// and returns a ServiceLoggingMiddleware.
func ServiceLoggingMiddleware(logger logging.Logger) ServiceMiddleware {
	return func(next EncryptionService) EncryptionService {
		return loggingMiddleware{logger, next}
	}
}

type loggingMiddleware struct {
	logger logging.Logger
	next   EncryptionService
}

// Encrypt implements the Service so that it can be used transparently
// the provided parameters are logged and the execution is passed on to the underlying/next service
func (mw loggingMiddleware) Encrypt(ctx context.Context, req Request) (encPayload []byte, err error) {
	defer func() {
		mw.logger.Info("Encrypt called",
			logging.LogV("payload(length)", fmt.Sprintf("%d", len(req.Payload))),
			logging.LogV("payloadType", string(req.Type)),
			logging.LogV("encPayload(lenght)", fmt.Sprintf("%d", len(encPayload))),
			logging.ErrV(err),
		)
	}()
	// pass on to the real service
	return mw.next.Encrypt(ctx, req)
}

// --------------------------------------------------------------------------
// Service implementation
// --------------------------------------------------------------------------

// compile time assertions for our response types implementing endpoint.Failer.
var (
	_ EncryptionService = &encryptionSvc{}
)

// implment the Service
type encryptionSvc struct {
	logger logging.Logger
}

// Encrypt performs the encryption operation for different file-types.
// currently only PDF encryption is possible.
// When deployed in a containerized environment the PDF library pdfcpu needs some configuration.
// This configuration can be done in a yaml file; the path to the file can be provided as an
// environement variable.
// For the current setup this is done in this way:
//
//	"XDG_CONFIG_HOME=/opt/mydms/etc"
//
// where the configuration file is located in the directory
//
//	pdfcpu/config.yml
//
// The configuration file is needed for a headless/server mode for pdfcpu: https://github.com/pdfcpu/pdfcpu/blob/873542105220142fae8e44e5a5a52a16cd391d2e/pkg/pdfcpu/model/resources/config.yml#L4
//
//	#########################
//	# Default configuration #
//	#########################
//
//	reader15: true
//	decodeAllStreams: false
//
//	# validationMode:
//	# ValidationStrict,
//	# ValidationRelaxed,
//	# ValidationNone
//	validationMode: ValidationRelaxed
//
//	# eol for writing:
//	# EolLF
//	# EolCR
//	# EolCRLF
//	eol: EolLF
//
//	writeObjectStream: true
//	writeXRefStream: true
//	encryptUsingAES: true
//
//	# encryptKeyLength: max 256
//	encryptKeyLength: 256
//
//	# permissions for encrypted files:
//	# -3901 = 0xF0C3 (PermissionsNone)
//	#    -1 = 0xFFFF (PermissionsAll)
//	permissions: -3901
//
//	# displayUnits:
//	# points
//	# inches
//	# cm
//	# mm
//	units: points
func (e *encryptionSvc) Encrypt(ctx context.Context, req Request) ([]byte, error) {
	err := validateEncryptionRequest(req.NewPass, req.Payload)
	if err != nil {
		e.logger.Error("Encrypt error", logging.ErrV(fmt.Errorf("validation arguments error: %v", err)))
		return nil, err
	}

	// encrypt the payload per respective type
	switch req.Type {
	case PDF:
		return encryptPdfPayload(req.Payload, req.InitPass, req.NewPass)
	}
	return nil, ErrInvlaidPayloadType
}

// check for provided values
func validateEncryptionRequest(newPwd string, payload []byte) error {
	if newPwd == "" {
		return fmt.Errorf("cannot encrypt with empty password")
	}
	if len(payload) == 0 {
		return fmt.Errorf("no payload supplied")
	}
	return nil
}

// encrypt pdf or change password of pdf
func encryptPdfPayload(payload []byte, initPass, newPass string) ([]byte, error) {
	var (
		conf *pdfModel.Configuration
		err  error
	)

	conf = pdfModel.NewRC4Configuration(initPass, initPass, 40)
	conf.Cmd = pdfModel.CHANGEUPW
	conf.UserPW = initPass
	conf.UserPWNew = &newPass
	conf.OwnerPWNew = &newPass

	// if no initial password is supplied the api does not change the
	// password but sets a password via the ENCRYPT command
	if initPass == "" && newPass != "" {
		conf = pdfModel.NewRC4Configuration(newPass, newPass, 40)
		conf.Cmd = pdfModel.ENCRYPT
		conf.UserPWNew = &newPass
	}

	buff := new(bytes.Buffer)
	if err = pdfApi.Optimize(bytes.NewReader(payload), buff, conf); err != nil {
		return nil, fmt.Errorf("could not process PDF payload: %v", err)
	}

	return buff.Bytes(), nil
}
