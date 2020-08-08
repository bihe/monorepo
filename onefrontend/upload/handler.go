package upload

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	apperr "golang.binggl.net/monorepo/pkg/errors"
	"golang.binggl.net/monorepo/pkg/handler"
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
	Service Service
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
		return apperr.BadRequestError{Err: fmt.Errorf("no file provided: %v", err), Request: r}
	}
	defer file.Close()

	a := r.FormValue("name1")
	fmt.Print(a)

	id, err := h.Service.Save(File{
		Name:     fileHeader.Filename,
		Size:     fileHeader.Size,
		MimeType: fileHeader.Header.Get("Content-Type"),
		File:     file,
	})
	if err != nil {
		if errors.Is(err, ErrValidation) {
			return apperr.BadRequestError{
				Err:     err,
				Request: r,
			}
		}
		return apperr.ServerError{Err: fmt.Errorf("could not save upload file: %v", err), Request: r}
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

	item, err := h.Service.Read(id)
	if err != nil {

		if errors.Is(err, ErrInvalidParameters) {
			return apperr.BadRequestError{
				Err:     err,
				Request: r,
			}
		}
		return apperr.NotFoundError{
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

	err := h.Service.Delete(id)
	if err != nil {

		if errors.Is(err, ErrInvalidParameters) {
			return apperr.BadRequestError{
				Err:     err,
				Request: r,
			}
		}
		return apperr.NotFoundError{
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
