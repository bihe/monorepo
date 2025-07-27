package web

import (
	"fmt"
	"net/http"

	"golang.binggl.net/monorepo/internal/common"
	"golang.binggl.net/monorepo/internal/core/web/html"
	base "golang.binggl.net/monorepo/pkg/handler/html"
)

const ageSearchURL = "/age/search"

// DisplayAgeStartPage is used to show the start-page of the util app for age
func (t *TemplateHandler) DisplayAgeStartPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := common.EnsureUser(r)
		search := ""
		t.Logger.InfoRequest(fmt.Sprintf("display age start-page for user: '%s'", user.Username), r)

		base.Layout(
			common.CreatePageModel("/age", "helpers to work with age", search, "/public/folder.svg", t.versionString(), t.Env, *user),
			html.AgeStyle(),
			html.AgeNavigation(search),
			html.AgeContent(),
			ageSearchURL,
		).Render(w)
	}
}
