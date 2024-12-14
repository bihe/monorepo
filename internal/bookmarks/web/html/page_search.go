package html

import (
	"fmt"

	"golang.binggl.net/monorepo/internal/bookmarks/app/bookmarks"
	"golang.binggl.net/monorepo/pkg/handler/html"
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

func SearchContent(items []bookmarks.Bookmark, ell EllipsisValues) g.Node {
	return h.Div(h.Class("bookmark_list"),
		g.Map(items, func(b bookmarks.Bookmark) g.Node {
			return h.Div(h.Class("bookmark_item"),
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
						g.Text(html.Ellipsis(b.Path, ell.PathLen, "")),
					),
				),
				g.Text(" "),
				g.If(b.Type == bookmarks.Node, h.A(
					h.Class("bookmark_name"),
					h.Href(b.URL),
					h.Title(b.DisplayName),
					g.Text(html.Ellipsis(b.DisplayName, ell.NodeLen, "...")),
				)),
				g.If(b.Type == bookmarks.Folder, h.A(
					h.Class("bookmark_name"),
					h.Href("/bm/~"+html.EnsureTrailingSlash(b.Path)+b.DisplayName),
					h.Title(b.DisplayName),
					g.Text(html.Ellipsis(b.DisplayName, ell.FolderLen, "...")),
				)),
			)
		}),
	)
}

func SearchStyles() g.Node {
	return g.Text("")
}

func SearchNavigation(search string) g.Node {
	return h.Div(h.Class("application_name"),
		h.Div(g.Text("~ searching for:"),
			h.Span(h.Class("badge text-bg-success"), h.Style("font-size:small"), g.Text(search)),
		),
	)
}
