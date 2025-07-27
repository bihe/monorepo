package html

import (
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

func AgeContent() g.Node {
	return h.Div(h.Class("container-fluid"),
		h.Div(h.Class("row"),

			h.P(h.Class("mb-3 page_label"),
				g.Text("To encrypt and decrypt content the tool "), h.A(h.Href("https://github.com/FiloSottile/age"), g.Text("age")), g.Text(" is used. The process is using a passphrase to simplify the overall interaction. The passphrase needs to be remembered to decrypt a given input."),
			),

			h.Div(h.Class("mb-3"),
				h.Label(h.For("age_passphrase"), h.Class("form-label"), g.Text("Passphrase: ")),
				h.Input(h.Type("test"), h.Class("form-control"), h.ID("age_passphrase"), h.Placeholder("passphrase")),
			),

			h.Div(h.Class("mb-3"),
				h.Label(h.For("age_input"), h.Class("form-label"), g.Text("Input: ")),
				h.Textarea(h.Class("form-control"), h.ID("age_input"), h.Placeholder("raw unencrypted text"), h.Rows("5")),
			),

			h.Div(h.Class("mb-3"),
				h.Label(h.For("age_output"), h.Class("form-label"), g.Text("Encrypted: ")),
				h.Textarea(h.Class("form-control"), h.ID("age_output"), h.Placeholder("encrypted text"), h.Rows("5")),
			),
		),
	)
}

func AgeStyle() g.Node {
	return h.StyleEl(
		h.Type("text/css"),
		g.Raw(".page_label { margin-top: 10px;}"),
	)

}

func AgeNavigation(search string) g.Node {
	return h.Div(h.Class("application_name"),
		h.Div(g.Text("~ age:")),
		h.Span(h.Class("right-action"),
			h.Div(h.ID("request_indicator"), h.Class("request_indicator htmx-indicator"),
				h.Div(h.Class("spinner-border text-light"), h.Role("status"),
					h.Span(h.Class("visually-hidden"), g.Text("Loading...")),
				),
			),
			h.A(h.Href("/age/action"), h.Type("button"), h.Class("btn btn-primary"), h.I(h.Class("bi bi-nut")), g.Text(" Go")),
		),
	)
}
