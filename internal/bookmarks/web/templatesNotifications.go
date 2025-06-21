package web

import (
	"net/http"

	"golang.binggl.net/monorepo/pkg/handler/html"
)

// DisplayToastNotification shows a toast notification
// This helper action can be used to trigger a toast-message
// from "somewhere" in the application logic. Currently one use-case is to trigger
// a toast-message from Javascript via a htmx.ajax (copyClipboard)
func (t *TemplateHandler) DisplayToastNotification() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		msg := r.FormValue("notification_message")
		triggerToast(w,
			html.MsgInfo,
			"Information!",
			msg)
		// there is no other content than the header-information
		// show this in a proper way by using http 204
		w.WriteHeader(http.StatusNoContent)
	}
}
