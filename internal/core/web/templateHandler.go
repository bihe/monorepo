package web

import (
	"encoding/json"
	"fmt"
	"net/http"

	"golang.binggl.net/monorepo/internal/common"
	"golang.binggl.net/monorepo/internal/core/app/sites"
	"golang.binggl.net/monorepo/internal/core/web/html"
	"golang.binggl.net/monorepo/pkg/handler"
	base "golang.binggl.net/monorepo/pkg/handler/html"
)

// TemplateHandler takes care of providing HTML templates.
// This is the new approach with a template + htmx based UI to replace the angular frontend
// and have a more go-oriented approach towards UI and user-interaction. This reduces the
// cognitive load because less technology mix is needed. Via template + htmx approach only
// a limited amount of javascript is needed to achieve the frontend.
// As additional benefit the build should be faster, because the nodejs build can be removed
type TemplateHandler struct {
	*handler.TemplateHandler
	SiteSvc sites.Service
	Version string
	Build   string
}

const searchURL = "/sites/search"

// DisplaySites determined the sites of the current user and displays them
func (t *TemplateHandler) DisplaySites() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := common.EnsureUser(r)
		search := ""
		t.Logger.InfoRequest(fmt.Sprintf("display the sites for user: '%s'", user.Username), r)

		usrSites, err := t.SiteSvc.GetSitesForUser(*user)
		if err != nil {
			t.Logger.ErrorRequest(fmt.Sprintf("could not get sites for user '%s'; '%v'", user.Username, err), r)
		}

		base.Layout(
			common.CreatePageModel("/sites", "Available Apps", search, "/public/folder.svg", t.versionString(), t.Env, *user),
			html.SiteStyles(),
			html.SiteNavigation(search),
			html.SiteContent(usrSites),
			searchURL,
		).Render(w)
	}
}

// ShowEditSites displays the edit form to change the application JSON
func (t *TemplateHandler) ShowEditSites() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := common.EnsureUser(r)
		search := ""
		t.Logger.InfoRequest(fmt.Sprintf("display the sites for user: '%s'", user.Username), r)

		usrSites, err := t.SiteSvc.GetSitesForUser(*user)
		if err != nil {
			t.Logger.ErrorRequest(fmt.Sprintf("could not get sites for user '%s'; '%v'", user.Username, err), r)
		}
		// for simple edit purposes we display the whole payload as JSON
		jsonPayload := handler.JsonIndent[sites.UserSites](usrSites)

		base.Layout(
			common.CreatePageModel("/sites", "Edit Apps", search, "/public/folder.svg", t.versionString(), t.Env, *user),
			html.SiteEditStyles(),
			html.SiteEditNavigation(search),
			html.SiteEditContent(jsonPayload, ""),
			searchURL,
		).Render(w)
	}
}

// SaveSites processed the provided JSON payload and saves the applications
func (t *TemplateHandler) SaveSites() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := common.EnsureUser(r)
		t.Logger.InfoRequest(fmt.Sprintf("receive the payload to save the applications: '%s'", user.Username), r)

		err := r.ParseForm()
		if err != nil {
			t.Logger.ErrorRequest(fmt.Sprintf("could not parse supplied form data; '%v'", err), r)
			t.RenderErr(r, w, fmt.Sprintf("could not parse supplied form data; '%v'", err))
			return
		}

		payload := r.FormValue("payload")
		var usrSites sites.UserSites
		err = json.Unmarshal([]byte(payload), &usrSites)
		if err != nil {
			t.Logger.ErrorRequest(fmt.Sprintf("the supplied JSON payload is invalid and cannot be processed; '%v'", err), r)
			// return to edit page
			html.SiteEditContent(payload, fmt.Sprintf("the supplied JSON payload is invalid and cannot be processed; '%v'", err)).Render(w)
			return
		}

		err = t.SiteSvc.SaveSitesForUser(usrSites, *user)
		if err != nil {
			t.Logger.ErrorRequest(fmt.Sprintf("could not save the sites; '%v'", err), r)
			// return to edit page
			html.SiteEditContent(payload, fmt.Sprintf("could not save the sites; '%v'", err)).Render(w)
			return
		}

		// operation was successful return to the overview
		w.Header().Add("HX-Location", "/sites")
	}
}

// --------------------------------------------------------------------------
//  Internals
// --------------------------------------------------------------------------

func (t *TemplateHandler) versionString() string {
	return fmt.Sprintf("%s-%s", t.Version, t.Build)
}
