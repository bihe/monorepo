package web

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"golang.binggl.net/monorepo/internal/bookmarks/app/bookmarks"
	"golang.binggl.net/monorepo/internal/common"
	"golang.binggl.net/monorepo/pkg/handler"
	base "golang.binggl.net/monorepo/pkg/handler/html"
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
	App     *bookmarks.Application
	Version string
	Build   string
}

// --------------------------------------------------------------------------
//  Internals
// --------------------------------------------------------------------------

type triggerDef struct {
	base.ToastMessage
	Refresh string `json:"refreshBookmarkList,omitempty"`
}

// define a header for htmx to trigger events
// https://htmx.org/headers/hx-trigger/
const htmxHeaderTrigger = "HX-Trigger"

func (t *TemplateHandler) pageModel(pageTitle, searchStr, favicon string, user security.User) base.LayoutModel {
	return common.CreatePageModel("/bm", pageTitle, searchStr, favicon, t.Version, t.Build, t.Env, user)
}

func queryParam(r *http.Request, name string) string {
	keys, ok := r.URL.Query()[name]
	if !ok || len(keys[0]) < 1 {
		return ""
	}
	return keys[0]
}

func pathParam(r *http.Request, name string) string {
	return chi.URLParam(r, name)
}

func getIntFromString(val string) int {
	if val == "1" {
		return 1
	}
	return 0
}

func triggerRefreshWithToast(w http.ResponseWriter, errType, title, text string) {
	triggerEvent := triggerDef{
		ToastMessage: base.ToastMessage{
			Event: base.ToastMessageContent{
				Type:  errType,
				Title: title,
				Text:  text,
			},
		},
		Refresh: "now",
	}
	triggerJson := handler.Json(triggerEvent)
	w.Header().Add(htmxHeaderTrigger, triggerJson)
}

func triggerToast(w http.ResponseWriter, errType, title, text string) {
	triggerEvent := triggerDef{
		ToastMessage: base.ToastMessage{
			Event: base.ToastMessageContent{
				Type:  errType,
				Title: title,
				Text:  text,
			},
		},
	}
	triggerJson := handler.Json(triggerEvent)
	w.Header().Add(htmxHeaderTrigger, triggerJson)
}
