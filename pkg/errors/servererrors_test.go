package errors

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

const errText = "error occurred"

func TestErrorHandler(t *testing.T) {
	// Setup
	var (
		pd  ProblemDetail
		s   string
		req *http.Request
		rec *httptest.ResponseRecorder
		r   chi.Router
	)

	errReq := httptest.NewRequest(http.MethodGet, "/", nil)
	testcases := []struct {
		Name     string
		Status   int
		Error    error
		Accept   string
		Redirect string
	}{
		{
			Name:   "NotFoundError",
			Status: http.StatusNotFound,
			Error:  NotFoundError{Err: errors.New(errText), Request: errReq},
			Accept: "application/json",
		},
		{
			Name:   "BadRequestError",
			Status: http.StatusBadRequest,
			Error:  BadRequestError{Err: errors.New(errText), Request: errReq},
			Accept: "application/json",
		},
		{
			Name:   "ServerError",
			Status: http.StatusInternalServerError,
			Error:  ServerError{Err: errors.New(errText), Request: errReq},
			Accept: "application/json",
		},
		{
			Name:   "error",
			Status: http.StatusInternalServerError,
			Error:  errors.New(errText),
			Accept: "application/json",
		},
		{
			Name:   "SecurityError_Forbidden",
			Status: http.StatusForbidden,
			Error:  SecurityError{Err: errors.New(errText), Request: errReq},
			Accept: "application/json",
		},
		{
			Name:   "SecurityError_UnAuthorized",
			Status: http.StatusUnauthorized,
			Error:  SecurityError{Err: errors.New(errText), Request: errReq, Status: http.StatusUnauthorized},
			Accept: "application/json",
		},
		{
			Name:   "no-error",
			Status: http.StatusOK,
			Error:  nil,
			Accept: "application/json",
		},
	}

	r = chi.NewRouter()
	for _, tc := range testcases {
		t.Run(tc.Name, func(t *testing.T) {
			rec = httptest.NewRecorder()
			req = httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Add("Accept", tc.Accept)

			r.Get("/", func(w http.ResponseWriter, r *http.Request) {
				if tc.Error != nil {
					WriteError(w, r, tc.Error)
					return
				}
				w.WriteHeader(http.StatusOK)
			})
			r.ServeHTTP(rec, req)

			assert.Equal(t, tc.Status, rec.Code)
			s = rec.Body.String()
			if s != "" {
				assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &pd))
			}
		})
	}
}
