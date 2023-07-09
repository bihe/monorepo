package site

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-resty/resty/v2"
)

// types
// --------------------------------------------------------------------------
type UserSite struct {
	Editable  bool       `json:"editable,omitempty"`
	User      string     `json:"user,omitempty"`
	UserSites []SiteInfo `json:"userSites,omitempty"`
}

type SiteInfo struct {
	Name        string   `json:"name,omitempty"`
	URL         string   `json:"url,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
}

func GetSiteClient(apiEndpoint, authToken string) *SiteClient {
	return &SiteClient{
		ApiEndpoint: apiEndpoint,
		AuthToken:   authToken,
		client:      resty.New(),
	}
}

// SiteClient is used to interact with the sites api
type SiteClient struct {
	client      *resty.Client
	ApiEndpoint string
	AuthToken   string
}

// GetSites retrieves the sites for the current user
func (s *SiteClient) GetSites() (*UserSite, error) {
	var sites UserSite
	resp, err := s.client.R().
		SetAuthToken(s.AuthToken).
		Get(s.ApiEndpoint)
	if err != nil {
		return nil, fmt.Errorf("could not fetch sites; %w", err)
	}

	err = json.Unmarshal(resp.Body(), &sites)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal payload; %w", err)
	}

	return &sites, nil
}

// SaveSites saves the given payload
func (s *SiteClient) SaveSites(data UserSite) error {
	resp, err := s.client.R().
		SetAuthToken(s.AuthToken).
		SetBody(data).
		Post(s.ApiEndpoint)
	if err != nil {
		return fmt.Errorf("could not save sites; %w", err)
	}
	if resp.StatusCode() != http.StatusCreated {
		return fmt.Errorf("wrong status provided: %d", resp.StatusCode())
	}
	return nil
}
