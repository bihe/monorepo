package handler

import (
	"net/http"
)

// --------------------------------------------------------------------------
// Objects
// --------------------------------------------------------------------------

// Meta specifies application metadata
// swagger:model
type Meta struct {
	UserInfo UserInfo    `json:"userInfo"`
	Version  VersionInfo `json:"versionInfo"`
}

// UserInfo provides information about the currently logged-in user
// swagger:model
type UserInfo struct {
	DisplayName string   `json:"displayName"`
	UserID      string   `json:"userId"`
	UserName    string   `json:"userName"`
	Email       string   `json:"email"`
	Roles       []string `json:"roles"`
}

// VersionInfo is used to provide version and build
// swagger:model
type VersionInfo struct {
	Version string `json:"version"`
	Build   string `json:"buildNumber"`
}

// --------------------------------------------------------------------------
// Request and Response objects using go-chi render
// --------------------------------------------------------------------------

// AppInfoResponse wraps the data struct into a framework response
type AppInfoResponse struct {
	*Meta
}

// Render the specific response
func (a AppInfoResponse) Render(w http.ResponseWriter, r *http.Request) error {
	// Pre-processing before a response is marshalled and sent across the wire
	return nil
}
