package api

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"golang.binggl.net/monorepo/internal/core/app/shared"
	"golang.binggl.net/monorepo/internal/core/app/sites"
	"golang.binggl.net/monorepo/pkg/logging"
)

// A SitesHandler provides the handling functions for the sites API
type SitesHandler struct {
	SitesSvc sites.Service
	Logger   logging.Logger
}

// handleGetSitesForUser returns the available sites for the given user - who is authenticated via JWT
func (s SitesHandler) HandleGetSitesForUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := ensureUser(r)
		s.Logger.Debug("retrieve the site with name for user", logging.LogV("username", user.DisplayName))
		sites, err := s.SitesSvc.GetSitesForUser(*user)
		if err != nil {
			s.Logger.Error("could not retrieve sites for user", logging.LogV("username", user.DisplayName), logging.ErrV(err))
			encodeError(err, s.Logger, w)
			return
		}
		s.Logger.Debug("sites", logging.LogV("site_count", fmt.Sprintf("%d", len(sites.Sites))))
		respondJSON(w, sites)
	}
}

// handleGetUsersForSite take a **siteName** and returns the users for the given site
func (s SitesHandler) HandleGetUsersForSite() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := ensureUser(r)
		siteName := chi.URLParam(r, "siteName")
		s.Logger.Debug("get users for site", logging.LogV("sitname", siteName))

		users, err := s.SitesSvc.GetUsersForSite(siteName, *user)
		if err != nil {
			s.Logger.Error("could not get users for site", logging.LogV("siteName", siteName), logging.ErrV(err))
			encodeError(err, s.Logger, w)
			return
		}
		s.Logger.Debug("users", logging.LogV("user_count", fmt.Sprintf("%d", len(users))))
		respondJSON(w, users)
	}
}

// --------------------------------------------------------------------------
//   UserSitesRequest encapsulates a save request where a UserSites payload
//   is provided
// --------------------------------------------------------------------------

// UserSitesRequest is the request payload for the UserSites data model.
type UserSitesRequest struct {
	*sites.UserSites
}

// Bind assigns the the provided data to a UserSitesRequest
func (s *UserSitesRequest) Bind(r *http.Request) error {
	// a UserSites is nil if no UserSites fields are sent in the request. Return an
	// error to avoid a nil pointer dereference.
	if s.UserSites == nil {
		return fmt.Errorf("missing required UserSites fields")
	}
	return nil
}

// String returns a string representation of a UserSitesRequest
func (s *UserSitesRequest) String() string {
	return fmt.Sprintf("UserSites for '%s'; sites: %d", s.User, len(s.Sites))
}

// --------------------------------------------------------------------------

func (s SitesHandler) HandleSaveSitesForUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		payload := &UserSitesRequest{}
		user := ensureUser(r)
		s.Logger.Debug("save site data for user", logging.LogV("username", user.Email))
		if err := render.Bind(r, payload); err != nil {
			s.Logger.Error("could not bind payload", logging.LogV("user", user.Username), logging.ErrV(err))
			err = shared.ErrValidation(fmt.Sprintf("could not process the payload: %s", err.Error()))
			encodeError(err, s.Logger, w)
			return
		}
		s.Logger.Debug("will save sites", logging.LogV("site-info", payload.String()))
		err := s.SitesSvc.SaveSitesForUser(*payload.UserSites, *user)
		if err != nil {
			s.Logger.Error("could not save sites for user", logging.LogV("username", user.Email), logging.ErrV(err))
			encodeError(err, s.Logger, w)
			return
		}
		w.WriteHeader(http.StatusCreated)
	}
}
