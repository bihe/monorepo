package web

import (
	"net/http"

	"golang.binggl.net/monorepo/internal/common"
)

// DisplayToastNotification shows a toast notification
func (t *TemplateHandler) DisplayToastNotification() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		msg := r.FormValue("notification_message")
		triggerToast(w,
			common.MsgInfo,
			"Information!",
			msg)
	}
}
