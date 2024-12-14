package html

import (
	_ "embed"
	"fmt"
	"strings"

	"golang.binggl.net/monorepo/internal/bookmarks/app/bookmarks"
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

type Bookmark struct {
	ID                 ValidatorInput
	Path               ValidatorInput
	DisplayName        ValidatorInput
	URL                ValidatorInput
	Type               bookmarks.NodeType
	CustomFavicon      ValidatorInput
	InvertFaviconColor bool
	UseCustomFavicon   bool
	Error              string
	Close              bool
	TStamp             string
}

type ValidatorInput struct {
	Val     string
	Valid   bool
	Message string
}

func nodeTypeInput(bm Bookmark, t bookmarks.NodeType) g.Node {
	id_bookmark_type := "type_Bookmark"
	bookmark_type := "Node"
	label := "Bookmark"
	if t == bookmarks.Folder {
		id_bookmark_type = "type_Folder"
		bookmark_type = "Folder"
		label = "Folder"
	}
	var checked g.Node
	if bm.Type == t {
		checked = h.Checked()
	}

	return h.Div(h.Class("form-check form-check-inline"),
		h.Input(h.Class(classCond("form-check-input", "disable", bm.ID.Val != "-1")), h.Type("radio"), h.Name("bookmark_Type"), h.ID(id_bookmark_type), h.Value(bookmark_type), checked),
		h.Label(h.Class(classCond("form-check-label", "disable", bm.ID.Val != "-1")), h.For(id_bookmark_type), g.Text(label)),
	)
}

func classCond(starter, conditional string, condition bool) string {
	classes := make([]string, 1)
	classes = append(classes, starter)
	if condition {
		classes = append(classes, conditional)
	}
	return strings.Join(classes, " ")
}

func selected(condition bool) g.Node {
	if condition {
		return h.Selected()
	}
	return nil
}

func checked(condition bool) g.Node {
	if condition {
		return h.Checked()
	}
	return nil
}

//go:embed editBookmark.min.js
var editLogic string

func EditBookmarks(bm Bookmark, paths []string) g.Node {
	return h.Div(h.Class("modal-dialog modal-xl"), h.ID("bookmark_edit_dialog"),
		g.If(bm.Close, g.El("script", g.Attr("type", "text/javascript"), g.Raw("bootstrap.Modal.getInstance('#modals-here').toggle();"))),
		g.If(!bm.Close,
			h.Div(h.Class("modal-content"),
				h.Div(h.Class("modal-header"),
					g.If(bm.ID.Val != "-1", h.H5(h.Class("modal-title"), g.Text(fmt.Sprintf("Edit Bookmark '%s'", bm.DisplayName.Val)))),
					g.If(bm.ID.Val == "-1", h.H5(h.Class("modal-title"), g.Text("Create Bookmark"))),
					h.Div(h.ID("indicator"), h.Class("htmx-indicator"),
						h.Div(h.Class("spinner-border text-light"), h.Role("status"),
							h.Span(h.Class("visually-hidden"), g.Text("Loading...")),
						),
					),
				),
				h.Form(h.Class("bookmark_edit_form"),
					h.Input(h.Type("hidden"), h.Name("bookmark_ID"), h.Value(bm.ID.Val)),
					h.Div(h.Class("modal-body"),
						nodeTypeInput(bm, bookmarks.Node),
						nodeTypeInput(bm, bookmarks.Folder),
						h.Div(h.Class("spacer")),
						h.Div(h.Class("flex_layout"),
							h.Div(h.Class("bookmark_edit_layout_flex_5"),
								h.Img(h.ID("bookmark_favicon_display"), h.Class(classCond("bookmark_favicon_preview", "invert", bm.InvertFaviconColor)), h.Src("/bm/favicon/"+bm.ID.Val+"?t="+bm.TStamp))),
							h.Div(h.Class("bookmark_edit_layout_flex_95 form-floating mb-3"),
								h.Input(h.Type("text"), h.Class(classCond("form-control", "control_invalid", !bm.DisplayName.Valid)), h.ID("bookmark_DisplayName"), h.Placeholder("Displayname"), h.Name("bookmark_DisplayName"), h.Value(bm.DisplayName.Val), h.Required()),
								h.Label(h.For("bookmark_DisplayName"), g.Text("Displayname")),
							),
						),
						g.If(bm.Type == bookmarks.Node,
							h.Div(h.Class("input-group mb-3"), h.ID("url_section"),
								h.Span(h.Class("input-group-text"), h.ID("url"), h.I(h.Class("bi bi-link-45deg"))),
								h.Input(h.Type("text"), h.ID("bookmark_URL"), h.Class(classCond("form-control", "control_invalid", !bm.URL.Valid)), h.Placeholder("URL"), h.Name("bookmark_URL"), h.Value(bm.URL.Val)),
								h.Button(
									h.Type("button"),
									h.Class("btn btn-outline-secondary"),
									g.Attr("hx-post", "/bm/favicon/page"),
									g.Attr("hx-trigger", "click"),
									g.Attr("hx-target", "#bookmark_favicon_display"),
									g.Attr("hx-params", "bookmark_URL"),
									g.Attr("hx-swap", "outerHTML"),
									g.Attr("hx-indicator", "#indicator"),
									h.I(h.Class("bi bi-arrow-clockwise")),
								),
							),
						),
						h.Div(h.Class("form-floating mb-3"),
							h.Select(h.Class("form-select"), h.ID("bookmark_Path"), h.Name("bookmark_Path"), h.Required(),
								g.If(bm.ID.Val != "-1",
									g.Map(paths, func(p string) g.Node {
										return h.Option(h.Value(p), selected(bm.Path.Val == p), g.Text(p))
									}),
								),
								g.If(bm.ID.Val == "-1",
									h.Option(h.Value(bm.Path.Val), h.Selected(), g.Text(bm.Path.Val)),
								),
							),
							h.Label(h.For("bookmark_Path"), g.Text("Path")),
						),
						h.Div(h.Class("form-check form-switch"),
							h.Input(h.Class("form-check-input"), h.Type("checkbox"), h.Role("switch"), h.ID("bookmark_Invert"), h.Name("bookmark_InvertFaviconColor"), h.Value("1"), checked(bm.InvertFaviconColor)),
							h.Label(h.Class("form-check-label"), h.For("bookmark_Invert"), g.Text("Invert Favicon Color")),
						),
						h.Div(h.Class("form-check form-switch"),
							h.Input(h.Class("form-check-input"), h.Type("checkbox"), h.Role("switch"), h.ID("bookmark_Custom_Favicon"), h.Name("bookmark_UseCustomFavicon"), h.Value("1"), checked(bm.UseCustomFavicon)),
							h.Label(h.Class("form-check-label"), h.For("bookmark_Custom_Favicon"), g.Text("Custom Favicon")),
						),
						h.Div(h.ID("custom_favicon_section"), h.Class(classCond("", "d-none", !bm.UseCustomFavicon)),
							h.Div(h.Class("spacer")),
							h.Div(h.Class("input-group mb-3"),
								h.Input(
									h.Type("text"),
									h.Class(classCond("form-control", "control_invalid", !bm.CustomFavicon.Valid)),
									h.Placeholder("Favicon URL"),
									h.Name("bookmark_CustomFavicon"),
									h.Value(bm.CustomFavicon.Val),
								),
								h.Button(
									h.Type("button"),
									h.ID("btnCustomFaviconURL"),
									h.Class("btn btn-outline-secondary"),
									g.Attr("hx-post", "/bm/favicon/url"),
									g.Attr("hx-trigger", "click"),
									g.Attr("hx-target", "#bookmark_favicon_display"),
									g.Attr("hx-params", "bookmark_CustomFavicon"),
									g.Attr("hx-swap", "outerHTML"),
									g.Attr("hx-indicator", "#indicator"),
									h.I(h.Class("bi bi-arrow-clockwise")),
								),
							),
							h.Label(h.Class("form-label"), h.For("customFaviconUpload"), g.Text("Upload a custom icon")),
							h.Div(h.Class("input-group mb-3"),
								h.Input(
									h.Type("file"),
									h.Class("form-control"),
									h.ID("customFaviconUpload"),
									h.Name("bookmark_customFaviconUpload"),
									h.Accept("image/*,.png,.jpeg,.jpg,.gif,.svg"),
								),
								h.Button(
									h.Type("button"),
									h.ID("btnUploadCustomFavicon"),
									h.Class("btn btn-outline-secondary"),
									g.Attr("hx-post", "/bm/favicon/upload"),
									g.Attr("hx-encoding", "multipart/form-data"),
									g.Attr("hx-trigger", "click"),
									g.Attr("hx-target", "#bookmark_favicon_display"),
									g.Attr("hx-params", "bookmark_customFaviconUpload"),
									g.Attr("hx-swap", "outerHTML"),
									g.Attr("hx-indicator", "#indicator"),
									h.I(h.Class("bi bi-upload")),
								),
							),
						),
						h.Div(h.ID("error_section"), h.Class(classCond("", "d-none", bm.Error == "")),
							h.Div(h.Class("spacer")),
							h.I(h.Class("bi bi-exclamation-diamond")),
							g.Raw("&nbsp;"),
							h.Span(g.Text(bm.Error)),
						),
						h.Div(h.ID("info_section"), h.Class("info_text d-none"),
							h.Div(h.Class("spacer")),
							h.I(h.Class("bi bi-info-circle")),
							g.Raw("&nbsp;"),
							h.Span(h.ID("info_section_text"), g.Text("TEXT")),
						),
					),
					h.Div(h.Class("modal-footer"),
						h.Button(h.Type("button"), h.Class("btn btn-secondary"), g.Attr("data-bs-dismiss", "modal"), g.Text("Close")),
						h.Button(
							h.ID("btn-bookmark-save"),
							h.Type("button"),
							h.Class("btn btn-success"),
							g.Attr("hx-post", "/bm"),
							g.Attr("hx-target", "#bookmark_edit_dialog"),
							g.Text("Save"),
						),
					),
				),
			),
		),
		g.El("script", g.Attr("type", "text/javascript"),
			g.Raw(editLogic),
		),
	)
}
