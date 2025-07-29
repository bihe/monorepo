package html

import (
	"golang.binggl.net/monorepo/internal/common"
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

type ValidatorInput struct {
	Val     string
	Valid   bool
	Message string
}

type AgeModel struct {
	Passphrase ValidatorInput
	InputText  ValidatorInput
	OutputText ValidatorInput
}

const changePasswordJS = `
try {
  document.querySelector('#toggle_age_passphrase').addEventListener('click', (event) => {
    	let x = document.getElementById("age_passphrase");
	if (x.type === "password") {
		x.type = "text";
	} else {
		x.type = "password";
	}
  });
} catch(error) {
  console.error(error);
}
`

func AgeContent(model AgeModel) g.Node {
	return h.Div(h.ID("age_content_area"), h.Class("container-fluid"),
		h.Div(h.Class("row"),
			h.Form(g.Attr("hx-post", "/age"), g.Attr("hx-trigger", "performAgeAction from:document"), g.Attr("hx-swap", "outerHTML"), g.Attr("hx-indicator", "#request_indicator"),
				h.P(h.Class("mb-3 page_label"),
					g.Text("To encrypt and decrypt content the tool "), h.A(h.Href("https://github.com/FiloSottile/age"), g.Text("age")), g.Text(" is used. The process is using a passphrase to simplify the overall interaction. The passphrase needs to be remembered to decrypt a given input."),
				),

				h.Div(h.Class("mb-3"),
					h.Label(h.For("age_passphrase"), h.Class("form-label"), g.Text("Passphrase: ")),

					h.Div(h.Class("input-group mb-3"),
						h.Input(
							h.Type("password"),
							h.ID("age_passphrase"),
							h.Placeholder("passphrase"),
							h.Class(common.ClassCond("form-control", "control_invalid", !model.Passphrase.Valid)),
							h.Name("age_passphrase"),
							h.Value(model.Passphrase.Val),
						),
						h.Button(
							h.ID("toggle_age_passphrase"),
							h.Class("btn btn-outline-secondary"),
							h.Type("button"),
							h.I(h.Class("bi bi-eye")),
							g.Attr("data-bs-toggle", "button"),
						),
					),
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
			h.Script(h.Type("text/javascript"), g.Raw(changePasswordJS)),
		),
	)
}

func AgeStyle() g.Node {
	return h.StyleEl(
		h.Type("text/css"),
		g.Raw(".page_label { margin-top: 10px;}"),
	)

}

const triggerAgeAction = `
try {
  document.querySelector('#age_perform_action').addEventListener('click', (event) => {
    htmx.trigger('#age_perform_action', 'performAgeAction');
  });
} catch(error) {
  console.error(error);
}
`

func AgeNavigation(search string) g.Node {
	return h.Div(h.Class("application_name"),
		h.Div(g.Text("~ age:")),
		h.Span(h.Class("right-action"),

			h.Div(h.ID("request_indicator"), h.Class("request_indicator htmx-indicator"),
				h.Div(h.Class("spinner-border text-light"), h.Role("status"),
					h.Span(h.Class("visually-hidden"), g.Text("Loading...")),
				),
			),

			h.Button(
				h.Type("button"),
				h.ID("age_perform_action"),
				h.Class("btn btn-primary"),
				h.I(h.Class("bi bi-nut")),
				g.Text(" Go"),
			),
		),
		h.Script(h.Type("text/javascript"), g.Raw(triggerAgeAction)),
	)
}
