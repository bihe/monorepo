package html

import (
	"encoding/base64"
	"fmt"
	"strings"

	"golang.binggl.net/monorepo/internal/bookmarks/app/bookmarks"
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

func base64enc(input string) string {
	data := []byte(input)
	encodedString := base64.StdEncoding.EncodeToString(data)
	return encodedString
}

func isHtmlLike(payload []byte) bool {
	input := string(payload)
	return strings.Contains(input, "<html")
}

func FaviconDialog(currFaviconID string, favicons []bookmarks.ObjectInfo) g.Node {
	return h.Div(h.ID("modal"), g.Attr("_", "on closeModal add .closing then wait for animationend then remove me"),
		h.Div(h.Class("modal-underlay"), g.Attr("_", "on click trigger closeModal")),
		h.Div(h.Class("modal-content-area"),
			h.Div(h.ID("favicon_grid"), h.Class("favicon_grid_container"),
				g.Map(favicons, func(f bookmarks.ObjectInfo) g.Node {
					if isHtmlLike(f.Payload) {
						// unfortunately we have stored some BS, so do not display it
						return g.Text("")
					}
					var classNode g.Node = h.Class("favicon_view")
					if currFaviconID == f.Name {
						classNode = h.Class("favicon_view_selected")
					}

					return h.Img(
						g.Attr("hx-get", fmt.Sprintf("/bm/favicon/select/%s", base64enc(f.Name))),
						g.Attr("hx-trigger", "click"),
						g.Attr("hx-target", "#bookmark_favicon_display"),
						g.Attr("hx-swap", "outerHTML"),
						g.Attr("_", "on click trigger closeModal"),
						h.Width("42px"), h.Height("42px"), h.Title(fmt.Sprintf("payload-size: %d", len(f.Payload))), classNode, h.Alt("fi"), h.Src(fmt.Sprintf("/bm/favicon/raw/%s?t=%d", base64enc(f.Name), f.Modified.Nanosecond())), h.Loading("lazy"),
					)
				}),
			),
			h.Div(h.Class("mx-auto p-2"), h.Style("width: 90px;"),
				h.Button(h.Class("btn btn-secondary"), h.Style("width: 90px;"), g.Attr("_", "on click trigger closeModal"), g.Text("Close")),
			),
		),
	)
}
