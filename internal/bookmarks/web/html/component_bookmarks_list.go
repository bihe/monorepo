package html

import (
	_ "embed"
	"fmt"

	"golang.binggl.net/monorepo/internal/bookmarks/app/bookmarks"
	"golang.binggl.net/monorepo/internal/common"
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

func displayBookmarkType(bm bookmarks.Bookmark, ell EllipsisValues) g.Node {
	var link g.Node

	switch bm.Type {
	case bookmarks.Node:
		link = h.A(
			h.Class("bookmark_name"),
			h.Href(bm.URL),
			h.Title(bm.DisplayName),
			h.Target("_blank"),
			g.Text(common.Ellipsis(bm.DisplayName, ell.NodeLen, "...")),
		)
	case bookmarks.FileItem:
		fileSize := 0
		if bm.FileMeta != nil {
			fileSize = bm.FileMeta.Size
		}
		link = h.A(
			h.Class("bookmark_name"),
			h.Href("/bm/GetBookmarkFile/"+bm.ID),
			h.Title(bm.DisplayName),
			h.Target("_blank"),
			g.Text(common.Ellipsis(bm.DisplayName, ell.NodeLen, "...")+formatFileSize(fileSize)),
		)
	case bookmarks.Folder:
		link = h.A(
			h.Class("bookmark_name"),
			h.Href("/bm/~"+common.EnsureTrailingSlash(bm.Path)+bm.DisplayName),
			h.Title(bm.DisplayName),
			g.Text(common.Ellipsis(bm.DisplayName, ell.FolderLen, "...")),
		)
	}
	return link
}

func formatFileSize(size int) string {
	const megabyte = 1024 * 1024
	if size < megabyte {
		// Format as kilobytes
		kb := float64(size) / 1024
		return fmt.Sprintf(" (%.1f KB)", kb)
	}
	// Format as megabytes
	mb := float64(size) / float64(megabyte)
	return fmt.Sprintf(" (%.2f MB)", mb)
}

//go:embed copyClipboard.min.js
var copyClipboard string

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
						displayBookmarkType(b, ell),
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
				)
			}),
		),
		g.El("script", g.Attr("type", "text/javascript"),
			g.Raw(copyClipboard),
		),
	)
}
