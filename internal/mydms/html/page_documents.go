package html

import (
	_ "embed"

	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

func DocumentsContent(documentList g.Node, search string) g.Node {
	return h.Div(h.Class("container-fluid"),
		h.Form(
			h.Input(h.Type("hidden"), h.Name("q"), h.Value(search)),
			h.Div(
				h.Class("row be_my_center"),
				h.ID("document_list"),
				g.Attr("hx-put", "/mydms/partial/list"),
				g.Attr("hx-trigger", "refreshDocumentList from:body"),
				g.Attr("hx-params", "q"),
				g.Attr("hx-swap", "innerHTML"),
				documentList,
			),
		),
	)
}

func DocumentsNavigation(search string) g.Node {
	searchText := []g.Node{
		g.Text("~ mydms:"),
	}
	if search != "" {
		searchText = append(searchText, g.Text(">> searching for "), h.Span(h.Class("badge text-bg-success"), h.Style("font-size:small"), g.Text(search)))
	}
	return h.Div(h.Class("application_name"),
		h.Div(g.Group(searchText)),
		h.Span(h.Class("right-action"),
			h.Div(h.ID("request_indicator"), h.Class("request_indicator htmx-indicator"),
				h.Div(h.Class("spinner-border text-light"), h.Role("status"),
					h.Span(h.Class("visually-hidden"), g.Text("Loading...")),
				),
			),
			h.Button(h.Type("button"),
				h.Class("btn btn-primary new_button"),
				g.Attr("data-testid", "link-add-document"),
				g.Attr("data-bs-toggle", "modal"),
				g.Attr("data-bs-target", "#modals-here"),
				g.Attr("hx-target", "#modals-here"),
				g.Attr("hx-trigger", "click"),
				g.Attr("hx-post", "/mydms/dialog/NEW"),
				h.I(h.Class("bi bi-plus"), g.Text(" Add")),
			),
		),
	)
}

//go:embed page_documents.css
var page_documents_styles string

func DocumentsStyles() g.Node {
	return g.El("style", g.Attr("type", "text/css"), g.Raw(page_documents_styles))
}
