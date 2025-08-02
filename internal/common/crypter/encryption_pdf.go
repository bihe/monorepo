package crypter

import (
	"bytes"
	"fmt"

	pdfApi "github.com/pdfcpu/pdfcpu/pkg/api"
	pdfModel "github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

// PDF encryption usind pdfcpu

// When deployed in a containerized environment the PDF library pdfcpu needs some configuration.
// This configuration can be done in a yaml file; the path to the file can be provided as an
// environment variable.
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
// #############################
// #   Default configuration   #
// #############################

// # toggle for inFilename extension check (.pdf)
// checkFileNameExt: true

// reader15: true

// decodeAllStreams: false

// # validationMode:
// # ValidationStrict,
// # ValidationRelaxed,
// validationMode: ValidationRelaxed

// # eol for writing:
// # EolLF
// # EolCR
// # EolCRLF
// eol: EolLF

// writeObjectStream: true
// writeXRefStream: true
// encryptUsingAES: true

// # encryptKeyLength: max 256
// encryptKeyLength: 256

// # permissions for encrypted files:
// # 0xF0C3 (PermissionsNone)
// # 0xF8C7 (PermissionsPrint)
// # 0xFFFF (PermissionsAll)
// # See more at model.PermissionFlags and PDF spec table 22
// permissions: 0xF0C3

// # displayUnit:
// # points
// # inches
// # cm
// # mm
// unit: points

// # timestamp format: yyyy-mm-dd hh:mm
// # Switch month and year by using: 2006-02-01 15:04
// # See more at https://pkg.go.dev/time@go1.17.1#pkg-constants
// timestampFormat: 2006-01-02 15:04

// # date format: yyyy-mm-dd
// dateFormat: 2006-01-02

// # optimize duplicate content streams across pages
// optimizeDuplicateContentStreams: false

// # merge creates bookmarks
// createBookmarks: true

// # Viewer is expected to supply appearance streams for form fields.
// needAppearances: false

// encrypt pdf or change password of pdf
func encryptPdfPayload(payload []byte, initPass, newPass string) ([]byte, error) {
	var (
		conf *pdfModel.Configuration
		err  error
	)

	conf = pdfModel.NewRC4Configuration(initPass, initPass, 40)
	conf.ValidationMode = pdfModel.ValidationRelaxed
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
