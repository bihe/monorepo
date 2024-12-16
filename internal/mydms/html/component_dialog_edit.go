package html

import (
	_ "embed"

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

//go:embed component_dialog_edit.min.js
var component_dialog_edit_script string

func EditDocumentDialog(doc Document, docDownload g.Node) g.Node {
	if doc.Close {
		return g.El("script", g.Attr("type", "text/javascript"), g.Raw("bootstrap.Modal.getInstance('#modals-here').toggle();"))
	}

	script := g.El("script", g.Attr("type", "module"), g.Raw(component_dialog_edit_script))
	styles := g.El("style", g.Attr("type", "text/css"), g.Raw(component_dialog_edit_styles))

	return h.Div(script, styles,
		h.Div(h.Class("modal-dialog modal-xl"), h.ID("document_edit_dialog")),
	)
}
