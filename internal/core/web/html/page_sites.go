package html

import (
	"golang.binggl.net/monorepo/internal/core/app/sites"
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

const sitesHeaderStyle = `
.header {
  background: darkred;
}

.header-search-field {
  background-color: #A02003;
  border: var(--bs-border-width) solid #731702;
  color: #dddddd;
}

.header-search-field:focus {
  border-color: yellow;
}

.header-search-field-prefix {
  background-color: #A02003;
  border: var(--bs-border-width) solid #731702;
}

.content_area {
    height: calc(100vh - var(--total-head-height));
    overflow-y: auto;
    padding-bottom: 0rem;
}`

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
	return h.StyleEl(
		h.Type("text/css"),
		g.Raw(sitesHeaderStyle),
	)
}

func SiteNavigation(search string) g.Node {

	return h.Nav(h.Class("navbar navbar-expand application_name"),
		h.Div(h.Class("container-fluid"),
			h.A(h.Class("navbar-brand application_title"), h.Href("#"), h.I(h.Class("bi bi-diagram-2"))),

			h.Div(h.Class("collapse navbar-collapse"),
				h.Ul(h.Class("navbar-nav me-auto"),
					h.Li(h.Class("nav-item"), h.A(h.Class("nav-link"), g.Text("> sites"))),
				),
				h.Form(
					h.Div(h.ID("request_indicator"), h.Class("request_indicator htmx-indicator"),
						h.Div(h.Class("spinner-border text-light"), h.Role("status"),
							h.Span(h.Class("visually-hidden"), g.Text("Loading...")),
						),
					),

					h.A(h.Href("/sites/edit"), h.Type("button"), h.Class("btn btn-primary"), h.I(h.Class("bi bi-pen")), g.Text(" Edit")),
				),
			),
		),
	)
}
