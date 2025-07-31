package web

import (
	"fmt"
	"net/http"

	"golang.binggl.net/monorepo/internal/common"
	"golang.binggl.net/monorepo/internal/core/app/agecrypt"
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

		model := html.AgeModel{
			Passphrase: html.ValidatorInput{Valid: true},
			InputText:  html.ValidatorInput{Valid: true},
			OutputText: html.ValidatorInput{Valid: true},
		}
		base.Layout(
			common.CreatePageModel("/age", "helpers to work with age", search, "/public/folder.svg", t.versionString(), t.Env, *user),
			html.AgeStyle(),
			html.AgeNavigation(search),
			html.AgeContent(model),
			ageSearchURL,
		).Render(w)
	}
}

// PerformAgeAction takes the provided input (passphrase, inputText, encryptedText) and encrypts/decrypts using age
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
			inputText  string
			outputText string
		)

		passphrase = r.FormValue(formPrefix + "passphrase")
		inputText = r.FormValue(formPrefix + "input")
		outputText = r.FormValue(formPrefix + "output")

		// form model and validation
		validData := true
		form.Passphrase = html.ValidatorInput{Val: passphrase, Valid: true}
		form.InputText = html.ValidatorInput{Val: inputText, Valid: true}
		form.OutputText = html.ValidatorInput{Val: outputText, Valid: true}

		if passphrase == "" {
			form.Passphrase.Valid = false
			form.Passphrase.Message = "a passphrase is needed"
			validData = false
		}

		if inputText == "" && outputText == "" {
			validData = false
			form.InputText.Valid = false
			form.OutputText.Valid = false
			form.InputText.Message = "both fields are empty"
			form.OutputText.Message = "both fields are empty"
		}

		if inputText != "" && outputText != "" {
			validData = false
			form.InputText.Valid = false
			form.OutputText.Valid = false
			form.InputText.Message = "both fields are filled / cannot decide what to do"
			form.OutputText.Message = "both fields are filled / cannot decide what to do"
		}

		if validData {
			// happy path
			if inputText != "" {
				encrypted, err := agecrypt.EncryptStringPassphrase(inputText, passphrase)
				if err != nil {
					t.Logger.ErrorRequest(fmt.Sprintf("cannot encrypt provided data; '%v'", err), r)

					form.InputText.Valid = false
					form.InputText.Message = err.Error()

					html.AgeContent(form).Render(w)
					return
				}
				form.OutputText.Val = encrypted
				form.OutputText.Valid = true

			} else if outputText != "" {
				decrypted, err := agecrypt.DecryptStringPassphrase(outputText, passphrase)
				if err != nil {
					t.Logger.ErrorRequest(fmt.Sprintf("cannot decrypt provided data; '%v'", err), r)

					form.OutputText.Valid = false
					form.OutputText.Message = err.Error()

					html.AgeContent(form).Render(w)
					return
				}
				form.InputText.Val = decrypted
				form.InputText.Valid = true
			}

			triggerToast(w, base.MsgSuccess, "Age in action", "Processed the provided input via age!")

			html.AgeContent(form).Render(w)
			return
		}
		triggerToast(w, base.MsgWarning, "Validation", "Invalid input provided!")

		html.AgeContent(form).Render(w)
	}
}
