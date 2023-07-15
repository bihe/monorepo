package site

import (
	"encoding/json"
	"html/template"
	"net/http"

	"golang.binggl.net/monorepo/internal/core/web"
	"golang.binggl.net/monorepo/pkg/logging"
)

// TemplateHandler defines the html templates and interaction used for the sites UI
type TemplateHandler struct {
	web.TemplateHandler
	ApiEndpoint string
}

// SitePageModel defines the data relevant for the sites UI
type SitePageModel struct {
	web.PageModel
	Sites UserSite
}

func (t TemplateHandler) GetSites() http.HandlerFunc {
	tmpl := template.Must(template.ParseFS(web.TemplateFS, "templates/_layout.html", "templates/sites.html"))
	return func(w http.ResponseWriter, r *http.Request) {
		data := t.GetPageModel(r)
		client := GetSiteClient(t.ApiEndpoint, data.User.Token)
		sites, err := client.GetSites()
		if err != nil {
			t.Logger.Error("could not get sites", logging.ErrV(err))
		}
		model := SitePageModel{
			PageModel: data,
			Sites:     *sites,
		}
		tmpl.Execute(w, model)
	}
}

func (t TemplateHandler) GetSitesEdit() http.HandlerFunc {
	tmpl := template.Must(template.ParseFS(web.TemplateFS, "templates/sites_edit.html"))
	return func(w http.ResponseWriter, r *http.Request) {
		data := t.GetPageModel(r)
		client := GetSiteClient(t.ApiEndpoint, data.User.Token)
		sites, err := client.GetSites()
		if err != nil {
			t.Logger.Error("could not get sites", logging.ErrV(err))
		}
		j, err := json.MarshalIndent(sites, "", "  ")
		if err != nil {
			t.Logger.Error("could not marshal sites", logging.ErrV(err))
		}
		tmpl.Execute(w, string(j))
	}
}
