// Code generated by templ - DO NOT EDIT.

// templ: version: v0.2.707
package templates

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import "github.com/a-h/templ"
import "context"
import "io"
import "bytes"

import "golang.binggl.net/monorepo/internal/bookmarks/app/bookmarks"

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

func EditBookmarks(bm Bookmark, paths []string) templ.Component {
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
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<div class=\"modal-dialog modal-xl\" id=\"bookmark_edit_dialog\">")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		if bm.Close {
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<script type=\"text/javascript\">\n\t\t\t\tbootstrap.Modal.getInstance('#modals-here').toggle();\n\t\t\t</script>")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
		} else {
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<div class=\"modal-content\"><div class=\"modal-header\">")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			if bm.ID.Val != "-1" {
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<h5 class=\"modal-title\">Edit Bookmark '")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				var templ_7745c5c3_Var2 string
				templ_7745c5c3_Var2, templ_7745c5c3_Err = templ.JoinStringErrs(bm.DisplayName.Val)
				if templ_7745c5c3_Err != nil {
					return templ.Error{Err: templ_7745c5c3_Err, FileName: `component_dialog_edit_bookmark.templ`, Line: 35, Col: 65}
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var2))
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("'</h5>")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
			} else {
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<h5 class=\"modal-title\">Create Bookmark</h5>")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<div id=\"indicator\" class=\"htmx-indicator\"><div class=\"spinner-border text-light\" role=\"status\"><span class=\"visually-hidden\">Loading...</span></div></div></div><form class=\"bookmark_edit_form\"><input type=\"hidden\" name=\"bookmark_ID\" value=\"")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			var templ_7745c5c3_Var3 string
			templ_7745c5c3_Var3, templ_7745c5c3_Err = templ.JoinStringErrs(bm.ID.Val)
			if templ_7745c5c3_Err != nil {
				return templ.Error{Err: templ_7745c5c3_Err, FileName: `component_dialog_edit_bookmark.templ`, Line: 46, Col: 62}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var3))
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("\"><div class=\"modal-body\"><div class=\"form-check form-check-inline\">")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			var templ_7745c5c3_Var4 = []any{"form-check-input", templ.KV("disable", (bm.ID.Val != "-1"))}
			templ_7745c5c3_Err = templ.RenderCSSItems(ctx, templ_7745c5c3_Buffer, templ_7745c5c3_Var4...)
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<input class=\"")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			var templ_7745c5c3_Var5 string
			templ_7745c5c3_Var5, templ_7745c5c3_Err = templ.JoinStringErrs(templ.CSSClasses(templ_7745c5c3_Var4).String())
			if templ_7745c5c3_Err != nil {
				return templ.Error{Err: templ_7745c5c3_Err, FileName: `component_dialog_edit_bookmark.templ`, Line: 1, Col: 0}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var5))
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("\" type=\"radio\" name=\"bookmark_Type\" id=\"type_Bookmark\" value=\"Node\"")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			if bm.Type == bookmarks.Node {
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(" checked")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("> ")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			var templ_7745c5c3_Var6 = []any{"form-check-label", templ.KV("disable", (bm.ID.Val != "-1"))}
			templ_7745c5c3_Err = templ.RenderCSSItems(ctx, templ_7745c5c3_Buffer, templ_7745c5c3_Var6...)
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<label class=\"")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			var templ_7745c5c3_Var7 string
			templ_7745c5c3_Var7, templ_7745c5c3_Err = templ.JoinStringErrs(templ.CSSClasses(templ_7745c5c3_Var6).String())
			if templ_7745c5c3_Err != nil {
				return templ.Error{Err: templ_7745c5c3_Err, FileName: `component_dialog_edit_bookmark.templ`, Line: 1, Col: 0}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var7))
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("\" for=\"type_Bookmark\">Bookmark</label></div><div class=\"form-check form-check-inline\">")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			var templ_7745c5c3_Var8 = []any{"form-check-input", templ.KV("disable", (bm.ID.Val != "-1"))}
			templ_7745c5c3_Err = templ.RenderCSSItems(ctx, templ_7745c5c3_Buffer, templ_7745c5c3_Var8...)
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<input class=\"")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			var templ_7745c5c3_Var9 string
			templ_7745c5c3_Var9, templ_7745c5c3_Err = templ.JoinStringErrs(templ.CSSClasses(templ_7745c5c3_Var8).String())
			if templ_7745c5c3_Err != nil {
				return templ.Error{Err: templ_7745c5c3_Err, FileName: `component_dialog_edit_bookmark.templ`, Line: 1, Col: 0}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var9))
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("\" type=\"radio\" name=\"bookmark_Type\" id=\"type_Folder\" value=\"Folder\"")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			if bm.Type == bookmarks.Folder {
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(" checked")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("> ")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			var templ_7745c5c3_Var10 = []any{"form-check-label", templ.KV("disable", (bm.ID.Val != "-1"))}
			templ_7745c5c3_Err = templ.RenderCSSItems(ctx, templ_7745c5c3_Buffer, templ_7745c5c3_Var10...)
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<label class=\"")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			var templ_7745c5c3_Var11 string
			templ_7745c5c3_Var11, templ_7745c5c3_Err = templ.JoinStringErrs(templ.CSSClasses(templ_7745c5c3_Var10).String())
			if templ_7745c5c3_Err != nil {
				return templ.Error{Err: templ_7745c5c3_Err, FileName: `component_dialog_edit_bookmark.templ`, Line: 1, Col: 0}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var11))
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("\" for=\"type_Folder\">Folder</label></div><div class=\"spacer\"></div><div class=\"flex_layout\"><div class=\"bookmark_edit_layout_flex_5\">")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			var templ_7745c5c3_Var12 = []any{"bookmark_favicon_preview", templ.KV("invert", bm.InvertFaviconColor)}
			templ_7745c5c3_Err = templ.RenderCSSItems(ctx, templ_7745c5c3_Buffer, templ_7745c5c3_Var12...)
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<img id=\"bookmark_favicon_display\" class=\"")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			var templ_7745c5c3_Var13 string
			templ_7745c5c3_Var13, templ_7745c5c3_Err = templ.JoinStringErrs(templ.CSSClasses(templ_7745c5c3_Var12).String())
			if templ_7745c5c3_Err != nil {
				return templ.Error{Err: templ_7745c5c3_Err, FileName: `component_dialog_edit_bookmark.templ`, Line: 1, Col: 0}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var13))
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("\" src=\"")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			var templ_7745c5c3_Var14 string
			templ_7745c5c3_Var14, templ_7745c5c3_Err = templ.JoinStringErrs("/bm/favicon/" + bm.ID.Val + "?t=" + bm.TStamp)
			if templ_7745c5c3_Err != nil {
				return templ.Error{Err: templ_7745c5c3_Err, FileName: `component_dialog_edit_bookmark.templ`, Line: 59, Col: 175}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var14))
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("\"></div><div class=\"bookmark_edit_layout_flex_95 form-floating mb-3\">")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			var templ_7745c5c3_Var15 = []any{"form-control", templ.KV("control_invalid", !bm.DisplayName.Valid)}
			templ_7745c5c3_Err = templ.RenderCSSItems(ctx, templ_7745c5c3_Buffer, templ_7745c5c3_Var15...)
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<input type=\"text\" class=\"")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			var templ_7745c5c3_Var16 string
			templ_7745c5c3_Var16, templ_7745c5c3_Err = templ.JoinStringErrs(templ.CSSClasses(templ_7745c5c3_Var15).String())
			if templ_7745c5c3_Err != nil {
				return templ.Error{Err: templ_7745c5c3_Err, FileName: `component_dialog_edit_bookmark.templ`, Line: 1, Col: 0}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var16))
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("\" class=\"\" id=\"bookmark_DisplayName\" placeholder=\"Displayname\" name=\"bookmark_DisplayName\" value=\"")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			var templ_7745c5c3_Var17 string
			templ_7745c5c3_Var17, templ_7745c5c3_Err = templ.JoinStringErrs(bm.DisplayName.Val)
			if templ_7745c5c3_Err != nil {
				return templ.Error{Err: templ_7745c5c3_Err, FileName: `component_dialog_edit_bookmark.templ`, Line: 62, Col: 219}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var17))
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("\" required> <label for=\"bookmark_DisplayName\">DisplayName</label></div></div>")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			if bm.Type == bookmarks.Node {
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<div class=\"input-group mb-3\" id=\"url_section\"><span class=\"input-group-text\" id=\"url\"><i class=\"bi bi-link-45deg\"></i></span> ")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				var templ_7745c5c3_Var18 = []any{"form-control", templ.KV("control_invalid", !bm.URL.Valid)}
				templ_7745c5c3_Err = templ.RenderCSSItems(ctx, templ_7745c5c3_Buffer, templ_7745c5c3_Var18...)
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<input type=\"text\" id=\"bookmark_URL\" class=\"")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				var templ_7745c5c3_Var19 string
				templ_7745c5c3_Var19, templ_7745c5c3_Err = templ.JoinStringErrs(templ.CSSClasses(templ_7745c5c3_Var18).String())
				if templ_7745c5c3_Err != nil {
					return templ.Error{Err: templ_7745c5c3_Err, FileName: `component_dialog_edit_bookmark.templ`, Line: 1, Col: 0}
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var19))
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("\" placeholder=\"URL\" aria-label=\"URL\" aria-describedby=\"url\" name=\"bookmark_URL\" value=\"")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				var templ_7745c5c3_Var20 string
				templ_7745c5c3_Var20, templ_7745c5c3_Err = templ.JoinStringErrs(bm.URL.Val)
				if templ_7745c5c3_Err != nil {
					return templ.Error{Err: templ_7745c5c3_Err, FileName: `component_dialog_edit_bookmark.templ`, Line: 69, Col: 210}
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var20))
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("\"> <button type=\"button\" class=\"btn btn-outline-secondary\" hx-post=\"/bm/favicon/page\" hx-trigger=\"click\" hx-target=\"#bookmark_favicon_display\" hx-params=\"bookmark_URL\" hx-swap=\"outerHTML\" hx-indicator=\"#indicator\"><i class=\"bi bi-arrow-clockwise\"></i></button></div>")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<div class=\"form-floating mb-3\"><select class=\"form-select\" aria-label=\"Path\" id=\"bookmark_Path\" name=\"bookmark_Path\" required>")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			if bm.ID.Val != "-1" {
				for _, p := range paths {
					_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<option value=\"")
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
					var templ_7745c5c3_Var21 string
					templ_7745c5c3_Var21, templ_7745c5c3_Err = templ.JoinStringErrs(p)
					if templ_7745c5c3_Err != nil {
						return templ.Error{Err: templ_7745c5c3_Err, FileName: `component_dialog_edit_bookmark.templ`, Line: 86, Col: 27}
					}
					_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var21))
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
					_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("\"")
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
					if bm.Path.Val == p {
						_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(" selected")
						if templ_7745c5c3_Err != nil {
							return templ_7745c5c3_Err
						}
					}
					_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(">")
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
					var templ_7745c5c3_Var22 string
					templ_7745c5c3_Var22, templ_7745c5c3_Err = templ.JoinStringErrs(p)
					if templ_7745c5c3_Err != nil {
						return templ.Error{Err: templ_7745c5c3_Err, FileName: `component_dialog_edit_bookmark.templ`, Line: 86, Col: 64}
					}
					_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var22))
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
					_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</option>")
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
				}
			} else {
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<option value=\"")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				var templ_7745c5c3_Var23 string
				templ_7745c5c3_Var23, templ_7745c5c3_Err = templ.JoinStringErrs(bm.Path.Val)
				if templ_7745c5c3_Err != nil {
					return templ.Error{Err: templ_7745c5c3_Err, FileName: `component_dialog_edit_bookmark.templ`, Line: 89, Col: 36}
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var23))
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("\" selected>")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				var templ_7745c5c3_Var24 string
				templ_7745c5c3_Var24, templ_7745c5c3_Err = templ.JoinStringErrs(bm.Path.Val)
				if templ_7745c5c3_Err != nil {
					return templ.Error{Err: templ_7745c5c3_Err, FileName: `component_dialog_edit_bookmark.templ`, Line: 89, Col: 61}
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var24))
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</option>")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</select> <label for=\"bookmark_Path\">Path</label></div><div class=\"form-check form-switch\"><input class=\"form-check-input\" type=\"checkbox\" role=\"switch\" id=\"bookmark_Invert\" name=\"bookmark_InvertFaviconColor\" value=\"1\"")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			if bm.InvertFaviconColor {
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(" checked")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("> <label class=\"form-check-label\" for=\"bookmark_Invert\">Invert Favicon Color</label></div><div class=\"form-check form-switch\"><input class=\"form-check-input\" type=\"checkbox\" role=\"switch\" id=\"bookmark_Custom_Favicon\" name=\"bookmark_UseCustomFavicon\" value=\"1\"")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			if bm.UseCustomFavicon {
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(" checked")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("> <label class=\"form-check-label\" for=\"bookmark_Custom_Favicon\">Custom Favicon</label></div>")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			var templ_7745c5c3_Var25 = []any{templ.KV("d-none", !bm.UseCustomFavicon)}
			templ_7745c5c3_Err = templ.RenderCSSItems(ctx, templ_7745c5c3_Buffer, templ_7745c5c3_Var25...)
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<div id=\"custom_favicon_section\" class=\"")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			var templ_7745c5c3_Var26 string
			templ_7745c5c3_Var26, templ_7745c5c3_Err = templ.JoinStringErrs(templ.CSSClasses(templ_7745c5c3_Var25).String())
			if templ_7745c5c3_Err != nil {
				return templ.Error{Err: templ_7745c5c3_Err, FileName: `component_dialog_edit_bookmark.templ`, Line: 1, Col: 0}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var26))
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("\"><div class=\"spacer\"></div><div class=\"input-group mb-3\">")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			var templ_7745c5c3_Var27 = []any{"form-control", templ.KV("control_invalid", !bm.CustomFavicon.Valid)}
			templ_7745c5c3_Err = templ.RenderCSSItems(ctx, templ_7745c5c3_Buffer, templ_7745c5c3_Var27...)
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<input type=\"text\" class=\"")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			var templ_7745c5c3_Var28 string
			templ_7745c5c3_Var28, templ_7745c5c3_Err = templ.JoinStringErrs(templ.CSSClasses(templ_7745c5c3_Var27).String())
			if templ_7745c5c3_Err != nil {
				return templ.Error{Err: templ_7745c5c3_Err, FileName: `component_dialog_edit_bookmark.templ`, Line: 1, Col: 0}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var28))
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("\" placeholder=\"Favicon URL\" aria-label=\"Favicon URL\" aria-describedby=\"Favicon URL\" name=\"bookmark_CustomFavicon\" value=\"")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			var templ_7745c5c3_Var29 string
			templ_7745c5c3_Var29, templ_7745c5c3_Err = templ.JoinStringErrs(bm.CustomFavicon.Val)
			if templ_7745c5c3_Err != nil {
				return templ.Error{Err: templ_7745c5c3_Err, FileName: `component_dialog_edit_bookmark.templ`, Line: 112, Col: 37}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var29))
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("\"> <button type=\"button\" class=\"btn btn-outline-secondary\" hx-post=\"/bm/favicon/url\" hx-trigger=\"click\" hx-target=\"#bookmark_favicon_display\" hx-params=\"bookmark_CustomFavicon\" hx-swap=\"outerHTML\" hx-indicator=\"#indicator\"><i class=\"bi bi-arrow-clockwise\"></i></button></div><label for=\"customFaviconUpload\" class=\"form-label\">Upload a custom icon</label><div class=\"input-group mb-3\"><input class=\"form-control\" type=\"file\" name=\"bookmark_customFaviconUpload\" id=\"customFaviconUpload\" accept=\"image/*,.png,.jpeg,.jpg,.gif,.svg\"> <button type=\"button\" id=\"btnUploadCustomFavicon\" class=\"btn btn-outline-secondary\" hx-post=\"/bm/favicon/upload\" hx-encoding=\"multipart/form-data\" hx-trigger=\"click\" hx-target=\"#bookmark_favicon_display\" hx-params=\"bookmark_customFaviconUpload\" hx-swap=\"outerHTML\" hx-indicator=\"#indicator\"><i class=\"bi bi-upload\"></i></button></div></div>")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			var templ_7745c5c3_Var30 = []any{"error_text", templ.KV("d-none", bm.Error == "")}
			templ_7745c5c3_Err = templ.RenderCSSItems(ctx, templ_7745c5c3_Buffer, templ_7745c5c3_Var30...)
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<div id=\"error_section\" class=\"")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			var templ_7745c5c3_Var31 string
			templ_7745c5c3_Var31, templ_7745c5c3_Err = templ.JoinStringErrs(templ.CSSClasses(templ_7745c5c3_Var30).String())
			if templ_7745c5c3_Err != nil {
				return templ.Error{Err: templ_7745c5c3_Err, FileName: `component_dialog_edit_bookmark.templ`, Line: 1, Col: 0}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var31))
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("\"><div class=\"spacer\"></div><i class=\"bi bi-exclamation-diamond\"></i>&nbsp;<span>")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			var templ_7745c5c3_Var32 string
			templ_7745c5c3_Var32, templ_7745c5c3_Err = templ.JoinStringErrs(bm.Error)
			if templ_7745c5c3_Err != nil {
				return templ.Error{Err: templ_7745c5c3_Err, FileName: `component_dialog_edit_bookmark.templ`, Line: 150, Col: 70}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var32))
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</span></div><div id=\"info_section\" class=\"info_text d-none\"><div class=\"spacer\"></div><i class=\"bi bi-info-circle\"></i>&nbsp;<span id=\"info_section_text\">TEXT</span></div></div><div class=\"modal-footer\"><button type=\"button\" class=\"btn btn-secondary\" data-bs-dismiss=\"modal\">Close</button> <button id=\"btn-bookmark-save\" type=\"button\" class=\"btn btn-success\" hx-post=\"/bm\" hx-target=\"#bookmark_edit_dialog\">Save</button></div></form></div><script type=\"text/javascript\">\n\t\t\t\tif (document.querySelector('#type_Bookmark')) {\n\t\t\t\t\tdocument.querySelector('#type_Bookmark').addEventListener('change', (event) => {\n\t\t\t\t\t\tif (event.target.value === 'Node') {\n\t\t\t\t\t\tdocument.querySelector('#url_section').classList.remove('d-none');\n\t\t\t\t\t\t} else {\n\t\t\t\t\t\tdocument.querySelector('#url_section').classList.add('d-none');\n\t\t\t\t\t\t}\n\t\t\t\t\t});\n\t\t\t\t}\n\n\t\t\t\tif (document.querySelector('#type_Folder')) {\n\t\t\t\t\tdocument.querySelector('#type_Folder').addEventListener('change', (event) => {\n\t\t\t\t\t\tif (event.target.value === 'Folder') {\n\t\t\t\t\t\tdocument.querySelector('#url_section').classList.add('d-none');\n\t\t\t\t\t\t} else {\n\t\t\t\t\t\tdocument.querySelector('#url_section').classList.remove('d-none');\n\t\t\t\t\t\t}\n\t\t\t\t\t});\n\t\t\t\t}\n\n\t\t\t\tif (document.querySelector('#bookmark_Custom_Favicon')) {\n\t\t\t\t\tdocument.querySelector('#bookmark_Custom_Favicon').addEventListener('change', (event) => {\n\t\t\t\t\t\tif (event.currentTarget.checked) {\n\t\t\t\t\t\tdocument.querySelector('#custom_favicon_section').classList.remove('d-none');\n\t\t\t\t\t\t} else {\n\t\t\t\t\t\tdocument.querySelector('#custom_favicon_section').classList.add('d-none');\n\t\t\t\t\t\t}\n\t\t\t\t\t});\n\t\t\t\t}\n\n\t\t\t\tif (document.querySelector('#bookmark_Invert')) {\n\t\t\t\t\tdocument.querySelector('#bookmark_Invert').addEventListener('change', (event) => {\n\t\t\t\t\t\tif (event.currentTarget.checked) {\n\t\t\t\t\t\tdocument.querySelector('#bookmark_favicon_display').classList.add('invert');\n\t\t\t\t\t\t} else {\n\t\t\t\t\t\tdocument.querySelector('#bookmark_favicon_display').classList.remove('invert');\n\t\t\t\t\t\t}\n\t\t\t\t\t});\n\t\t\t\t}\n\n\t\t\t\tif (document.querySelector('.bookmark_edit_form')) {\n\t\t\t\t\tdocument.querySelector('.bookmark_edit_form').addEventListener('paste', e => {\n\t\t\t\t\t\tif (!e.clipboardData.items || e.clipboardData.items.length == 0) {\n\t\t\t\t\t\t\tshowInfoText(`Nothing to paste from clipboard!`);\n\t\t\t\t\t\t\treturn;\n\t\t\t\t\t\t}\n\t\t\t\t\t\ttry {\n\t\t\t\t\t\t\t// get the first item of the clipboard\n\t\t\t\t\t\t\tvar item = e.clipboardData.items[0];\n\t\t\t\t\t\t\tif (item.type.indexOf(\"image\") === 0 || item.type.indexOf(\"svg\") === 0) {\n\t\t\t\t\t\t\t\tlet fileInput = document.querySelector('#customFaviconUpload');\n\t\t\t\t\t\t\t\tlet dataTransfer = new DataTransfer();\n\t\t\t\t\t\t\t\tlet blob = item.getAsFile();\n\t\t\t\t\t\t\t\tlet uuid = window.crypto.randomUUID()\n\t\t\t\t\t\t\t\tdataTransfer.items.add(blob);\n\t\t\t\t\t\t\t\tfileInput.files = dataTransfer.files;\n\n\t\t\t\t\t\t\t\tconsole.log('files for upload: ' + fileInput.files.length);\n\t\t\t\t\t\t\t\tshowInfoText(`Pasted file '${blob.name}' from clipboard!`);\n\t\t\t\t\t\t\t} else {\n\t\t\t\t\t\t\t\tshowInfoText(`No image in clipboard!`);\n\t\t\t\t\t\t\t}\n\t\t\t\t\t\t} catch (e) {\n\t\t\t\t\t\t\tconsole.log(\"could not set clipboard image!\");\n\t\t\t\t\t\t\tconsole.log(e);\n\t\t\t\t\t\t}\n\n\t\t\t\t\t})\n\t\t\t\t}\n\n\t\t\t\tfunction showInfoText(text) {\n\t\t\t\t\tif (document.querySelector('#info_section')) {\n\t\t\t\t\t\tif (document.querySelector('#info_section_text')) {\n\t\t\t\t\t\t\tdocument.querySelector('#info_section_text').textContent = text;\n\t\t\t\t\t\t\tdocument.querySelector('#info_section').classList.remove('d-none');\n\n\t\t\t\t\t\t\tsetTimeout(() => {\n\t\t\t\t\t\t\t\tdocument.querySelector('#info_section_text').textContent = '';\n\t\t\t\t\t\t\t\tdocument.querySelector('#info_section').classList.add('d-none')\n\t\t\t\t\t\t\t}, 2000);\n\t\t\t\t\t\t}\n\t\t\t\t\t}\n\t\t\t\t}\n\t\t\t</script>")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</div>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		if !templ_7745c5c3_IsBuffer {
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteTo(templ_7745c5c3_W)
		}
		return templ_7745c5c3_Err
	})
}
