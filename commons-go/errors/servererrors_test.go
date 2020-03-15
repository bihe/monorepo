package errors // import "golang.binggl.net/commons/errors"

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
	"golang.binggl.net/commons/cookies"
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
			Error:  NotFoundError{Err: fmt.Errorf(errText), Request: errReq},
			Accept: "application/json",
		},
		{
			Name:   "BadRequestError",
			Status: http.StatusBadRequest,
			Error:  BadRequestError{Err: fmt.Errorf(errText), Request: errReq},
			Accept: "application/json",
		},
		{
			Name:   "ServerError",
			Status: http.StatusInternalServerError,
			Error:  ServerError{Err: fmt.Errorf(errText), Request: errReq},
			Accept: "application/json",
		},
		{
			Name:     "RedirectErrorBrowser",
			Status:   http.StatusTemporaryRedirect,
			Error:    RedirectError{Err: fmt.Errorf(errText), Request: errReq, URL: "http://redirect", Status: http.StatusTemporaryRedirect},
			Accept:   "text/html",
			Redirect: "http://redirect",
		},
		{
			Name:   "RedirectErrorBrowserJSON",
			Status: http.StatusTemporaryRedirect,
			Error:  RedirectError{Err: fmt.Errorf(errText), Request: errReq, URL: "http://redirect", Status: http.StatusTemporaryRedirect},
			Accept: "application/json",
		},
		{
			Name:   "error",
			Status: http.StatusInternalServerError,
			Error:  fmt.Errorf(errText),
			Accept: "application/json",
		},
		{
			Name:     "error.HTML",
			Status:   http.StatusTemporaryRedirect,
			Error:    fmt.Errorf(errText),
			Accept:   "text/html",
			Redirect: "/error",
		},
		{
			Name:   "security",
			Status: http.StatusForbidden,
			Error:  SecurityError{Err: fmt.Errorf(errText), Request: errReq},
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
	errRep := &ErrorReporter{
		ErrorPath: "error",
		CookieSettings: cookies.Settings{
			Secure: false,
			Prefix: "test",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.Name, func(t *testing.T) {
			rec = httptest.NewRecorder()
			req = httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Add("Accept", tc.Accept)

			r.Get("/", func(w http.ResponseWriter, r *http.Request) {
				if tc.Error != nil {
					errRep.Negotiate(w, r, tc.Error)
					return
				}
				w.WriteHeader(http.StatusOK)
			})
			r.ServeHTTP(rec, req)

			assert.Equal(t, tc.Status, rec.Code)

			if tc.Redirect != "" {
				assert.Equal(t, tc.Redirect, rec.Header().Get("Location"))
				// check that the correct cookie was set
				assert.True(t, strings.Contains(rec.Header().Get("Set-Cookie"), errRep.CookieSettings.Prefix+"_"+FlashKeyError))
				return
			}
			s = rec.Body.String()
			if s != "" {
				assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &pd))
			}
		})
	}
}

func TestContentNegotiation(t *testing.T) {

	tests := []struct {
		name   string
		header string
		want   content
	}{{
		name:   "empty",
		header: "",
		want:   JSON,
	}, {
		name:   "html",
		header: "text/html",
		want:   HTML,
	}, {
		name:   "json",
		header: "application/json",
		want:   JSON,
	}, {
		name:   "text",
		header: "text/plain",
		want:   TEXT,
	}, {
		name:   "nosubtype",
		header: "text/",
		want:   JSON,
	}, {
		name:   "fancysubtype",
		header: "text/fancy",
		want:   JSON,
	}, {
		name:   "complext",
		header: "text/plain; q=0.5, application/json, text/x-dvi; q=0.8, text/x-c",
		want:   JSON,
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("Accept", test.header)

			content := negotiateContent(req)

			if content != test.want {
				t.Errorf("Unexpected value\ngot:  %+v\nwant: %+v", content, test.want)
			}
		})
	}
}
