package html

import (
	"fmt"

	"golang.binggl.net/monorepo/internal/bookmarks/app/bookmarks"
	"golang.binggl.net/monorepo/pkg/handler/html"
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

func getFaviconClass(b bookmarks.Bookmark) string {
	faviconClass := "bookmark_favicon"
	if b.InvertFaviconColor == 1 {
		faviconClass += " invert"
	}
	return faviconClass
}

func BookmarkList(path string, items []bookmarks.Bookmark, ell EllipsisValues) g.Node {
	return h.Div(h.Class("bookmark_list"), h.ID("bookmark_list"),
		g.Attr("hx-get", "/bm/partial/~"+path),
		g.Attr("hx-trigger", "refreshBookmarkList from:body once"),
		g.Attr("hx-swap", "outerHTML"),
		h.Form(h.Name("sortform"), h.Class("sortable"),
			g.Attr("hx-post", "/bm/sort"),
			g.Attr("hx-trigger", "sortBookmarkList from:body once"),
			g.Attr("hx-swap", "none"),
			g.If(len(items) == 0, h.Div(h.Class("no_bookmarks"),
				g.Text("no entries available"),
			)),
			g.Map(items, func(b bookmarks.Bookmark) g.Node {
				return h.Div(h.Class("bookmark_item hstack gap-3"),
					h.Div(h.Class("p2"),
						h.Span(
							h.Img(h.Width("24px"), h.Height("24px"), h.Alt("favicon"), h.Class(getFaviconClass(b)), h.Src(fmt.Sprintf("/bm/favicon/%s?t=%s", b.ID, b.TStamp())), h.Loading("lazy")),
							g.If(b.ChildCount > 0, h.Span(h.Class("top-0 start-100 translate-middle badge rounded-pill bg-danger"), g.Text(fmt.Sprintf("%d", b.ChildCount)))),
						),
						g.Text(" "),
						g.If(b.Type == bookmarks.Node, h.A(h.Class("bookmark_name"), h.Href(b.URL), h.Title(b.DisplayName), g.Text(html.Ellipsis(b.DisplayName, ell.NodeLen, "...")))),
						g.If(b.Type == bookmarks.Folder, h.A(h.Class("bookmark_name"), h.Href("/bm/~"+html.EnsureTrailingSlash(b.Path)+b.DisplayName), h.Title(b.DisplayName), g.Text(html.Ellipsis(b.DisplayName, ell.FolderLen, "...")))),
						h.Input(h.Type("hidden"), h.Name("ID"), h.Value(b.ID)),
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
				)
			}),
		),
	)
}
