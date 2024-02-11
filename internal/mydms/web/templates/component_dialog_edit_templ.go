// Code generated by templ - DO NOT EDIT.

// templ: version: v0.2.543
package templates

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import "github.com/a-h/templ"
import "context"
import "io"
import "bytes"

import "golang.binggl.net/monorepo/internal/mydms/app/document"
import "golang.binggl.net/monorepo/pkg/handler/templates"
import "fmt"

func EditDocumentDialog(doc document.Document) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, templ_7745c5c3_W io.Writer) (templ_7745c5c3_Err error) {
		templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := templ_7745c5c3_W.(*bytes.Buffer)
		if !templ_7745c5c3_IsBuffer {
			templ_7745c5c3_Buffer = templ.GetBuffer()
			defer templ.ReleaseBuffer(templ_7745c5c3_Buffer)
		}
		ctx = templ.InitializeContext(ctx)
		templ_7745c5c3_Var1 := templ.GetChildren(ctx)
		if templ_7745c5c3_Var1 == nil {
			templ_7745c5c3_Var1 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<div class=\"modal-dialog modal-xl\" id=\"document_edit_dialog\"><div class=\"modal-content\"><!-- header --><div class=\"modal-header\">")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		if doc.ID != "" {
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<h5 class=\"modal-title\">Edit: ")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			var templ_7745c5c3_Var2 string
			templ_7745c5c3_Var2, templ_7745c5c3_Err = templ.JoinStringErrs(doc.Title)
			if templ_7745c5c3_Err != nil {
				return templ.Error{Err: templ_7745c5c3_Err, FileName: `component_dialog_edit.templ`, Line: 12, Col: 46}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var2))
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</h5>")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
		} else {
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<h5 class=\"modal-title\">Create Document</h5>")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<div id=\"indicator\" class=\"htmx-indicator\"><div class=\"spinner-border text-light\" role=\"status\"><span class=\"visually-hidden\">Loading...</span></div></div></div><form class=\"document_edit_form\"><input type=\"hidden\" name=\"doc-id\" value=\"")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(doc.ID))
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("\"><!-- body --><div class=\"modal-body\"><div class=\"mb-3\"><div class=\"input-group\"><span class=\"input-group-text\">Document Title</span> <input class=\"form-control\" id=\"document_title\" placeholder=\"Document Title\" name=\"doc-title\" value=\"")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(doc.Title))
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("\"></div></div><div class=\"row\"><div class=\"col\"><div class=\"input-group\"><span class=\"input-group-text\">Amount</span> <input class=\"form-control\" id=\"document_amount\" placeholder=\"Amount\" name=\"doc-amount\" value=\"")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(fmt.Sprintf("%f", doc.Amount)))
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("\"> <span class=\"input-group-text\">€</span></div></div><div class=\"col\"><div class=\"input-group\"><span class=\"input-group-text\">Number</span> <input class=\"form-control\" id=\"document_number\" placeholder=\"Invoice Number\" name=\"doc-invoicenumber\" value=\"")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(doc.InvoiceNumber))
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("\"></div></div></div><div class=\"mb-3\"><div class=\"document_download\"><i class=\"bi bi-cloud-arrow-down\"></i> <a class=\"document_download_link\" href=\"")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		var templ_7745c5c3_Var3 templ.SafeURL = templ.URL("/mydms/file/" + doc.PreviewLink)
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(string(templ_7745c5c3_Var3)))
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("\" target=\"_NEW\">")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		var templ_7745c5c3_Var4 string
		templ_7745c5c3_Var4, templ_7745c5c3_Err = templ.JoinStringErrs(doc.FileName)
		if templ_7745c5c3_Err != nil {
			return templ.Error{Err: templ_7745c5c3_Err, FileName: `component_dialog_edit.templ`, Line: 51, Col: 22}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var4))
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</a></div></div><div class=\"mb-3\">")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		for _, t := range doc.Tags {
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<span class=\"badge text-bg-secondary document-tag\">#")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			var templ_7745c5c3_Var5 string
			templ_7745c5c3_Var5, templ_7745c5c3_Err = templ.JoinStringErrs(t)
			if templ_7745c5c3_Err != nil {
				return templ.Error{Err: templ_7745c5c3_Err, FileName: `component_dialog_edit.templ`, Line: 57, Col: 62}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var5))
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</span>")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</div><!-- https://github.com/lekoala/bootstrap5-autocomplete?tab=readme-ov-file --><div class=\"mb-3\"><div class=\"input-group\"><span class=\"input-group-text\">#Tag</span> <input class=\"form-control\" id=\"document_tag\" placeholder=\"Document Tag\" name=\"doc-tag\"> <button type=\"button\" class=\"btn btn-outline-secondary\" id=\"button_add_tag\">Add</button></div></div><div class=\"mb-3\">")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		for _, s := range doc.Senders {
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<span class=\"badge text-bg-light document-tag\"><a title=\"")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(s))
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("\"><i class=\"bi bi-truck\"></i> ")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			var templ_7745c5c3_Var6 string
			templ_7745c5c3_Var6, templ_7745c5c3_Err = templ.JoinStringErrs(templates.Ellipsis(s, 30, "~"))
			if templ_7745c5c3_Err != nil {
				return templ.Error{Err: templ_7745c5c3_Err, FileName: `component_dialog_edit.templ`, Line: 70, Col: 129}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var6))
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</a></span>")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</div><div class=\"mb-3\"><div class=\"input-group\"><span class=\"input-group-text\"><i class=\"bi bi-truck\"></i>&nbsp;Sender</span> <input class=\"form-control\" id=\"document_sender\" placeholder=\"Document Sender\" name=\"doc-sender\"> <button type=\"button\" class=\"btn btn-outline-secondary\" id=\"button_add_tag\">Add</button></div></div></div><!-- footer --><div class=\"modal-footer\"><button type=\"button\" class=\"btn btn-secondary\" data-bs-dismiss=\"modal\">Close</button> <button id=\"btn-document-save\" type=\"button\" class=\"btn btn-success\">Save</button></div></form></div></div>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		if !templ_7745c5c3_IsBuffer {
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteTo(templ_7745c5c3_W)
		}
		return templ_7745c5c3_Err
	})
}
