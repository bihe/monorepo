package html

import (
	_ "embed"
	"fmt"

	"golang.binggl.net/monorepo/internal/common"
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

type Document struct {
	ID            string
	Title         ValidStr
	Amount        ValidFloat
	FileName      ValidStr
	PreviewLink   ValidStr
	UploadToken   ValidStr
	Tags          ValidStrSlice
	Senders       ValidStrSlice
	InvoiceNumber ValidStr
	Error         string
	Close         bool
}

type ValidStr struct {
	Val     string
	Valid   bool
	Message string
}

type ValidStrSlice struct {
	Val     []string
	Valid   bool
	Message string
}

type ValidFloat struct {
	Val     float32
	Valid   bool
	Message string
}

//go:embed component_dialog_edit.css
var component_dialog_edit_styles string

// use the bootstrap logic to toggle the modal
const jsToggleModel = "bootstrap.Modal.getInstance('#modals-here').toggle();"

// initialize the tags javascript logic
const jsTagsInit = `
import Tags from "/public/js/tags/tags.min.js";
Tags.init();
`

func EditDocumentDialog(doc Document, docDownload g.Node) g.Node {
	if doc.Close {
		return g.El("script", g.Attr("type", "text/javascript"), g.Raw(jsToggleModel))
	}

	script := g.El("script", g.Attr("type", "module"), g.Raw(jsTagsInit))
	styles := g.El("style", g.Attr("type", "text/css"), g.Raw(component_dialog_edit_styles))

	emptyOption := func(text string) g.Node {
		return h.Option(h.Disabled(), g.Attr("hidden", "true"), h.Value(""), g.Text(text))
	}
	selectedOption := func(value string) g.Node {
		return h.Option(h.Value(value), h.Selected(), g.Text(value))
	}

	return h.Div(script, styles,
		h.Div(h.Class("modal-dialog modal-xl"), h.ID("document_edit_dialog"),
			h.Div(h.Class("modal-content"),
				h.Div(h.Class("modal-header"),
					g.If(doc.ID != "", h.H5(h.Class("modal-title"), g.Text("Edit: "+doc.Title.Val))),
					g.If(doc.ID == "", h.H5(h.Class("modal-title"), g.Text("Create Document"))),
					common.HtmxIndicatorNode(),
				),
				h.Form(h.Class("document_edit_form"),
					h.Input(h.Type("hidden"), h.Name("doc-id"), h.Value(doc.ID)),
					h.Div(h.Class("modal-body"),
						h.Div(h.Class("mb-3"),
							h.Div(h.Class("input-group"),
								h.Span(h.Class("input-group-text"), g.Text("Document Title")),
								h.Input(h.Type("text"), h.ID("document_title"), h.Placeholder("Document Title"), h.Name("doc-title"), h.Class(common.ClassCond("form-control", "control_invalid", !doc.Title.Valid)), h.Value(doc.Title.Val), h.Required()),
							),
						),
						h.Div(h.Class("row"),
							h.Div(h.Class("col"),
								h.Div(h.Class("input-group"),
									h.Span(h.Class("input-group-text"), g.Text("Amount")),
									h.Input(h.Type("text"), h.Class("form-control"), h.ID("document_amount"), h.Placeholder("Amount"), h.Name("doc-amount"), h.Value(fmt.Sprintf("%.f", doc.Amount.Val))),
									h.Span(h.Class("input-group-text"), g.Text("€")),
								),
							),
							h.Div(h.Class("col mb-3"),
								h.Div(h.Class("input-group"),
									h.Span(h.Class("input-group-text"), g.Text("Number")),
									h.Input(h.Type("text"), h.Class("form-control"), h.ID("document_number"), h.Placeholder("Invoice Number"), h.Name("doc-invoicenumber"), h.Value(doc.InvoiceNumber.Val)),
								),
							),
						),
						h.Div(h.Class("mb-3"),
							docDownload,
						),
						h.Div(h.Class("mb-3"),
							h.Div(h.Class(common.ClassCond("input-group", "control_invalid", !doc.Tags.Valid)),
								h.Span(h.Class(common.ClassCond("input-group-text", "control_invalid", !doc.Tags.Valid)), g.Text("#Tag")),
								h.Select(
									h.ID("tags-input"),
									h.Class("form-select"),
									h.Name("doc-tags[]"),
									h.Multiple(),
									g.Attr("data-allow-clear", "true"),
									g.Attr("data-allow-new", "true"),
									g.Attr("data-server", "/mydms/list/tags"),
									g.Attr("data-live-server", "1"),
									g.If(doc.Tags.Message != "", emptyOption(doc.Tags.Message)),
									g.If(doc.Tags.Message == "", emptyOption("Choose a tag...")),
									g.Map(doc.Tags.Val, func(t string) g.Node {
										return selectedOption(t)
									}),
								),
							),
							h.Div(h.Class("invalid-feedback"), g.Text("Please select a valid tag.")),
						),
						h.Div(h.Class("mb-3"),
							h.Div(h.Class(common.ClassCond("input-group", "control_invalid", !doc.Senders.Valid)),
								h.Span(h.Class(common.ClassCond("input-group-text", "control_invalid", !doc.Senders.Valid)), h.I(h.Class("bi bi-truck")), g.Text(" Sender")),
								h.Select(
									h.ID("senders-input"),
									h.Class("form-select"),
									h.Name("doc-senders[]"),
									h.Multiple(),
									g.Attr("data-allow-clear", "true"),
									g.Attr("data-allow-new", "true"),
									g.Attr("data-server", "/mydms/list/senders"),
									g.Attr("data-live-server", "1"),
									g.If(doc.Senders.Message != "", emptyOption(doc.Senders.Message)),
									g.If(doc.Senders.Message == "", emptyOption("Choose a sender...")),
									g.Map(doc.Senders.Val, func(t string) g.Node {
										return selectedOption(t)
									}),
								),
							),
							h.Div(h.Class("invalid-feedback"), g.Text("Please select a valid sender.")),
						),
						g.If(doc.Error != "", h.Div(h.Class("alert alert-danger"), h.Role("alert"), h.I(h.Class("bi bi-exclamation-triangle")), g.Text(" "+doc.Error))),
					),
					h.Div(h.Class("modal-footer"),
						h.Button(h.Type("button"), h.Class("btn btn-secondary"), g.Attr("data-bs-dismiss", "modal"), g.Text("Close")),
						h.Button(
							h.Type("button"),
							h.ID("btn-document-save"),
							h.Class("btn btn-success"),
							g.Attr("hx-post", "/mydms"),
							g.Attr("hx-target", "#document_edit_dialog"),
							g.Attr("hx-indicator", "#indicator"),
							g.Text("Save"),
						),
					),
				),
			),
		),
	)
}
