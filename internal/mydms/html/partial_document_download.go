package html

import (
	_ "embed"
	"fmt"

	"golang.binggl.net/monorepo/pkg/handler/html"
	"golang.binggl.net/monorepo/pkg/handler/templates"
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

var removeLink g.Node = h.Div(h.Class("float-end"),
	h.I(h.Class("bi bi-x-lg")),
	h.A(h.Class("remove_document_link"),
		g.Attr("hx-delete", "/mydms/partial/upload"),
		g.Attr("hx-target", "#document_download_link"),
		g.Attr("hx-swap", "outerHTML"),
		g.Text("remove"),
	),
)

func DisplayDocumentDownload(doc Document) g.Node {
	var documentDownload g.Node

	if doc.ID != "" && doc.FileName.Val != "" {
		documentDownload = h.Div(h.Class("document_download"), h.ID("document_download_link"),
			h.I(h.Class("bi bi-cloud-arrow-down")),
			h.A(h.Class("document_download_link"), h.Href("/mydms/file/"+doc.PreviewLink.Val), h.Target("_NEW"), g.Text(doc.FileName.Val)),
			h.Input(h.Type("hidden"), h.Name("doc-tempID"), h.Value("-")),
			h.Input(h.Type("hidden"), h.Name("doc-filename"), h.Value(doc.FileName.Val)),
			removeLink,
		)

	} else if doc.UploadToken.Val != "" && doc.FileName.Val != "" {
		documentDownload = h.Div(h.Class("document_download"), h.ID("document_download_link"),
			h.I(h.Class("bi bi-cloud-arrow-down")),
			h.A(h.Class("document_download_link"), g.Text(doc.FileName.Val)),
			h.Input(h.Type("hidden"), h.Name("doc-tempID"), h.Value(doc.UploadToken.Val)),
			h.Input(h.Type("hidden"), h.Name("doc-filename"), h.Value(doc.FileName.Val)),
			removeLink,
		)
	} else {
		documentDownload = h.Div(h.ID("document_upload_area"),
			h.Div(h.Class("input-group mb-3"),
				h.Div(h.Class("dropdown"),
					h.Button(h.Type("button"),
						h.Class("btn btn-warning dropdown-toggle"),
						g.Attr("data-bs-toggle", "dropdown"),
						g.Attr("aria-expanded", "false"),
						g.Attr("data-bs-auto-close", "outside"),
						g.Text("Encryption"),
					),
					h.Div(h.Class("dropdown-menu p-4"),
						h.Div(h.Class("mb-3"),
							h.Label(h.For("initialPass"), h.Class("form-label"), g.Text("Password")),
							h.Input(h.Type("password"), h.Class("form-control"), h.ID("initialPass"), h.Name("doc-initPass"), h.Placeholder("initial")),
						),
						h.Div(h.Class("mb-3"),
							h.Label(h.For("pass"), h.Class("form-label"), g.Text("Password")),
							h.Input(h.Type("password"), h.Class("form-control"), h.ID("pass"), h.Name("doc-newPass"), h.Placeholder("new")),
						),
					),
				),
				g.Raw("&nbsp;"),
				h.Input(h.Class(html.ClassCond("form-control", "control_invalid", !doc.UploadToken.Valid)),
					h.Type("file"), h.Name("doc-fileupload"), h.ID("documentFileUpload"),
				),
				h.Button(
					h.Type("button"),
					h.ID("btn-doc-fileupload"),
					h.Class("btn btn-outline-secondary"),
					g.Attr("hx-post", "/mydms/upload"),
					g.Attr("hx-encoding", "multipart/form-data"),
					g.Attr("hx-trigger", "click"),
					g.Attr("hx-target", "#document_upload_area"),
					g.Attr("hx-swap", "outerHTML"),
					g.Attr("hx-params", "doc-fileupload,doc-initPass,doc-newPass"),
					g.Attr("hx-indicator", "#indicator"),
					h.I(h.Class("bi bi-upload")),
				),
			),
		)
	}
	return documentDownload
}

func DisplayTempDocumentUpload(fileName, tempID, errMsg string) g.Node {
	elements := []g.Node{
		h.I(h.Class("bi bi-cloud-arrow-down")),
		h.A(h.Class("document_download_link"), g.Text(fileName)),
		h.Input(h.Type("hidden"), h.Name("doc-tempID"), h.Value(tempID)),
		h.Input(h.Type("hidden"), h.Name("doc-filename"), h.Value(fileName)),
		removeLink,
	}

	return h.Div(h.Class("document_download"), h.ID("document_download_link"),
		g.If(errMsg == "",
			g.Group(elements),
		),
		g.If(errMsg != "",
			h.Div(h.Class("alert alert-danger"), h.Role("alert"),
				h.I(h.Class("bi bi-exclamation-triangle"), g.Text(fmt.Sprintf(" %s", templates.Ellipsis(errMsg, 60, "...")))),
				removeLink,
			),
		),
	)
}
