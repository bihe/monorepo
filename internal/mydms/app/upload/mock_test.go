package upload_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
)

// the following logic was copied from https://git.sr.ht/~ewintr/erikwinternl/
// found in the article: https://erikwinter.nl/notes/2020/unit_test_outbound_http_requests_in_golang/

// MockResponse represents a response for the mock server to serve
type MockResponse struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
}

type MockServerProcedure struct {
	URI        string
	HTTPMethod string
	Response   MockResponse
}

// MockRecorder provides a way to record request information from every
// successful request.
type MockRecorder interface {
	Record(r *http.Request)
}

// recordedRequest represents recorded structured information about each request
type recordedRequest struct {
	hits     int
	requests []*http.Request
	bodies   [][]byte
}

// MockAssertion represents a common assertion for requests
type MockAssertion struct {
	indexes map[string]int    // indexation for key
	recs    []recordedRequest // request catalog
}

// Record records request hit information
func (m *MockAssertion) Record(r *http.Request) {
	k := m.index(r.RequestURI, r.Method)

	b, _ := ioutil.ReadAll(r.Body)
	if len(b) == 0 {
		b = nil
	}

	if k < 0 {
		m.newIndex(r.RequestURI, r.Method)
		m.recs = append(m.recs, recordedRequest{
			hits:     1,
			requests: []*http.Request{r},
			bodies:   [][]byte{b},
		})
		return
	}

	m.recs[k].hits++
	m.recs[k].requests = append(m.recs[k].requests, r)
	m.recs[k].bodies = append(m.recs[k].bodies, b)
}

// Hits returns the number of hits for a uri and method
func (m *MockAssertion) Hits(uri, method string) int {
	k := m.index(uri, method)
	if k < 0 {
		return 0
	}

	return m.recs[k].hits
}

// Headers returns a slice of request headers
func (m *MockAssertion) Headers(uri, method string) []http.Header {
	k := m.index(uri, method)
	if k < 0 {
		return nil
	}

	headers := make([]http.Header, len(m.recs[k].requests))
	for i, r := range m.recs[k].requests {

		// remove default headers
		if _, ok := r.Header["Content-Length"]; ok {
			r.Header.Del("Content-Length")
		}

		if v, ok := r.Header["User-Agent"]; ok {
			if _, yes := equals([]string{"Go-http-client/1.1"}, v); yes {
				r.Header.Del("User-Agent")
			}
		}

		if v, ok := r.Header["Accept-Encoding"]; ok {
			if _, yes := equals([]string{"gzip"}, v); yes {
				r.Header.Del("Accept-Encoding")
			}
		}

		if len(r.Header) == 0 {
			continue
		}

		headers[i] = r.Header
	}
	return headers
}

// Body returns request body
func (m *MockAssertion) Body(uri, method string) [][]byte {
	k := m.index(uri, method)
	if k < 0 {
		return nil
	}

	return m.recs[k].bodies
}

// Reset sets all unexpected properties to their zero value
func (m *MockAssertion) Reset() error {
	m.indexes = make(map[string]int)
	m.recs = make([]recordedRequest, 0)
	return nil
}

// index indexes a key composed of the uri and method and returns the position
// for this key in a list if it was indexed before.
func (m *MockAssertion) index(uri, method string) int {
	if isZero(m.indexes) {
		m.indexes = make(map[string]int)
	}

	k := strings.ToLower(uri + method)

	if i, ok := m.indexes[k]; ok {
		return i
	}

	return -1
}

func (m *MockAssertion) newIndex(uri, method string) int {
	k := strings.ToLower(uri + method)
	m.indexes[k] = len(m.indexes)
	return m.indexes[k]
}

// NewMockServer return a mock HTTP server to test requests
func NewMockServer(rec MockRecorder, procedures ...MockServerProcedure) *httptest.Server {
	var handler http.Handler

	handler = http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {

			for _, proc := range procedures {

				if proc.URI == r.URL.RequestURI() && proc.HTTPMethod == r.Method {

					headers := w.Header()
					for hkey, hvalue := range proc.Response.Headers {
						headers[hkey] = hvalue
					}

					code := proc.Response.StatusCode
					if code == 0 {
						code = http.StatusOK
					}

					w.WriteHeader(code)
					w.Write(proc.Response.Body)

					if rec != nil {
						rec.Record(r)
					}
					return
				}
			}

			w.WriteHeader(http.StatusNotFound)
			return
		})

	return httptest.NewServer(handler)
}

func isZero(anything interface{}) bool {
	refZero := reflect.Zero(reflect.ValueOf(anything).Type())
	return reflect.DeepEqual(refZero.Interface(), anything)
}

func equals(exp, act interface{}) (b *bytes.Buffer, ok bool) {
	b = new(bytes.Buffer)
	fmt.Fprintf(b, "\texp: %s\n\n\tgot: %s", stringer(exp), stringer(act))
	return b, reflect.DeepEqual(exp, act)
}

func stringer(a interface{}) string {
	switch s := a.(type) {
	case string:
		return s
	case []byte:
		return string(s)
	default:
		return fmt.Sprintf("%#v", s)
	}
}
