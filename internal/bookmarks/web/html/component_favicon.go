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

// IDEA: click on favicon - take over the payload ID and instruct htmx to use the given ID as
// the favicon (and display it); use the given ID as the hidden id of the favIconImage @see FetchCustomFaviconURL

func FaviconDialog(favicons []bookmarks.ObjectInfo) g.Node {
	return h.Div(h.ID("modal"), g.Attr("_", "on closeModal add .closing then wait for animationend then remove me"),
		h.Div(h.Class("modal-underlay"), g.Attr("_", "on click trigger closeModal")),
		h.Div(h.Class("modal-content-area"),
			h.Div(h.ID("favicon_grid"),
				g.Map(favicons, func(f bookmarks.ObjectInfo) g.Node {
					if isHtmlLike(f.Payload) {
						return g.Text("")
					}
					return h.Img(h.Width("42px"), h.Height("42px"), h.Title(fmt.Sprintf("payload-size: %d", len(f.Payload))), h.Class("favicon_view"), h.Alt("fi"), h.Src(fmt.Sprintf("/bm/favicon/raw/%s?t=%d", base64enc(f.Name), f.Modified.Nanosecond())), h.Loading("lazy"))
				}),
			),
			h.Div(h.Class("mx-auto p-2"), h.Style("width: 90px;"),
				h.Button(h.Class("btn btn-secondary"), h.Style("width: 90px;"), g.Attr("_", "on click trigger closeModal"), g.Text("Close")),
			),
		),
	)
}
