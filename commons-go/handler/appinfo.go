package handler // import "golang.binggl.net/commons/handler"

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"golang.binggl.net/commons"
	"golang.binggl.net/commons/security"
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

// --------------------------------------------------------------------------
// Handler implementation
// --------------------------------------------------------------------------

// AppInfoHandler is responsible to return meta-information about the application
type AppInfoHandler struct {
	Handler
	// Version of the application using https://semver.org/
	Version string
	// Build identifies the specific build, e.g. git-hash
	Build string
}

// GetHandler returns the appinfo handler
func (h *AppInfoHandler) GetHandler() http.Handler {
	r := chi.NewRouter()
	r.Get("/", h.Secure(h.HandleAppInfo))
	return r
}

// HandleAppInfo provides information about the application
// swagger:operation GET /appinfo appinfo HandleAppInfo
//
// provides information about the application
//
// meta-data of the application including authenticated user and version
//
// ---
// produces:
// - application/json
// responses:
//   '200':
//     description: Meta
//     schema:
//       "$ref": "#/definitions/Meta"
//   '401':
//     description: ProblemDetail
//     schema:
//       "$ref": "#/definitions/ProblemDetail"
//   '403':
//     description: ProblemDetail
//     schema:
//       "$ref": "#/definitions/ProblemDetail"
func (h *AppInfoHandler) HandleAppInfo(user security.User, w http.ResponseWriter, r *http.Request) error {
	commons.LogWithReq(r, h.Handler.Log, "handler.HandleAppInfo").Debugf("return the application metadata info")
	info := Meta{
		Version: VersionInfo{
			Version: h.Version,
			Build:   h.Build,
		},
		UserInfo: UserInfo{
			UserID:      user.UserID,
			UserName:    user.Username,
			Email:       user.Email,
			DisplayName: user.DisplayName,
			Roles:       user.Roles,
		},
	}
	return render.Render(w, r, AppInfoResponse{Meta: &info})
}
