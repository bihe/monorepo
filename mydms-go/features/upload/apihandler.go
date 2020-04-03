package upload

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/bihe/mydms/internal/errors"
	"github.com/bihe/mydms/internal/persistence"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
	"golang.binggl.net/commons"
)

// Result represents status of the upload opeation
type Result struct {
	Token   string `json:"token"`
	Message string `json:"message"`
}

// Handler defines the upload API
type Handler struct {
	r      Repository
	config Config
	log    *log.Entry
}

// NewHandler returns a pointer to a new handler instance
func NewHandler(r Repository, config Config, logger *log.Entry) *Handler {
	return &Handler{r: r, config: config, log: logger}
}

// UploadFile godoc
// @Summary upload a document
// @Description temporarily stores a file and creates a item in the repository
// @Tags upload
// @Consumes multipart/form-data
// @Produce  json
// @Param file formData file true "file to upload"
// @Success 200 {object} upload.Result
// @Failure 401 {object} errors.ProblemDetail
// @Failure 403 {object} errors.ProblemDetail
// @Failure 500 {object} errors.ProblemDetail
// @Router /api/v1/uploads/file [post]
func (h *Handler) UploadFile(c echo.Context) error {
	// Source
	file, err := c.FormFile("file")
	if err != nil {
		return errors.BadRequestError{Err: fmt.Errorf("no file provided: %v", err), Request: c.Request()}
	}

	if file.Size > h.config.MaxUploadSize {
		return errors.BadRequestError{
			Err:     fmt.Errorf("the upload exceeds the maximum size of %d - filesize is: %d", h.config.MaxUploadSize, file.Size),
			Request: c.Request()}
	}

	commons.LogWithReq(c.Request(), h.log, "upload.UploadFile").Debugf("trying to upload file: '%s'", file.Filename)

	ext := filepath.Ext(file.Filename)
	ext = strings.Replace(ext, ".", "", 1)
	var typeAllowed = false
	for _, t := range h.config.AllowedFileTypes {
		if t == ext {
			typeAllowed = true
			break
		}
	}
	if !typeAllowed {
		return errors.BadRequestError{
			Err:     fmt.Errorf("the uploaded file-type '%s' is not allowed, only use: '%s'", ext, strings.Join(h.config.AllowedFileTypes, ",")),
			Request: c.Request()}
	}
	mimeType := file.Header.Get("Content-Type")

	src, err := file.Open()
	if err != nil {
		return errors.BadRequestError{Err: fmt.Errorf("could not open upload file: %v", err), Request: c.Request()}
	}
	defer src.Close()

	// Destination
	id := uuid.New().String()
	var tempFileName = id + "." + ext
	uploadPath := path.Join(h.config.UploadPath, tempFileName)
	dst, err := os.Create(uploadPath)
	if err != nil {
		return errors.ServerError{Err: fmt.Errorf("could not create file: %v", err), Request: c.Request()}
	}
	defer dst.Close()

	// Copy
	if _, err = io.Copy(dst, src); err != nil {
		return errors.ServerError{Err: fmt.Errorf("could not copy file: %v", err), Request: c.Request()}
	}

	u := Upload{
		ID:       id,
		FileName: file.Filename,
		MimeType: mimeType,
		Created:  time.Now().UTC(),
	}
	err = h.r.Write(u, persistence.Atomic{})
	if err != nil {
		ioerr := os.Remove(uploadPath)
		if ioerr != nil {
			commons.LogWithReq(c.Request(), h.log, "upload.UploadFile").Warnf("Clean-Up file-upload. Could not delete temp file: '%s': %v", uploadPath, ioerr)
		}
		return errors.ServerError{Err: fmt.Errorf("could not save upload item in store: %v", err), Request: c.Request()}
	}
	c.JSON(http.StatusCreated, Result{Token: id, Message: fmt.Sprintf("File '%s' was uploaded successfully!", file.Filename)})

	return nil
}
