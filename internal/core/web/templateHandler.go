package web

import (
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

func (t *TemplateHandler) DisplaySites() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := ensureUser(r)
		search := ""
		t.Logger.InfoRequest(fmt.Sprintf("display the sites for user: '%s'", user.Username), r)

		sites, err := t.SiteSvc.GetSitesForUser(*user)
		if err != nil {
			t.Logger.ErrorRequest(fmt.Sprintf("could not get sites for user '%s'; '%v'", user.Username, err), r)
		}
		tmpl.Layout(
			t.layoutModel("Available Apps", search, "/public/folder.svg", *user),
			templates.SiteStyles(),
			templates.SiteNavigation(search),
			templates.SiteContent(sites),
		).Render(r.Context(), w)
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
