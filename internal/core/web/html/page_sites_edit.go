package html

import (
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

func SiteEditContent(payload, err string) g.Node {
	return h.Form(h.Name("jsonForm"), g.Attr("hx-post", "/sites"), g.Attr("hx-trigger", "saveApplications from:document"), g.Attr("hx-swap", "outerHTML"),
		h.Div(h.Class(""),
			g.If(err != "", h.Div(h.Class("alert alert-danger"), h.Role("alert"), g.Text(err))),
			h.Div(h.Class("form-floating"),
				h.Textarea(h.Class("form-control"), h.ID("json-edit-area"), h.Style("height: calc(100vh - 110px)"), h.Name("payload"), g.Raw(payload)),
			),
		),
	)
}

func SiteEditStyles() g.Node {
	return h.StyleEl(h.Type("text/css"), g.Raw(`.right-action {
	  position: absolute;
	  right: 20px;
	}`))
}

const javascriptContent = `
try {
  document.querySelector('#btn_save_sites').addEventListener('click', (event) => {
    htmx.trigger('#btn_save_sites', 'saveApplications');
  });
} catch(error) {
  console.error(error);
}
`

func SiteEditNavigation(search string) g.Node {
	return h.Div(h.Class("application_name"),
		h.Div(g.Raw("~ sites [edit]:")),
		h.Span(h.Class("right-action"),

			h.Div(h.ID("request_indicator"), h.Class("request_indicator htmx-indicator"),
				h.Div(h.Class("spinner-border text-light"), h.Role("status"),
					h.Span(h.Class("visually-hidden"), g.Text("Loading...")),
				),
			),
			h.A(h.Href("/sites"), h.Type("button"), h.Class("btn btn-secondary"), h.I(h.Class("bi bi-x")), g.Text(" Cancel")),
			g.Raw("&nbsp;"),
			h.Button(h.Type("button"), h.ID("btn_save_sites"), h.Class("btn btn-success"), h.I(h.Class("bi bi-save")), g.Text(" Save")),
		),
		h.Script(h.Type("text/javascript"), g.Raw(javascriptContent)),
	)
}