package filestore

import (
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/bihe/mydms/internal/errors"
	"github.com/labstack/echo/v4"
)

// Handler defines the filestore API
type Handler struct {
	fs FileService
}

// NewHandler returns a pointer to a new handler instance
func NewHandler(fs FileService) *Handler {
	return &Handler{fs}
}

// GetFile godoc
// @Summary get a file from the backend store
// @Description use a base64 encoded path to fetch the binary payload of a file from the store
// @Tags filestore
// @Param path query string true "Path"
// @Success 200 {array} byte
// @Failure 401 {object} errors.ProblemDetail
// @Failure 403 {object} errors.ProblemDetail
// @Failure 400 {object} errors.ProblemDetail
// @Failure 404 {object} errors.ProblemDetail
// @Failure 500 {object} errors.ProblemDetail
// @Router /api/v1/file [get]
func (h *Handler) GetFile(c echo.Context) error {
	path := c.QueryParam("path")
	decodedPath, err := base64.StdEncoding.DecodeString(path)
	if err != nil {
		return errors.BadRequestError{
			Err:     fmt.Errorf("the supplied path param cannot be decoded. %v", err),
			Request: c.Request()}
	}

	file, err := h.fs.GetFile(string(decodedPath))
	if err != nil {
		return errors.NotFoundError{
			Err:     fmt.Errorf("file not found '%s'. %v", decodedPath, err),
			Request: c.Request(),
		}
	}

	return c.Blob(http.StatusOK, file.MimeType, file.Payload)
}
