package web

import (
	"fmt"
	"net/http"

	"golang.binggl.net/monorepo/internal/mydms/app/filestore"
	"golang.binggl.net/monorepo/pkg/logging"
	"golang.binggl.net/monorepo/pkg/text"
)

// FileHandler interacts with the backend filestore
type FileHandler struct {
	FileSvc filestore.FileService
	Logger  logging.Logger
}

// GetDocumentPayload retrieves the payload from the backend store
func (f *FileHandler) GetDocumentPayload() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := pathParam(r, "path")
		user := ensureUser(r)
		f.Logger.InfoRequest(fmt.Sprintf("fetch the document payload by id: '%s' for user: '%s'", path, user.Username), r)

		// the provided path is base64 encoded
		decodedPath := text.DecBase64SafePath(path)
		if decodedPath == "" {
			f.Logger.ErrorRequest("could not access the document payload", r)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		f.Logger.Debug(fmt.Sprintf("get payload for path '%s'", string(decodedPath)))
		file, err := f.FileSvc.GetFile(string(decodedPath))
		if err != nil {
			f.Logger.ErrorRequest(fmt.Sprintf("could not access the document payload; %v", err), r)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", file.MimeType)
		_, err = w.Write(file.Payload)
		if err != nil {
			f.Logger.Error(fmt.Sprintf("could not write document payload to client; %v", err))
		}
	}
}
