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
			html.AgeContent(html.AgeModel{Passphrase: html.ValidatorInput{Valid: true}}),
			ageSearchURL,
		).Render(w)
	}
}

func (t *TemplateHandler) PerformAgeAction() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := common.EnsureUser(r)
		t.Logger.InfoRequest(fmt.Sprintf("perform page action for user: '%s'", user.Username), r)

		err := r.ParseForm()
		if err != nil {
			t.Logger.ErrorRequest(fmt.Sprintf("could not parse supplied form data; '%v'", err), r)
			t.RenderErr(r, w, fmt.Sprintf("could not parse supplied form data; '%v'", err))
			return
		}

		var (
			form       html.AgeModel
			formPrefix = "age_"
			passphrase string
			//inputText  string
			// outputText string
		)

		passphrase = r.FormValue(formPrefix + "passphrase")
		//inputText = r.FormValue(formPrefix + "passphrase")

		// form model and validation
		validData := true
		form.Passphrase = html.ValidatorInput{Val: passphrase, Valid: true}
		if passphrase == "" {
			form.Passphrase.Valid = false
			validData = false
		}

		if validData {
			// happy stuff
		}

		html.AgeContent(form).Render(w)
	}
}
