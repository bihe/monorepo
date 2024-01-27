package web

import (
	"encoding/json"
	"fmt"
	"net/http"

	"golang.binggl.net/monorepo/internal/common"
	"golang.binggl.net/monorepo/internal/core/app/sites"
	"golang.binggl.net/monorepo/internal/core/web/templates"
	"golang.binggl.net/monorepo/pkg/config"
	"golang.binggl.net/monorepo/pkg/develop"
	"golang.binggl.net/monorepo/pkg/handler"
	tmpl "golang.binggl.net/monorepo/pkg/handler/templates"
	"golang.binggl.net/monorepo/pkg/security"
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

// DisplaySites determined the sites of the current user and displays them
func (t *TemplateHandler) DisplaySites() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := ensureUser(r)
		search := ""
		t.Logger.InfoRequest(fmt.Sprintf("display the sites for user: '%s'", user.Username), r)

		usrSites, err := t.SiteSvc.GetSitesForUser(*user)
		if err != nil {
			t.Logger.ErrorRequest(fmt.Sprintf("could not get sites for user '%s'; '%v'", user.Username, err), r)
		}
		tmpl.Layout(
			t.layoutModel("Available Apps", search, "/public/folder.svg", *user),
			templates.SiteStyles(),
			templates.SiteNavigation(search),
			templates.SiteContent(usrSites),
		).Render(r.Context(), w)
	}
}

// ShowEditSites displays the edit form to change the application JSON
func (t *TemplateHandler) ShowEditSites() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := ensureUser(r)
		search := ""
		t.Logger.InfoRequest(fmt.Sprintf("display the sites for user: '%s'", user.Username), r)

		usrSites, err := t.SiteSvc.GetSitesForUser(*user)
		if err != nil {
			t.Logger.ErrorRequest(fmt.Sprintf("could not get sites for user '%s'; '%v'", user.Username, err), r)
		}
		// for simple edit purposes we display the whole payload as JSON
		jsonPayload := tmpl.JsonIndent[sites.UserSites](usrSites)
		tmpl.Layout(
			t.layoutModel("Edit Apps", search, "/public/folder.svg", *user),
			templates.SiteEditStyles(),
			templates.SiteEditNavigation(search),
			templates.SiteEditContent(jsonPayload, ""),
		).Render(r.Context(), w)
	}
}

// SaveSites processed the provided JSON payload and saves the applications
func (t *TemplateHandler) SaveSites() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := ensureUser(r)
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
			templates.SiteEditContent(payload, fmt.Sprintf("the supplied JSON payload is invalid and cannot be processed; '%v'", err)).Render(r.Context(), w)
			return
		}

		err = t.SiteSvc.SaveSitesForUser(usrSites, *user)
		if err != nil {
			t.Logger.ErrorRequest(fmt.Sprintf("could not save the sites; '%v'", err), r)
			// return to edit page
			templates.SiteEditContent(payload, fmt.Sprintf("could not save the sites; '%v'", err)).Render(r.Context(), w)
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

func (t *TemplateHandler) layoutModel(pageTitle, search, favicon string, user security.User) tmpl.LayoutModel {
	appNav := make([]tmpl.NavItem, 0)
	var title string
	for _, a := range common.AvailableApps {
		if a.URL == "/sites" {
			a.Active = true
			title = a.DisplayName
		}
		appNav = append(appNav, a)
	}
	if pageTitle == "" {
		pageTitle = title
	}
	model := tmpl.LayoutModel{
		PageTitle:  pageTitle,
		Favicon:    favicon,
		Version:    t.versionString(),
		User:       user,
		Search:     search,
		Navigation: appNav,
	}

	if model.Favicon == "" {
		model.Favicon = "/public/folder.svg"
	}
	var jsReloadLogic string
	if t.Env == config.Development {
		jsReloadLogic = develop.PageReloadClientJS
	}
	model.PageReloadClientJS = tmpl.PageReloadClientJS(jsReloadLogic)
	return model
}

func ensureUser(r *http.Request) *security.User {
	user, ok := security.UserFromContext(r.Context())
	if !ok || user == nil {
		panic("the security context user is not available!")
	}
	return user
}
