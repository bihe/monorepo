package api

// import (
// 	"errors"
// 	"fmt"
// 	"net/http"

// 	"github.com/go-chi/chi/v5"
// 	"golang.binggl.net/monorepo/internal/core/app/shared"
// 	"golang.binggl.net/monorepo/internal/core/app/upload"
// 	"golang.binggl.net/monorepo/pkg/logging"
// )

// // UploadHandler defines the api logic for the upload-service
// type UploadHandler struct {
// 	Service upload.Service
// 	Logger  logging.Logger
// }

// // --------------------------------------------------------------------------
// // Types
// // --------------------------------------------------------------------------

// // UploadResult represents status of the upload opeation
// type UploadResult struct {
// 	ID      string `json:"id"`
// 	Message string `json:"message"`
// }

// // --------------------------------------------------------------------------

// // Upload saves the provided payload using the store
// func (h *UploadHandler) Upload() http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		user := ensureUser(r)
// 		// Source
// 		file, fileHeader, err := r.FormFile("file")
// 		if err != nil {
// 			err = shared.ErrValidation(fmt.Sprintf("no file provided: %s", err.Error()))
// 			encodeError(err, h.Logger, w)
// 			return
// 		}
// 		defer file.Close()

// 		id, err := h.Service.Save(upload.File{
// 			Name:     fileHeader.Filename,
// 			Size:     fileHeader.Size,
// 			MimeType: fileHeader.Header.Get("Content-Type"),
// 			File:     file,
// 			Enc: upload.EncryptionRequest{
// 				InitPassword: r.FormValue("initPass"),
// 				Password:     r.FormValue("pass"),
// 				Token:        user.Token,
// 			},
// 		})
// 		if err != nil {
// 			h.Logger.Error("could not save upload file", logging.ErrV(err))
// 			if errors.Is(err, upload.ErrValidation) {
// 				err = shared.ErrValidation(fmt.Sprintf("error saving upload file: %s", err.Error()))
// 				encodeError(err, h.Logger, w)
// 				return
// 			}
// 			encodeError(err, h.Logger, w)
// 			return
// 		}
// 		w.WriteHeader(http.StatusCreated)
// 		respondJSON(w, UploadResult{
// 			ID:      id,
// 			Message: fmt.Sprintf("File '%s' was uploaded successfully!", fileHeader.Filename),
// 		})
// 	}
// }

// // GetItemByID returns the uploaded item by it's id
// func (h *UploadHandler) GetItemByID() http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		id := chi.URLParam(r, "id")

// 		item, err := h.Service.Read(id)
// 		if err != nil {
// 			if errors.Is(err, upload.ErrInvalidParameters) {
// 				err = shared.ErrValidation(err.Error())
// 				encodeError(err, h.Logger, w)
// 				return
// 			}
// 			err = shared.ErrNotFound(fmt.Sprintf("cannot get item by id '%s': %v", id, err))
// 			encodeError(err, h.Logger, w)
// 			return
// 		}
// 		respondJSON(w, item)
// 	}
// }

// // DeleteItemByID removes an upload item specified by it's id
// func (h *UploadHandler) DeleteItemByID() http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		id := chi.URLParam(r, "id")

// 		err := h.Service.Delete(id)
// 		if err != nil {

// 			if errors.Is(err, upload.ErrInvalidParameters) {
// 				err = shared.ErrValidation(err.Error())
// 				encodeError(err, h.Logger, w)
// 				return
// 			}
// 			err = shared.ErrNotFound(fmt.Sprintf("cannot get item by id '%s': %v", id, err))
// 			encodeError(err, h.Logger, w)
// 			return
// 		}
// 		respondJSON(w, UploadResult{
// 			ID:      id,
// 			Message: "Item with was deleted successfully!",
// 		})
// 	}
// }
