package html

import (
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

// DialogConfirmDeleteHx provides a confirmation dialog using htmx
func DialogConfirmDeleteHx(name, hxDeleteURL string) g.Node {
	return h.Div(h.Class("modal-dialog modal-dialog-centered"),
		h.Div(h.Class("modal-content"),
			h.Div(h.Class("modal-header"),
				h.H5(h.Class("modal-title"), g.Text("Confirm delete")),
			),
			h.Div(h.Class("modal-body"),
				h.P(
					g.Text("Do you really want to delete the item "),
					g.Text("'"),
					h.Strong(g.Text(name)),
					g.Text("'?"),
				),
			),
			h.Div(h.Class("modal-footer"),
				h.Button(h.Type("button"), h.Class("btn btn-secondary"), g.Attr("data-bs-dismiss", "modal"), g.Text("Close")),
				h.Button(
					h.Type("button"),
					h.ID("btn-confirm"),
					h.Class("btn btn-danger"),
					g.Attr("data-bs-dismiss", "modal"),
					g.Attr("hx-delete", hxDeleteURL),
					g.Text("Delete"),
				),
			),
		),
	)
}
