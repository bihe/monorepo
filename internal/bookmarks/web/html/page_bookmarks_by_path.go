package html

import (
	_ "embed"

	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

type BookmarkPathEntry struct {
	UrlPath     string
	DisplayName string
	LastItem    bool
}

func getPath(entries []BookmarkPathEntry) string {
	return entries[len(entries)-1].UrlPath
}

func BookmarksByPathStyles() g.Node {
	return h.StyleEl(
		h.Type("text/css"),
		g.Raw(".breadcrumb-item{--bs-breadcrumb-divider-color:#ffffff !important;--bs-breadcrumb-divider:'>';font-size:medium}.breadcrumb-item.active{color:#ffffff}li.breadcrumb-item > a{color:#ffffff}div.btn-group > button.btn.dropdown-toggle{--bs-btn-color:#ffffff}.delete{font-weight:bold;color:red}.right-action{position:absolute;right:20px}.sortInput{position:relative;top:18px}@media only screen and (min-device-width: 375px) and (max-device-width: 812px){.breadcrumb-item{--bs-breadcrumb-divider-color:#ffffff !important;--bs-breadcrumb-divider:'>';font-size:smaller}.breadcrumb-item.active{color:#ffffff}li.breadcrumb-item > a{color:#ffffff}}"),
	)
}

//go:embed sortingLogic.min.js
var sortingLogic string

func BookmarksByPathNavigation(entries []BookmarkPathEntry) g.Node {
	breadcrumbs := make([]g.Node, 0)
	for i, e := range entries {
		if e.LastItem {
			if i == 0 {
				breadcrumbs = append(breadcrumbs, h.Li(
					h.Class("breadcrumb-item active"),
					g.Attr("aria-current", "page"),
					h.I(h.Class("bi bi-house")),
				))
			} else {
				breadcrumbs = append(breadcrumbs, h.Li(
					h.Class("breadcrumb-item active"),
					g.Attr("aria-current", "page"),
					g.Text(e.DisplayName),
				))
			}
		} else {
			if i == 0 {
				breadcrumbs = append(breadcrumbs, h.Li(
					h.Class("breadcrumb-item"),
					h.A(
						h.Class("rootroot"),
						h.Href("/bm/~"+e.UrlPath),
						h.I(h.Class("bi bi-house")),
					),
				))
			} else {
				breadcrumbs = append(breadcrumbs, h.Li(
					h.Class("breadcrumb-item"),
					h.A(
						h.Href("/bm/~"+e.UrlPath),
						g.Text(e.DisplayName),
					),
				))
			}
		}
	}

	return h.Div(h.Class("application_name"),
		h.Nav(
			g.Attr("aria-label", "breadcrumb"),
			h.Ol(
				h.Class("breadcrumb"),
				g.Group(breadcrumbs),
			),
		),
		h.Span(
			h.Class("right-action"),
			h.Div(h.ID("request_indicator"), h.Class("request_indicator htmx-indicator"),
				h.Div(h.Class("spinner-border text-light"), h.Role("status"),
					h.Span(h.Class("visually-hidden"),
						g.Text("Loading..."),
					),
				),
			),
			h.Button(h.ID("btn_toggle_sorting"), h.Type("button"), g.Attr("data-bs-toggle", "button"), h.Class("btn sort_button"),
				h.I(h.Class("bi bi-arrow-down-up"),
					g.Text(" Sort"),
				),
			),
			h.Span(h.ID("save_list_sort_order"), h.Class("sort_button d-none"),
				h.Button(h.ID("btn_save_sorting"), h.Type("button"), h.Class("btn btn-success sort_button"),
					h.I(h.Class("bi bi-sort-numeric-down"),
						g.Text(" Save"),
					),
				),
			),
			h.Button(
				h.Type("button"),
				g.Attr("data-testid", "link-add-bookmark"),
				h.Class("btn btn-primary new_button"),
				g.Attr("data-bs-toggle", "modal"),
				g.Attr("data-bs-target", "#modals-here"),
				g.Attr("hx-target", "#modals-here"),
				g.Attr("hx-trigger", "click"),
				g.Attr("hx-get", "/bm/-1?path="+getPath(entries)),
				h.I(h.Class("bi bi-plus"),
					g.Text(" Add"),
				),
			),
		),
		g.El("script", g.Attr("type", "text/javascript"), g.Raw(sortingLogic)),
	)
}
