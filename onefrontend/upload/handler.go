package upload

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/google/uuid"
	"golang.binggl.net/monorepo/onefrontend/config"
	"golang.binggl.net/monorepo/pkg/errors"
	"golang.binggl.net/monorepo/pkg/handler"
	"golang.binggl.net/monorepo/pkg/logging"
	"golang.binggl.net/monorepo/pkg/security"
)

// --------------------------------------------------------------------------
// Types
// --------------------------------------------------------------------------

// Result represents status of the upload opeation
type Result struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

// --------------------------------------------------------------------------
// ResultRespone
// --------------------------------------------------------------------------

// ResultResponse returns Result
type ResultResponse struct {
	*Result
	Status int `json:"-"` // ignore this
}

// Render the specific response
func (b ResultResponse) Render(w http.ResponseWriter, r *http.Request) error {
	if b.Status == 0 {
		render.Status(r, http.StatusOK)
	} else {
		render.Status(r, b.Status)
	}
	return nil
}

// ItemResponse returns Result
type ItemResponse struct {
	*Upload
}

// Render the specific response
func (b ItemResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

// --------------------------------------------------------------------------
// Handler logic
// --------------------------------------------------------------------------

// Handler defines the api logic for the upload-service
type Handler struct {
	handler.Handler
	Store  Store
	Config config.UploadSettings
}

// GetHandlers returns the upload handler routes
func (h *Handler) GetHandlers() http.Handler {
	r := chi.NewRouter()
	r.Post("/file", h.Secure(h.Upload))
	r.Get("/{id}", h.Secure(h.GetItemByID))
	r.Delete("/{id}", h.Secure(h.DeleteItemByID))
	return r
}

// Upload saves the provided payload using the store
func (h *Handler) Upload(user security.User, w http.ResponseWriter, r *http.Request) error {
	// Source
	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		return errors.BadRequestError{Err: fmt.Errorf("no file provided: %v", err), Request: r}
	}
	defer file.Close()

	if fileHeader.Size > h.Config.MaxUploadSize {
		return errors.BadRequestError{
			Err:     fmt.Errorf("the upload exceeds the maximum size of %d - filesize is: %d", h.Config.MaxUploadSize, fileHeader.Size),
			Request: r}
	}

	logging.LogWithReq(r, h.Log, "upload.UploadFile").Debugf("trying to upload file: '%s'", fileHeader.Filename)

	ext := filepath.Ext(fileHeader.Filename)
	ext = strings.Replace(ext, ".", "", 1)
	var typeAllowed = false
	for _, t := range h.Config.AllowedFileTypes {
		if t == ext {
			typeAllowed = true
			break
		}
	}
	if !typeAllowed {
		return errors.BadRequestError{
			Err:     fmt.Errorf("the uploaded file-type '%s' is not allowed, only use: '%s'", ext, strings.Join(h.Config.AllowedFileTypes, ",")),
			Request: r}
	}
	mimeType := fileHeader.Header.Get("Content-Type")

	// Copy
	b := &bytes.Buffer{}
	if _, err := io.Copy(b, file); err != nil {
		return errors.ServerError{Err: fmt.Errorf("could not copy file: %v", err), Request: r}
	}
	id := uuid.New().String()
	u := Upload{
		ID:       id,
		FileName: fileHeader.Filename,
		MimeType: mimeType,
		Payload:  b.Bytes(),
		Created:  time.Now().UTC(),
	}
	if err = h.Store.Write(u); err != nil {
		return errors.ServerError{Err: fmt.Errorf("could not save upload file: %v", err), Request: r}
	}

	return render.Render(w, r, ResultResponse{
		Result: &Result{
			ID:      id,
			Message: fmt.Sprintf("File '%s' was uploaded successfully!", fileHeader.Filename),
		},
		Status: http.StatusCreated,
	})
}

// GetItemByID returns the uploaded item by it's id
func (h *Handler) GetItemByID(user security.User, w http.ResponseWriter, r *http.Request) error {
	id := chi.URLParam(r, "id")

	item, err := h.Store.Read(id)
	if err != nil {
		return errors.NotFoundError{
			Err:     fmt.Errorf("cannot get item by id '%s': %v", id, err),
			Request: r,
		}
	}

	return render.Render(w, r, ItemResponse{
		&item,
	})
}

// DeleteItemByID removes an upload item specified by it's id
func (h *Handler) DeleteItemByID(user security.User, w http.ResponseWriter, r *http.Request) error {
	id := chi.URLParam(r, "id")

	err := h.Store.Delete(id)
	if err != nil {
		return errors.NotFoundError{
			Err:     fmt.Errorf("cannot get item by id '%s': %v", id, err),
			Request: r,
		}
	}

	return render.Render(w, r, ResultResponse{
		Result: &Result{
			ID:      id,
			Message: fmt.Sprint("Item with was deleted successfully!"),
		},
		Status: http.StatusOK,
	})
}
