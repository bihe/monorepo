package html

import (
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

func FaviconDialog() g.Node {
	return h.Div(h.ID("modal"), g.Attr("_", "on closeModal add .closing then wait for animationend then remove me"),
		h.Div(h.Class("modal-underlay"), g.Attr("_", "on click trigger closeModal")),
		h.Div(h.Class("modal-content-area"),
			h.Button(h.Class("btn danger"), g.Attr("_", "on click trigger closeModal"), g.Text("Close")),
		),
	)
}
