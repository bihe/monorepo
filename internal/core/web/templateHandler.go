package web

import (
	"fmt"
	"net/http"

	"golang.binggl.net/monorepo/internal/common/crypter"
	"golang.binggl.net/monorepo/internal/core/app/sites"
	"golang.binggl.net/monorepo/pkg/handler"
	"golang.binggl.net/monorepo/pkg/handler/html"
)

// TemplateHandler takes care of providing HTML templates.
// This is the new approach with a template + htmx based UI to replace the angular frontend
// and have a more go-oriented approach towards UI and user-interaction. This reduces the
// cognitive load because less technology mix is needed. Via template + htmx approach only
// a limited amount of javascript is needed to achieve the frontend.
// As additional benefit the build should be faster, because the nodejs build can be removed
type TemplateHandler struct {
	*handler.TemplateHandler
	SiteSvc    sites.Service
	CrypterSvc crypter.EncryptionService
	Version    string
	Build      string
}

// DisplayToastNotification shows a toast notification
// This helper action can be used to trigger a toast-message
// from "somewhere" in the application logic. Currently one use-case is to trigger
// a toast-message from Javascript via a htmx.ajax (copyClipboard)
func (t *TemplateHandler) DisplayToastNotification() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		msg := r.FormValue("notification_message")
		msgType := r.FormValue("notification_type")

		toastType := html.MsgInfo
		toastTitle := "Information!"

		switch msgType {
		case "error":
			toastType = html.MsgError
			toastTitle = "Error!"
		}

		triggerToast(w,
			toastType,
			toastTitle,
			msg)
		// there is no other content than the header-information
		// show this in a proper way by using http 204
		w.WriteHeader(http.StatusNoContent)
	}
}

// --------------------------------------------------------------------------
//  Internals
// --------------------------------------------------------------------------

type triggerDef struct {
	html.ToastMessage
	Refresh string `json:"refreshBookmarkList,omitempty"`
}

// define a header for htmx to trigger events
// https://htmx.org/headers/hx-trigger/
const htmxHeaderTrigger = "HX-Trigger"

func triggerToast(w http.ResponseWriter, errType, title, text string) {
	triggerEvent := triggerDef{
		ToastMessage: html.ToastMessage{
			Event: html.ToastMessageContent{
				Type:  errType,
				Title: title,
				Text:  text,
			},
		},
	}
	triggerJson := handler.Json(triggerEvent)
	w.Header().Add(htmxHeaderTrigger, triggerJson)
}

func (t *TemplateHandler) versionString() string {
	return fmt.Sprintf("%s-%s", t.Version, t.Build)
}
