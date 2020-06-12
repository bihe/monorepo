package server

import (
	"bytes"
	"context"
	"fmt"

	pdfApi "github.com/pdfcpu/pdfcpu/pkg/api"
	pdf "github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
	"golang.binggl.net/monorepo/filecrypt"
	"golang.binggl.net/monorepo/pkg/logging"
	"golang.binggl.net/monorepo/pkg/security"
	"golang.binggl.net/monorepo/proto"

	log "github.com/sirupsen/logrus"
)

// FileCrypter implements the RPC service: Crypter
type FileCrypter struct {
	logger   *log.Entry
	basePath string
	jwtAuth  *security.JWTAuthorization
}

// NewFileCrypter creates a new FileCrypter iinstance
func NewFileCrypter(basePath string, logger *log.Entry, tokenSetting filecrypt.TokenSecurity) *FileCrypter {
	var useCache = tokenSetting.CacheDuration != ""
	return &FileCrypter{
		logger:   logger,
		basePath: basePath,
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
}

// check for provided values
func validateEncryptionRequest(req *proto.CrypterRequest) error {
	if req.AuthToken == "" {
		return fmt.Errorf("no authentication token supplied")
	}
	if req.NewPassword == "" {
		return fmt.Errorf("cannot encrypt with empty password")
	}
	if len(req.Payload) == 0 {
		return fmt.Errorf("no payload supplied")
	}
	return nil
}

// Encrypt takes a CrypterRequest and encrypts the payload according to the rules of the supplied data
// the result of the operation is a CrypterResponse
func (f *FileCrypter) Encrypt(ctxt context.Context, req *proto.CrypterRequest) (resp *proto.CrypterResponse, err error) {
	err = validateEncryptionRequest(req)
	if err != nil {
		logging.LogWith(f.logger, "server.Encrypt").Debugf("validation errors: %v", err)
		return nil, err
	}
	user, err := f.jwtAuth.EvaluateToken(req.AuthToken)
	if err != nil {
		logging.LogWith(f.logger, "server.Encrypt").Warnf("error during token evaluation: %v", err)
		return nil, err
	}
	logging.LogWith(f.logger, "server.Encrypt").Debugf("user '%s' requests encryption of payload", user.UserID)

	// encrypt the payload per respective type
	switch req.Type {
	case proto.CrypterRequest_PDF:
		resp, err = encryptPdfPayload(req)
		return
	}
	return resp, nil
}

// encrypt pdf or change password of pdf
func encryptPdfPayload(req *proto.CrypterRequest) (*proto.CrypterResponse, error) {
	var (
		conf *pdf.Configuration
		err  error
	)

	conf = pdf.NewRC4Configuration(req.InitPassword, req.InitPassword, 40)
	conf.Cmd = pdf.CHANGEUPW
	conf.UserPW = req.InitPassword
	conf.UserPWNew = &req.NewPassword

	// if no initial password is supplied the api does not change the
	// password but sets a password via the ENCRYPT command
	if req.InitPassword == "" && req.NewPassword != "" {
		conf = pdf.NewRC4Configuration(req.NewPassword, req.NewPassword, 40)
		conf.Cmd = pdf.ENCRYPT
		conf.UserPWNew = &req.NewPassword
	}

	buff := new(bytes.Buffer)
	if err = pdfApi.Optimize(bytes.NewReader(req.Payload), buff, conf); err != nil {
		return nil, fmt.Errorf("could not process PDF payload: %v", err)
	}

	return &proto.CrypterResponse{
		Payload: buff.Bytes(),
		Result:  proto.CrypterResponse_SUCCESS,
		Message: fmt.Sprintf("successfully encrypted the provided payload"),
	}, nil
}
