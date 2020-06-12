package api

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"golang.binggl.net/monorepo/login/persistence"
	"golang.binggl.net/monorepo/pkg/errors"
	"golang.binggl.net/monorepo/pkg/logging"
	"golang.binggl.net/monorepo/pkg/security"

	per "golang.binggl.net/monorepo/pkg/persistence"
)

// swagger:operation GET /sites sites HandleGetSites
//
// sites of current user
//
// returns all the sites of the current loged-in user
//
// ---
// produces:
// - application/json
// responses:
//   '200':
//     description: UserSites
//     schema:
//       "$ref": "#/definitions/UserSites"
//   '401':
//     description: ProblemDetail
//     schema:
//       "$ref": "#/definitions/ProblemDetail"
//   '403':
//     description: ProblemDetail
//     schema:
//       "$ref": "#/definitions/ProblemDetail"
//   '404':
//     description: ProblemDetail
//     schema:
//       "$ref": "#/definitions/ProblemDetail"
func (a *loginAPI) HandleGetSites(user security.User, w http.ResponseWriter, r *http.Request) error {
	sites, err := a.repo.GetSitesByUser(user.Email)
	if err != nil {
		logging.LogWithReq(r, a.logEntry, "api.HandleGetSites").Warnf("cannot get sites of current user '%s', %v", user.Email, err)
		return errors.NotFoundError{Err: fmt.Errorf("no sites for given user '%s'", user.Email), Request: r}
	}

	u := UserSites{
		User: user.Username,
	}
	u.Editable = a.hasRole(user, a.editRole)
	for _, s := range sites {
		parts := strings.Split(s.PermList, ";")
		u.Sites = append(u.Sites, SiteInfo{
			Name: s.Name,
			URL:  s.URL,
			Perm: parts,
		})
	}
	a.respond(w, r, http.StatusOK, u)
	return nil
}

// swagger:operation POST /sites sites HandleSaveSites
//
// stores the given sites
//
// takes a list of sites and stores the supplied sites for the user
//
// ---
// consumes:
// - application/json
// produces:
// - application/json
// responses:
//   '201':
//     description: Success
//   '401':
//     description: ProblemDetail
//     schema:
//       "$ref": "#/definitions/ProblemDetail"
//   '403':
//     description: ProblemDetail
//     schema:
//       "$ref": "#/definitions/ProblemDetail"
//   '404':
//     description: ProblemDetail
//     schema:
//       "$ref": "#/definitions/ProblemDetail"
func (a *loginAPI) HandleSaveSites(user security.User, w http.ResponseWriter, r *http.Request) error {
	if !a.hasRole(user, a.editRole) {
		logging.LogWithReq(r, a.logEntry, "api.HandleSaveSites").Warnf("user '%s' tried to save but does not have required permissions", user.Email)
		return errors.SecurityError{Err: fmt.Errorf("user '%s' is not allowed to perform this action", user.Email), Request: r}
	}

	// payload is an array of SiteInfo
	var (
		payload []SiteInfo
		err     error
	)
	if err = a.decode(w, r, &payload); err != nil {
		return errors.BadRequestError{Err: fmt.Errorf("could not use supplied payload: %v", err), Request: r}
	}

	var sites []persistence.UserSite
	for _, p := range payload {
		sites = append(sites, persistence.UserSite{
			Name:     p.Name,
			URL:      p.URL,
			PermList: strings.Join(p.Perm, ";"),
			User:     user.Email,
			Created:  time.Now().UTC(),
		})
	}
	err = a.repo.StoreSiteForUser(user.Email, sites, per.Atomic{})
	if err != nil {
		logging.LogWithReq(r, a.logEntry, "api.HandleSaveSites").Errorf("could not save sites of user '%s': %v", user.Email, err)
		return errors.ServerError{Err: fmt.Errorf("could not save payload: %v", err), Request: r}
	}
	w.WriteHeader(http.StatusCreated)
	return nil
}

// swagger:operation GET /sites/users/{siteName} sites HandleGetUsersForSite
//
// returns the users for a given site
//
// determine users who have access to a given site and return them
//
// ---
// produces:
// - application/json
// parameters:
// - name: siteName
//   in: path
//   description: the name of the site
//   required: true
//   type: string
// responses:
//   '200':
//     description: UserList
//     schema:
//       "$ref": "#/definitions/UserList"
//   '401':
//     description: ProblemDetail
//     schema:
//       "$ref": "#/definitions/ProblemDetail"
//   '403':
//     description: ProblemDetail
//     schema:
//       "$ref": "#/definitions/ProblemDetail"
//   '404':
//     description: ProblemDetail
//     schema:
//       "$ref": "#/definitions/ProblemDetail"
func (a *loginAPI) HandleGetUsersForSite(user security.User, w http.ResponseWriter, r *http.Request) error {
	siteName := chi.URLParam(r, "siteName")
	users, err := a.repo.GetUsersForSite(siteName)
	if err != nil {
		logging.LogWithReq(r, a.logEntry, "api.HandleGetUsersForSite").Warnf("cannot get users for site '%s', %v", siteName, err)
		return errors.NotFoundError{Err: fmt.Errorf("no users available for given site '%s'", siteName), Request: r}
	}
	a.respond(w, r, http.StatusOK, UserList{
		Count: len(users),
		Users: users,
	})
	return nil
}
