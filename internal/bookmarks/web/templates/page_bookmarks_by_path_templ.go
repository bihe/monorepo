// Code generated by templ - DO NOT EDIT.

// templ: version: v0.2.707
package templates

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import "github.com/a-h/templ"
import "context"
import "io"
import "bytes"

func BookmarksByPathContent(bookmarkList templ.Component) templ.Component {
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
		templ_7745c5c3_Err = bookmarkList.Render(ctx, templ_7745c5c3_Buffer)
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		if !templ_7745c5c3_IsBuffer {
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteTo(templ_7745c5c3_W)
		}
		return templ_7745c5c3_Err
	})
}

func BookmarksByPathStyles() templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, templ_7745c5c3_W io.Writer) (templ_7745c5c3_Err error) {
		templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := templ_7745c5c3_W.(*bytes.Buffer)
		if !templ_7745c5c3_IsBuffer {
			templ_7745c5c3_Buffer = templ.GetBuffer()
			defer templ.ReleaseBuffer(templ_7745c5c3_Buffer)
		}
		ctx = templ.InitializeContext(ctx)
		templ_7745c5c3_Var2 := templ.GetChildren(ctx)
		if templ_7745c5c3_Var2 == nil {
			templ_7745c5c3_Var2 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<style type=\"text/css\">\n    .breadcrumb-item{--bs-breadcrumb-divider-color:#ffffff !important;--bs-breadcrumb-divider:'>';font-size:medium}.breadcrumb-item.active{color:#ffffff}li.breadcrumb-item > a{color:#ffffff}div.btn-group > button.btn.dropdown-toggle{--bs-btn-color:#ffffff}.delete{font-weight:bold;color:red}.right-action{position:absolute;right:20px}.sortInput{position:relative;top:18px}@media only screen and (min-device-width: 375px) and (max-device-width: 812px){.breadcrumb-item{--bs-breadcrumb-divider-color:#ffffff !important;--bs-breadcrumb-divider:'>';font-size:smaller}.breadcrumb-item.active{color:#ffffff}li.breadcrumb-item > a{color:#ffffff}}\n    </style>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		if !templ_7745c5c3_IsBuffer {
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteTo(templ_7745c5c3_W)
		}
		return templ_7745c5c3_Err
	})
}

type BookmarkPathEntry struct {
	UrlPath     string
	DisplayName string
	LastItem    bool
}

func getPath(entries []BookmarkPathEntry) string {
	return entries[len(entries)-1].UrlPath
}

func BookmarksByPathNavigation(entries []BookmarkPathEntry) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, templ_7745c5c3_W io.Writer) (templ_7745c5c3_Err error) {
		templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := templ_7745c5c3_W.(*bytes.Buffer)
		if !templ_7745c5c3_IsBuffer {
			templ_7745c5c3_Buffer = templ.GetBuffer()
			defer templ.ReleaseBuffer(templ_7745c5c3_Buffer)
		}
		ctx = templ.InitializeContext(ctx)
		templ_7745c5c3_Var3 := templ.GetChildren(ctx)
		if templ_7745c5c3_Var3 == nil {
			templ_7745c5c3_Var3 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<div class=\"application_name\"><nav aria-label=\"breadcrumb\"><ol class=\"breadcrumb\">")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		for i, e := range entries {
			if e.LastItem {
				if i == 0 {
					_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<li class=\"breadcrumb-item active\" aria-current=\"page\"><i class=\"bi bi-house\"></i></li>")
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
				} else {
					_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<li class=\"breadcrumb-item active\" aria-current=\"page\">")
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
					var templ_7745c5c3_Var4 string
					templ_7745c5c3_Var4, templ_7745c5c3_Err = templ.JoinStringErrs(e.DisplayName)
					if templ_7745c5c3_Err != nil {
						return templ.Error{Err: templ_7745c5c3_Err, FileName: `page_bookmarks_by_path.templ`, Line: 32, Col: 77}
					}
					_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var4))
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
					_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</li>")
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
				}
			} else {
				if i == 0 {
					_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<li class=\"breadcrumb-item\"><a class=\"rootroot\" href=\"")
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
					var templ_7745c5c3_Var5 templ.SafeURL = templ.URL("/bm/~" + e.UrlPath)
					_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(string(templ_7745c5c3_Var5)))
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
					_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("\"><i class=\"bi bi-house\"></i></a></li>")
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
				} else {
					_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<li class=\"breadcrumb-item\"><a href=\"")
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
					var templ_7745c5c3_Var6 templ.SafeURL = templ.URL("/bm/~" + e.UrlPath)
					_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(string(templ_7745c5c3_Var6)))
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
					_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("\">")
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
					var templ_7745c5c3_Var7 string
					templ_7745c5c3_Var7, templ_7745c5c3_Err = templ.JoinStringErrs(e.DisplayName)
					if templ_7745c5c3_Err != nil {
						return templ.Error{Err: templ_7745c5c3_Err, FileName: `page_bookmarks_by_path.templ`, Line: 38, Col: 93}
					}
					_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var7))
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
					_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</a></li>")
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
				}
			}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</ol></nav><span class=\"right-action\"><div id=\"request_indicator\" class=\"request_indicator htmx-indicator\"><div class=\"spinner-border text-light\" role=\"status\"><span class=\"visually-hidden\">Loading...</span></div></div><button id=\"btn_toggle_sorting\" type=\"button\" data-bs-toggle=\"button\" class=\"btn sort_button\"><i class=\"bi bi-arrow-down-up\"></i> Sort</button> <span id=\"save_list_sort_order\" class=\"sort_button d-none\"><button id=\"btn_save_sorting\" type=\"button\" class=\"btn btn-success sort_button\"><i class=\"bi bi-sort-numeric-down\"></i> Save</button></span> <button type=\"button\" data-testid=\"link-add-bookmark\" class=\"btn btn-primary new_button\" data-bs-toggle=\"modal\" data-bs-target=\"#modals-here\" hx-target=\"#modals-here\" hx-trigger=\"click\" hx-get=\"")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		var templ_7745c5c3_Var8 string
		templ_7745c5c3_Var8, templ_7745c5c3_Err = templ.JoinStringErrs("/bm/-1?path=" + getPath(entries))
		if templ_7745c5c3_Err != nil {
			return templ.Error{Err: templ_7745c5c3_Err, FileName: `page_bookmarks_by_path.templ`, Line: 71, Col: 46}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var8))
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("\"><i class=\"bi bi-plus\"></i> Add</button></span>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = sortingLogic().Render(ctx, templ_7745c5c3_Buffer)
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
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

func sortingLogic() templ.ComponentScript {
	return templ.ComponentScript{
		Name: `__templ_sortingLogic_d070`,
		Function: `function __templ_sortingLogic_d070(){try {
		document.querySelector('#btn_toggle_sorting').addEventListener('click', (event) => {
			if (event.target.classList.contains('active')) {
				console.log('Activate sorting');
				sortableRefresh(document);
			} else {
				console.log('Disable sorting - refresh the list');
				document.querySelector('#save_list_sort_order').classList.add('d-none');
				htmx.trigger('#btn_toggle_sorting', 'refreshBookmarkList');
			}
		});
		document.querySelector('#btn_save_sorting').addEventListener('click', (event) => {
			htmx.trigger('#btn_save_sorting', 'sortBookmarkList');
			document.querySelector('#btn_toggle_sorting').classList.remove('active');
			document.querySelector('#save_list_sort_order').classList.add('d-none');
		});
	} catch(error) {
		console.error(error);
	}
}`,
		Call:       templ.SafeScript(`__templ_sortingLogic_d070`),
		CallInline: templ.SafeScriptInline(`__templ_sortingLogic_d070`),
	}
}
