package api

import (
	"net/http"

	"golang.binggl.net/monorepo/pkg/handler"
	"golang.binggl.net/monorepo/pkg/logging"
	"golang.binggl.net/monorepo/pkg/security"
)

// A AppInfoHandler returns metadata about the application
type AppInfoHandler struct {
	Logger  logging.Logger
	Version string
	Build   string
}

// handleGetAppInfo returns metadata about the application and the current available user
func (s AppInfoHandler) handleGetAppInfo() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := ensureUser(r)
		s.Logger.Debug("get application info", logging.LogV("username", user.DisplayName))
		respondJSON(w, handler.Meta{
			UserInfo: handler.UserInfo{
				DisplayName: user.DisplayName,
				UserID:      user.UserID,
				UserName:    user.Username,
				Email:       user.Email,
				Roles:       user.Roles,
			},
			Version: handler.VersionInfo{
				Version: s.Version,
				Build:   s.Build,
			},
		})
	}
}

// A WhoAmIResponse is used to provide information about the user.
// It is a more generel information, not specific to an application but gives overall information
type WhoAmIResponse struct {
	DisplayName string           `json:"displayName"`
	UserID      string           `json:"userId"`
	UserName    string           `json:"userName"`
	Email       string           `json:"email"`
	PictureURL  string           `json:"pictureUrl"`
	Claims      []security.Claim `json:"claims"`
}

// handleGetAppInfo returns metadata about the application and the current available user
func (s AppInfoHandler) handleWhoAmI() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := ensureUser(r)
		s.Logger.Debug("get information about the user", logging.LogV("username", user.DisplayName))
		respondJSON(w, WhoAmIResponse{
			DisplayName: user.DisplayName,
			UserID:      user.UserID,
			UserName:    user.Username,
			Email:       user.Email,
			PictureURL:  user.ProfileURL,
			Claims:      user.Claims,
		})
	}
}
