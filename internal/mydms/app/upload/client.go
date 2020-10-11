package upload

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

// --------------------------------------------------------------------------
// Types
// --------------------------------------------------------------------------

// Upload defines an entity within the persistence store
type Upload struct {
	ID       string    `json:"id"`
	FileName string    `json:"fileName"`
	Payload  []byte    `json:"payload"`
	MimeType string    `json:"mimeType"`
	Created  time.Time `json:"created"`
}

// Client is used to interact with the backend upload services
type Client interface {
	// Get returns an object for the given ID
	Get(id, authToken string) (Upload, error)
	// Delete removes the upload-object specified by the given ID
	Delete(id, authToken string) error
}

// --------------------------------------------------------------------------
// Implementation
// --------------------------------------------------------------------------

// NewClient instantiates a new client to interact with the Upload service
func NewClient(baseURL string) Client {
	return &httpClient{
		baseURL: baseURL,
	}
}

var _ Client = &httpClient{} // compile time guard

type httpClient struct {
	baseURL string
}

func (c *httpClient) Get(id, authToken string) (Upload, error) {
	var u Upload

	url := getURL(c.baseURL, id)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return u, fmt.Errorf("error creating request, %v", err)
	}
	setHeader(req, authToken)

	client := &http.Client{Timeout: time.Second * 30}
	resp, err := client.Do(req)
	if err != nil {
		return u, fmt.Errorf("could not get upload by id '%s', %v", id, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return u, fmt.Errorf("could not get upload by id '%s', got status-code '%d'", id, resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return u, fmt.Errorf("could not read response for id '%s', %v", id, err)
	}
	if len(body) == 0 {
		return u, fmt.Errorf("empty response for id '%s'", id)
	}
	if err = json.Unmarshal(body, &u); err != nil {
		return u, fmt.Errorf("could not unmarshal response for id '%s', %v", id, err)
	}
	return u, nil
}

func (c *httpClient) Delete(id, authToken string) error {
	url := getURL(c.baseURL, id)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("error creating request, %v", err)
	}
	setHeader(req, authToken)

	client := &http.Client{Timeout: time.Second * 30}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("could not delete upload by id '%s', %v", id, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("could not delete upload by id '%s', got status-code '%d'", id, resp.StatusCode)
	}

	return nil
}

func getURL(baseURL, id string) string {
	url := baseURL
	url = strings.TrimSuffix(url, "/")
	url = fmt.Sprintf("%s/upload/%s", url, id)
	return url
}

func setHeader(req *http.Request, token string) {
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
}
