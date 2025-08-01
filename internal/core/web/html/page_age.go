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
	return h.Div(h.ID("age_content_area"), h.Class("container-fluid age_content"), g.Attr("data-bs-theme", "light"),
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
						g.If(!model.Passphrase.Valid, h.Div(h.Class("invalid_input"), g.Text(model.Passphrase.Message))),
					),
				),

				h.Div(h.Class("mb-3"),
					h.Label(h.For("age_input"), h.Class("form-label"), g.Text("Input: ")),
					h.Textarea(
						h.Class(common.ClassCond("form-control", "control_invalid", !model.InputText.Valid)),
						h.ID("age_input"),
						h.Name("age_input"),
						h.Placeholder("raw unencrypted text"),
						h.Rows("10"),
						g.Raw(model.InputText.Val),
					),
					g.If(!model.InputText.Valid, h.Div(h.Class("invalid_input"), g.Text(model.InputText.Message))),
				),

				h.Div(h.Class("mb-3"),
					h.Label(h.For("age_output"), h.Class("form-label"), g.Text("Encrypted: ")),
					h.Textarea(
						h.Class(common.ClassCond("form-control", "control_invalid", !model.OutputText.Valid)),
						h.ID("age_output"),
						h.Name("age_output"),
						h.Placeholder("encrypted text"),
						h.Rows("10"),
						g.Raw(model.OutputText.Val),
					),
					g.If(!model.OutputText.Valid, h.Div(h.Class("invalid_input"), g.Text(model.OutputText.Message))),
				),
			),
			h.Script(h.Type("text/javascript"), g.Raw(changePasswordJS)),
		),
	)
}

func AgeStyle() g.Node {
	return h.StyleEl(
		h.Type("text/css"),
		g.Raw(".page_label { margin-top: 10px;font-size:large;}"),
		g.Raw(".age_content{height:100%;background-color:#F2F2F2;color:black}"),
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

	return h.Nav(h.Class("navbar navbar-expand application_name"),
		h.Div(h.Class("container-fluid"),
			h.A(h.Class("navbar-brand application_title"), h.Href("#"), h.I(h.Class("bi bi-file-lock"))),

			h.Div(h.Class("collapse navbar-collapse"),
				h.Ul(h.Class("navbar-nav me-auto"),
					h.Li(h.Class("nav-item"), h.A(h.Class("nav-link"), g.Text("> age"))),
				),
				h.Form(
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
			),
		),
	)
}
