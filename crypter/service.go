package crypter

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/go-kit/kit/log"
	pdfApi "github.com/pdfcpu/pdfcpu/pkg/api"
	pdf "github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
	"golang.binggl.net/monorepo/pkg/security"
)

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
func NewService(logger log.Logger, tokenSetting TokenSecurity) EncryptionService {
	var useCache = tokenSetting.CacheDuration != ""
	var svc EncryptionService
	{
		svc = &encryptionSvc{
			logger: logger,
			jwtAuth: security.NewJWTAuthorization(security.JwtOptions{
				CacheDuration: tokenSetting.CacheDuration,
				JwtIssuer:     tokenSetting.JwtIssuer,
				JwtSecret:     tokenSetting.JwtSecret,
				RequiredClaim: security.Claim{
					Name:  tokenSetting.Claim.Name,
					Roles: tokenSetting.Claim.Roles,
					URL:   tokenSetting.Claim.URL,
				},
			}, useCache),
		}
		svc = ServiceLoggingMiddleware(logger)(svc)
	}
	return svc
}

// --------------------------------------------------------------------------
// Service Middleware
// --------------------------------------------------------------------------

// ServiceMiddleware describes a service (as opposed to endpoint) middleware.
// it is used to intercept the method execution and perform actions before/after the
// serivce method execution
type ServiceMiddleware func(EncryptionService) EncryptionService

// ServiceLoggingMiddleware takes a logger as a dependency
// and returns a ServiceLoggingMiddleware.
func ServiceLoggingMiddleware(logger log.Logger) ServiceMiddleware {
	return func(next EncryptionService) EncryptionService {
		return loggingMiddleware{logger, next}
	}
}

type loggingMiddleware struct {
	logger log.Logger
	next   EncryptionService
}

// Encrypt implements the Service so that it can be used transparently
// the provided parameters are logged and the execution is passed on to the underlying/next service
func (mw loggingMiddleware) Encrypt(ctx context.Context, req Request) (encPayload []byte, err error) {
	defer func() {
		mw.logger.Log("method", "Encrypt", "payload(length)", len(req.Payload), "payloadType", req.Type, "encPayload(lenght)", len(encPayload), "err", err)
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
	logger  log.Logger
	jwtAuth *security.JWTAuthorization
}

func (e *encryptionSvc) Encrypt(ctx context.Context, req Request) ([]byte, error) {
	err := validateEncryptionRequest(req.AuthToken, req.NewPass, req.Payload)
	if err != nil {
		e.logger.Log("method", "Encrypt", "err", fmt.Errorf("validation arguments error: %v", err))
		return nil, err
	}
	user, err := e.jwtAuth.EvaluateToken(req.AuthToken)
	if err != nil {
		e.logger.Log("method", "Encrypt", "err", fmt.Errorf("error during token evaluation: %v", err))
		return nil, err
	}
	e.logger.Log("method", "Encrypt", "user", user.UserID, "msg", fmt.Errorf("user '%s' requests encryption of payload", user.UserID))

	// encrypt the payload per respective type
	switch req.Type {
	case PDF:
		return encryptPdfPayload(req.Payload, req.InitPass, req.NewPass)
	}
	return nil, ErrInvlaidPayloadType
}

// check for provided values
func validateEncryptionRequest(authToken, newPwd string, payload []byte) error {
	if authToken == "" {
		return fmt.Errorf("no authentication token supplied")
	}
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
		conf *pdf.Configuration
		err  error
	)

	conf = pdf.NewRC4Configuration(initPass, initPass, 40)
	conf.Cmd = pdf.CHANGEUPW
	conf.UserPW = initPass
	conf.UserPWNew = &newPass

	// if no initial password is supplied the api does not change the
	// password but sets a password via the ENCRYPT command
	if initPass == "" && newPass != "" {
		conf = pdf.NewRC4Configuration(newPass, newPass, 40)
		conf.Cmd = pdf.ENCRYPT
		conf.UserPWNew = &newPass
	}

	buff := new(bytes.Buffer)
	if err = pdfApi.Optimize(bytes.NewReader(payload), buff, conf); err != nil {
		return nil, fmt.Errorf("could not process PDF payload: %v", err)
	}

	return buff.Bytes(), nil
}
