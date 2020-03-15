// Package api implements the HTTP API of the login-application.
//
// the purpose of this application is to provide centralized authentication
//
// Terms Of Service:
//
//     Schemes: https
//     Host: login.binggl.net
//     BasePath: /api/v1
//     Version: 1.0.0
//     License: Apache 2.0 https://opensource.org/licenses/Apache-2.0
//
//     Consumes:
//     - application/json
//
//     Produces:
//     - application/json
//
// swagger:meta
package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"golang.binggl.net/commons/cookies"
	"golang.binggl.net/commons/errors"
	"golang.binggl.net/commons/handler"
	"golang.binggl.net/commons/security"
	"github.com/bihe/login-go/internal"
	"github.com/bihe/login-go/internal/config"
	"github.com/bihe/login-go/internal/persistence"
	"golang.org/x/oauth2"

	log "github.com/sirupsen/logrus"
)

// --------------------------------------------------------------------------
// models
// --------------------------------------------------------------------------

// UserSites holds information about the current user and sites
// swagger:model
type UserSites struct {
	User     string     `json:"user"`
	Editable bool       `json:"editable"`
	Sites    []SiteInfo `json:"userSites"`
}

// SiteInfo holds data of a site
// swagger:model
type SiteInfo struct {
	Name string   `json:"name"`
	URL  string   `json:"url"`
	Perm []string `json:"permissions"`
}

// UserList holds the usernames for a given site
// swagger:model
type UserList struct {
	Count int      `json:"count"`
	Users []string `json:"users"`
}

// --------------------------------------------------------------------------
// Swagger specific definitions
// --------------------------------------------------------------------------

// swagger:parameters HandleSaveSites
type SiteInfoRequestSwagger struct {
	// In: body
	Body SiteInfo
}

// --------------------------------------------------------------------------
// API Interface
// --------------------------------------------------------------------------

// Login defines the HTTP handlers responding to requests
type Login interface {
	// oidc
	HandleError(w http.ResponseWriter, r *http.Request) error
	HandleOIDCRedirect(w http.ResponseWriter, r *http.Request) error
	HandleAuthFlow(w http.ResponseWriter, r *http.Request) error
	HandleOIDCRedirectFinal(w http.ResponseWriter, r *http.Request) error
	HandleOIDCLogin(w http.ResponseWriter, r *http.Request) error
	HandleLogout(user security.User, w http.ResponseWriter, r *http.Request) error

	// sites
	HandleGetSites(user security.User, w http.ResponseWriter, r *http.Request) error
	HandleSaveSites(user security.User, w http.ResponseWriter, r *http.Request) error
	HandleGetUsersForSite(user security.User, w http.ResponseWriter, r *http.Request) error

	// wrapper methods
	Secure(f func(user security.User, w http.ResponseWriter, r *http.Request) error) http.HandlerFunc
	Call(f func(w http.ResponseWriter, r *http.Request) error) http.HandlerFunc

	// configuration
	GetOIDCRedirectURL() string
}

var _ Login = (*loginAPI)(nil)

// loginAPI implements the handlers of the API definition
type loginAPI struct {
	handler.Handler
	internal.VersionInfo
	errRep         *errors.ErrorReporter
	cookieSettings cookies.Settings
	appCookie      *cookies.AppCookie
	basePath       string

	// oidc objects
	oauthConfig   *oauth2.Config
	oauthVerifier OIDCVerifier

	// store
	repo persistence.Repository
	// jwt logic
	jwt config.Security
	// user having this role are allowed to edit
	editRole string
}

// New creates a new instance of the API type
func New(basePath string, baseHandler handler.Handler, cs cookies.Settings, version internal.VersionInfo, oauth config.OAuthConfig, jwt config.Security, repo persistence.Repository) Login {
	c, v := NewOIDC(oauth)
	api := loginAPI{
		Handler:        baseHandler,
		VersionInfo:    version,
		cookieSettings: cs,
		errRep: &errors.ErrorReporter{
			CookieSettings: cs,
			ErrorPath:      "/error",
		},
		appCookie:     cookies.NewAppCookie(cs),
		basePath:      basePath,
		oauthConfig:   c,
		oauthVerifier: v,
		repo:          repo,
		jwt:           jwt,
		editRole:      jwt.Claim.Roles[0], // use the first role of the defined claims as the edit role
	}

	return &api
}

// --------------------------------------------------------------------------
// internal API helpers
// --------------------------------------------------------------------------

// respond converts data into appropriate responses for the client
// this can be JSON, Plaintext, ...
func (a *loginAPI) respond(w http.ResponseWriter, r *http.Request, code int, data interface{}) {
	w.Header().Set("Content-Type", "application/problem+json; charset=utf-8")
	w.WriteHeader(code)
	if data != nil {
		err := json.NewEncoder(w).Encode(data)
		if err != nil {
			log.WithField("func", "server.respond").Errorf("could not marshal json %v\n", err)
			a.errRep.Negotiate(w, r, errors.ServerError{
				Err:     fmt.Errorf("could not marshal json %v", err),
				Request: r,
			})
			return
		}
	}
}

// decode parses supplied JSON payload
func (a *loginAPI) decode(w http.ResponseWriter, r *http.Request, v interface{}) error {
	if r.Body == nil {
		return fmt.Errorf("no body payload available")
	}
	return json.NewDecoder(r.Body).Decode(v)
}

// hasRole checks if the given user has the given role
func (a *loginAPI) hasRole(user security.User, role string) bool {
	for _, p := range user.Roles {
		if p == role {
			return true
		}
	}
	return false
}

func query(r *http.Request, name string) string {
	keys, ok := r.URL.Query()[name]

	if !ok || len(keys[0]) < 1 {
		log.WithField("func", "server.getQuery").Debugf("Url Param '%s' is missing", name)
		return ""
	}

	return keys[0]
}
