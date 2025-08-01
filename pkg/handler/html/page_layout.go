package html

import (
	_ "embed"

	"golang.binggl.net/monorepo/pkg/config"
	"golang.binggl.net/monorepo/pkg/develop"
	"golang.binggl.net/monorepo/pkg/security"

	g "maragu.dev/gomponents"
	c "maragu.dev/gomponents/components"
	h "maragu.dev/gomponents/html"
)

type LayoutModel struct {
	PageTitle  string
	Favicon    string
	Version    string
	User       security.User
	Search     string
	WindowX    int
	WindowY    int
	Navigation []NavItem
	Env        config.Environment
}

type NavItem struct {
	DisplayName string
	Icon        string
	URL         string
	Active      bool
}

//go:embed fragment_modals.html
var modalsHTMLFragment string

//go:embed fragment_toast.html
var toastHTMLFragment string

// Layout provides the basic HTML layout for content pages.
// The layout is currently based on bootstrap
func Layout(model LayoutModel, style, navigation, content g.Node, searchURL string) g.Node {
	return c.HTML5(c.HTML5Props{
		Title:    model.PageTitle,
		Language: "en",
		Head: []g.Node{
			h.Base(h.Href("/")),
			h.Meta(g.Attr("creator", "https://www.gomponents.com/")),
			// page icon definition
			h.Link(h.Rel("icon"), h.Href(model.Favicon), g.Attr("size", "48x48")),
			h.Link(h.Rel("shortcut icon"), h.ID("site-favicon"), h.Type("image/x-icon"), h.Href(model.Favicon)),
			// styles
			// TODO: extension point to combine/minify styles
			h.Link(h.Rel("stylesheet"), h.Href("/public/bootstrap/css/bootstrap.min.css")),
			h.Link(h.Rel("stylesheet"), h.Href("/public/css/styles.css")),
			h.Link(h.Rel("stylesheet"), h.Href("/public/fonts/local.css")),
			h.Link(h.Rel("stylesheet"), h.Href("/public/bootstrap-icons/bootstrap-icons.min.css")),
			style,
		},

		Body: []g.Node{
			h.Body(g.Attr("data-bs-theme", "dark"),
				h.Header(
					h.Nav(h.Class("navbar navbar-expand-lg navbar-dark fixed-top header"),
						h.Div(h.Class("container-fluid"),
							h.A(h.Class("navbar-brand"), h.Href("/"), h.I(h.Class("bi bi-1-square"))),

							h.Button(h.Class("navbar-toggler"), h.Type("button"), g.Attr("data-bs-toggle", "collapse"), g.Attr("data-bs-target", "#navbarCollapse"),
								h.Span(h.Class("navbar-toggler-icon")),
							),
							h.Div(h.Class("collapse navbar-collapse"), h.ID("navbarCollapse"),
								h.Ul(h.Class("navbar-nav me-auto mb-2 mb-lg-0"),
									g.Map(model.Navigation, func(n NavItem) g.Node {
										return h.Li(h.Class("nav-item"),
											h.A(g.If(!n.Active, h.Class("nav-link")), g.If(n.Active, h.Class("nav-link active")), h.Href(n.URL),
												g.Raw(n.Icon),
												h.Span(h.Class("hide_mobile"), g.Text(n.DisplayName)),
											),
										)
									}),
								),

								h.Form(h.Class("me-3"), h.Role("search"), h.Method("GET"), h.Action(searchURL),
									h.Div(h.Class("input-group"),
										h.Span(h.Class("input-group-text search-field-prefix"), h.ID("search-addon"), h.I(h.Class("bi bi-search"))),
										h.Input(h.Type("search"), h.Name("q"), h.Class("form-control search-field"), h.Placeholder("Search... (Ctrl+B)"), h.ID("search-field"), g.Attr("control-id", "search-field"), h.AutoComplete("off"), h.Value(model.Search)),
									),
								),
								g.Raw("&nbsp;"),
								h.Div(h.Class("application_version"), h.Span(h.Class("badge text-bg-warning"), h.I(h.Class("bi bi-git")), g.Textf(" %s", model.Version))),
								g.Raw("&nbsp;"),
								h.Span(h.Class("badge d-flex align-items-center p-1 pe-2 text-dark-emphasis bg-light-subtle border border-dark-subtle rounded-pill"), h.Img(h.Class("rounded-circle me-1"), h.Width("24"), h.Height("24"), h.Src(model.User.ProfileURL)),
									g.Text(model.User.DisplayName),
								),
							),
						),
					),
				),
				h.Section(h.Class("sub-navigation"),
					navigation,
				),
				h.Main(
					h.Div(h.Class("content_area"),
						content,
					),
					g.Raw(modalsHTMLFragment),
					g.Raw(toastHTMLFragment),
				),
				// scripts
				// TODO: extension point to combine/minify scripts
				h.Script(h.Src("/public/js/htmx.min.js")),
				h.Script(h.Src("/public/js/_hyperscript.min.js")),
				h.Script(h.Src("/public/bootstrap/js/popper.min.js")),
				h.Script(h.Src("/public/bootstrap/js/bootstrap.bundle.min.js")),
				h.Script(h.Src("/public/js/Sortable.min.js")),
				h.Script(h.Src("/public/js/script.js")),
				// only during development the live-reload feature should be available
				g.If(model.Env == config.Development,
					g.Raw(develop.PageReloadClientJS),
				),
			),
		},
	})
}
