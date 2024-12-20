package html

import (
	"golang.binggl.net/monorepo/internal/core/app/sites"
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

func SiteContent(userSites sites.UserSites) g.Node {
	return h.Div(h.Class("container-fluid"),
		h.Div(h.Class("row"),
			g.Map(userSites.Sites, func(site sites.SiteInfo) g.Node {
				return h.Div(h.Class("card application"),
					h.Div(h.Class("card-body"),
						h.H5(h.Class("card-title"), g.Text(site.Name)),
						h.P(
							g.Map(site.Perm, func(p string) g.Node {
								return h.Span(h.Class("badge text-bg-info permission"), g.Textf("#%s", p))
							}),
						),
						h.Span(h.Class("badge text-bg-light"), g.Text(site.URL)),
					),
				)
			}),
		),
	)
}

func SiteStyles() g.Node {
	return g.Text("")
}

func SiteNavigation(search string) g.Node {
	return h.Div(h.Class("application_name"),
		h.Div(g.Text("~ sites:")),
		h.Span(h.Class("right-action"),
			h.Div(h.ID("request_indicator"), h.Class("request_indicator htmx-indicator"),
				h.Div(h.Class("spinner-border text-light"), h.Role("status"),
					h.Span(h.Class("visually-hidden"), g.Text("Loading...")),
				),
			),
			h.A(h.Href("/sites/edit"), h.Type("button"), h.Class("btn btn-light"), h.I(h.Class("bi bi-pen")), g.Text(" Edit")),
		),
	)
}
