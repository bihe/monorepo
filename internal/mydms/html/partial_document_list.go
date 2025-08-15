package html

import (
	_ "embed"
	"fmt"

	"golang.binggl.net/monorepo/internal/common"
	"golang.binggl.net/monorepo/internal/mydms/app/document"
	"golang.binggl.net/monorepo/pkg/text"
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

//go:embed partial_document_list.css
var partial_document_list_styles string

func getDocumentPath(doc document.Document) string {
	if doc.PreviewLink != "" {
		return doc.PreviewLink
	}
	return text.EncBase64(doc.FileName)
}

func DocumentList(docNum, skip int, pd document.PagedDocument) g.Node {
	elements := make([]g.Node, 0)

	doclist := g.Map(pd.Documents, func(doc document.Document) g.Node {
		return h.Div(h.Class("card be_my_document"),
			h.Div(h.Class("card-body"),
				h.H5(h.Class("card-title"), h.Title(doc.Title),
					h.A(h.Href("/mydms/file/"+text.SafePathEscapeBase64(getDocumentPath(doc))), h.Target("_NEW"),
						h.I(h.Class("bi bi-cloud-download")),
					),
					g.Text(" "),
					h.Span(
						h.Class("edit_document"),
						g.Attr("data-testid", "edit-document"),
						g.Attr("data-bs-toggle", "modal"),
						g.Attr("data-bs-target", "#modals-here"),
						g.Attr("hx-post", "/mydms/dialog/"+doc.ID),
						g.Attr("hx-target", "#modals-here"),
						g.Attr("hx-trigger", "click"),
						g.Text(common.Ellipsis(doc.Title, 23, "~")),
					),
				),
				h.Div(h.Class("btn-group card_menu"), h.Role("group"),
					h.Button(
						h.Type("button"),
						h.Class("btn dropdown-toggle"),
						g.Attr("data-bs-toggle", "dropdown"),
						g.Attr("aria-expanded", "false"),
					),
					h.Ul(h.Class("dropdown-menu"),
						h.Li(
							h.A(
								h.Class("dropdown-item delete"),
								h.ID("btn-document-delete"),
								h.Href("#"),
								g.Attr("data-bs-toggle", "modal"),
								g.Attr("data-bs-target", "#modals-here"),
								g.Attr("hx-post", "/mydms/confirm/"+doc.ID),
								g.Attr("hx-target", "#modals-here"),
								g.Attr("hx-trigger", "click"),
								g.Attr("hx-swap", "innerHTML"),
								h.I(h.Class("bi bi-x"), g.Text(" Delete")),
							),
						),
					),
				),
				g.If(doc.Amount != 0, h.Span(h.Class("amount"), g.Text(fmt.Sprintf("€ %.2f", doc.Amount)))),
			),
			h.Div(h.Class("card-body doc-content"),
				g.If(doc.InvoiceNumber != "", h.Span(h.Class("invoice-number"), h.I(h.Class("bi bi-123")), g.Text(doc.InvoiceNumber))),
				g.If(doc.InvoiceNumber == "", h.Span(h.Class("invoice-number"), g.Text("-"))),
			),
			h.Div(h.Class("card-body"),
				h.Div(h.Class("tags"),
					g.Map(doc.Tags, func(t string) g.Node {
						return h.Span(h.Class("badge text-bg-secondary tag"), g.Text(fmt.Sprintf("#%s", t)))
					}),
				),
				h.Div(h.Class("senders"),
					g.Map(doc.Senders, func(s string) g.Node {
						return h.Span(h.Class("badge text-bg-light tag"),
							h.A(h.Title(s),
								h.I(h.Class("bi bi-truck")),
								g.Text(fmt.Sprintf(" %s", common.Ellipsis(s, 30, "~"))),
							),
						)
					}),
				),
				h.Div(h.Class("meta"),
					h.Span(
						g.Text("c:"),
						h.Span(h.Class("meta_date"), g.Text(common.SubString(doc.Created, 10))),
						g.Iff(doc.Modified != "", func() g.Node {
							children := []g.Node{
								h.Br(),
								g.Text("m:"),
								h.Span(h.Class("meta_date"), g.Text(common.SubString(doc.Modified, 10))),
							}
							return g.Group(children)

						}),
					),
				),
			),
		)
	})
	elements = append(elements, doclist)

	more := h.Div(h.ID("page_content"), h.Class("show_more_results"),
		g.If(docNum == 0, h.Div(h.Class("center_aligned"),
			h.P(h.Class("noitems"), h.I(h.Class("bigger bi bi-balloon")), g.Text(" No results available!")),
		)),
		g.If(docNum > 0,
			h.Div(
				h.Input(h.Type("hidden"), h.Name("skip"), h.Value(fmt.Sprintf("%d", skip))),
				h.P(g.Text(fmt.Sprintf("Currently showing %d results of total %d", docNum, pd.TotalEntries))),
				h.Div(h.ID("request_indicator"), h.Class("request_indicator htmx-indicator"),
					h.Div(h.Class("spinner-border text-light"), h.Role("status"),
						h.Span(h.Class("visually-hidden"), g.Text("Loading...")),
					),
				),
				g.If(skip > 0,
					h.Button(
						h.Type("button"),
						h.Class("btn btn-light btn-sm"),
						g.Attr("hx-put", "/mydms/partial/list"),
						g.Attr("hx-target", "#page_content"),
						g.Attr("hx-swap", "outerHTML"),
						g.Attr("hx-params", "q,skip"),
						g.Text("..."),
					),
				),
			),
		),
	)
	elements = append(elements, more)

	style := g.El("style", g.Attr("type", "text/css"), g.Raw(partial_document_list_styles))
	elements = append(elements, style)

	return g.Group(elements)
}
