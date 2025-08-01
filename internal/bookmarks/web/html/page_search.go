package html

import (
	"fmt"

	"golang.binggl.net/monorepo/internal/bookmarks/app/bookmarks"
	"golang.binggl.net/monorepo/internal/common"
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

func SearchContent(search string, items []bookmarks.Bookmark, ell EllipsisValues) g.Node {
	return h.Div(h.Class("bookmark_list"),
		g.Attr("hx-get", "/bm/partial/search?q="+search),
		g.Attr("hx-trigger", "refreshBookmarkList from:body once"),
		g.Attr("hx-swap", "outerHTML"),
		g.Map(items, func(b bookmarks.Bookmark) g.Node {
			return h.Div(h.Class("bookmark_item hstack gap-3"),
				h.Div(h.Class("p2"),
					h.Span(
						h.Img(
							h.Width("24px"),
							h.Height("24px"),
							h.Alt("favicon"),
							g.If(b.InvertFaviconColor == 1,
								h.Class("bookmark_favicon invert"),
							),
							g.If(b.InvertFaviconColor == 0,
								h.Class("bookmark_favicon"),
							),
							h.Src(fmt.Sprintf("/bm/favicon/%s?t=%s", b.ID, b.TStamp())),
							h.Loading("lazy"),
						),
						g.If(b.ChildCount > 0,
							h.Span(
								h.Class("top-0 start-100 translate-middle badge rounded-pill bg-danger"),
								g.Text(fmt.Sprintf("%d", b.ChildCount)),
							),
						),
					),
					g.Text(" "),
					h.Span(
						h.Class("badge rounded-pill text-bg-secondary bookmark-path"),
						h.A(
							h.Class("bookmark_path"),
							h.Href("/bm/~"+b.Path),
							h.Title(b.Path),
							g.Text(common.Ellipsis(b.Path, ell.PathLen, "")),
						),
					),
					g.Text(" "),
					g.If(b.Type == bookmarks.Node, h.A(
						h.Class("bookmark_name"),
						h.Href(b.URL),
						h.Title(b.DisplayName),
						h.Target("_blank"),
						g.Text(common.Ellipsis(b.DisplayName, ell.NodeLen, "...")),
					)),
					g.If(b.Type == bookmarks.Folder, h.A(
						h.Class("bookmark_name"),
						h.Href("/bm/~"+common.EnsureTrailingSlash(b.Path)+b.DisplayName),
						h.Title(b.DisplayName),
						g.Text(common.Ellipsis(b.DisplayName, ell.FolderLen, "...")),
					)),
				),
				h.Div(h.Class("p2 ms-auto")),
				h.Div(h.Class("ps"),
					h.Div(h.Class("btn-group"), h.Role("group"),
						h.Button(h.Type("button"), h.Class("btn dropdown-toggle"), g.Attr("data-bs-toggle", "dropdown"), g.Attr("aria-expanded", "false")),
						h.Ul(h.Class("dropdown-menu"),
							h.Li(
								h.A(
									h.Class("dropdown-item"),
									h.ID("btn-bookmark-edit"),
									h.Href("#"),
									g.Attr("hx-target", "#modals-here"),
									g.Attr("hx-trigger", "click"),
									g.Attr("data-bs-toggle", "modal"),
									g.Attr("data-bs-target", "#modals-here"),
									g.Attr("hx-get", "/bm/"+b.ID),
									g.Attr("hx-swap", "innerHTML"),
									h.I(h.Class("bi bi-pencil"), g.Text(" Edit")),
								),
							),
							g.If(b.Type == bookmarks.Node,
								h.Li(
									h.A(
										h.Class("dropdown-item copy-clipboard-btn"),
										h.Href("#"),
										g.Attr("data-clipboard-text", b.URL),
										h.I(h.Class("bi bi-clipboard"), g.Text(" to Clipboard")),
									),
								),
							),
							h.Li(
								h.A(
									h.Class("dropdown-item delete"),
									h.ID("btn-bookmark-delete"),
									h.Href("#"),
									g.Attr("hx-target", "#modals-here"),
									g.Attr("hx-trigger", "click"),
									g.Attr("data-bs-toggle", "modal"),
									g.Attr("data-bs-target", "#modals-here"),
									g.Attr("hx-get", "/bm/confirm/delete/"+b.ID),
									g.Attr("hx-swap", "innerHTML"),
									h.I(h.Class("bi bi-x"), g.Text(" Delete")),
								),
							),
						),
					),
				),
				g.El("script", g.Attr("type", "text/javascript"),
					g.Raw(copyClipboard),
				),
			)
		}),
	)
}

func SearchStyles() g.Node {
	return g.El("style", g.Attr("type", "text/css"), g.Text(".delete{font-weight:bold;color:red}"))
}

func SearchNavigation(search string) g.Node {
	return h.Nav(h.Class("navbar navbar-expand application_name"),
		h.Div(h.Class("container-fluid"),
			h.A(h.Class("navbar-brand application_title"), h.Href("#"), h.I(h.Class("bi bi-bookmark-star"))),

			h.Div(h.Class("collapse navbar-collapse"),
				h.Ul(h.Class("navbar-nav me-auto"),
					h.Li(h.Class("nav-item"), h.A(h.Class("nav-link"),
						h.Div(g.Text("~ searching for: "),
							h.Span(h.Class("badge text-bg-success"), h.Style("font-size:small"), g.Text(search)),
						),
					)),
				),
			),
		),
	)
}
