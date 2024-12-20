package html

import (
	_ "embed"
	"net/http"

	"golang.binggl.net/monorepo/pkg/config"
	g "maragu.dev/gomponents"
	c "maragu.dev/gomponents/components"
	h "maragu.dev/gomponents/html"
)

// ErrorPage404 creates a HTML page used to indicate that a resource cannot be found
func ErrorPage404(basePath string) g.Node {
	body := h.Div(
		h.Class("container"),
		h.Div(
			h.Class("item"),
			g.Raw(`<svg xmlns="http://www.w3.org/2000/svg" width="64" height="64" fill="currentColor" class="bi bi-search" viewBox="0 0 16 16">
				<path d="M11.742 10.344a6.5 6.5 0 1 0-1.397 1.398h-.001c.03.04.062.078.098.115l3.85 3.85a1 1 0 0 0 1.415-1.414l-3.85-3.85a1.007 1.007 0 0 0-.115-.1zM12 6.5a5.5 5.5 0 1 1-11 0 5.5 5.5 0 0 1 11 0z"></path>
			</svg>`),
			h.H2(g.Text("Content is not available")),
			h.P(g.Text("Unfortunately there is no content for the given URL!")),
			h.A(
				h.ID("go_home"), h.Href("/"),
				h.Button(h.Type("button"), h.Class("btn btn-lg btn-primary"), g.Text("Go back to Start-Page")),
			),
		),
	)

	return errorLayout(basePath, body)
}

// ErrorPage403 is used to indicate that the request lacks the necessary authorization to access the resource
func ErrorPage403(basePath string, env config.Environment) g.Node {
	body := h.Div(
		h.Class("container"),
		h.Div(
			h.Class("item"),
			h.Img(h.Src("/public/access-denied.svg"), h.Width("100px")),
			h.H2(g.Text("Access denied")),
			h.P(g.Text("You are not logged in or you do not have permission to access this page!")),
			h.A(
				h.ID("link-oidc-start"), h.Href("https://one.binggl.net/oidc/start"),
				h.Button(h.Type("button"), h.Class("btn btn-lg btn-warning"), g.Text("Login to access the page")),
			),
			g.If(env == config.Development, h.Div(
				h.Br(),
				h.A(h.ID("link-gettoken"), h.Href("/gettoken"), h.Button(h.Type("button"), h.Class("btn btn-lg btn-danger"), g.Text("Show me the JWT token for development"))),
			)),
			g.If(env == config.Integration, h.Div(
				h.Br(),
				h.A(h.ID("link-gettoken"), h.Href("https://dev.binggl.net/gettoken"), h.Button(h.Type("button"), h.Class("btn btn-lg btn-danger"), g.Text("Generate development token"))),
			)),
		),
	)

	return errorLayout(basePath, body)
}

// ErrorApplication is used for general errors
func ErrorApplication(basePath string, env config.Environment, startPage string, r *http.Request, err string) g.Node {
	headerKeys := make([]string, 0)
	for k := range r.Header {
		headerKeys = append(headerKeys, k)
	}

	body := h.Div(
		h.Class("applicationError"),
		h.H2(
			g.Raw(`<svg xmlns="http://www.w3.org/2000/svg" width="32" height="32" fill="currentColor" class="bi bi-bug" viewBox="0 0 16 16">
				<path d="M4.355.522a.5.5 0 0 1 .623.333l.291.956A4.979 4.979 0 0 1 8 1c1.007 0 1.946.298 2.731.811l.29-.956a.5.5 0 1 1 .957.29l-.41 1.352A4.985 4.985 0 0 1 13 6h.5a.5.5 0 0 0 .5-.5V5a.5.5 0 0 1 1 0v.5A1.5 1.5 0 0 1 13.5 7H13v1h1.5a.5.5 0 0 1 0 1H13v1h.5a1.5 1.5 0 0 1 1.5 1.5v.5a.5.5 0 1 1-1 0v-.5a.5.5 0 0 0-.5-.5H13a5 5 0 0 1-10 0h-.5a.5.5 0 0 0-.5.5v.5a.5.5 0 1 1-1 0v-.5A1.5 1.5 0 0 1 2.5 10H3V9H1.5a.5.5 0 0 1 0-1H3V7h-.5A1.5 1.5 0 0 1 1 5.5V5a.5.5 0 0 1 1 0v.5a.5.5 0 0 0 .5.5H3c0-1.364.547-2.601 1.432-3.503l-.41-1.352a.5.5 0 0 1 .333-.623zM4 7v4a4 4 0 0 0 3.5 3.97V7H4zm4.5 0v7.97A4 4 0 0 0 12 11V7H8.5zM12 6a3.989 3.989 0 0 0-1.334-2.982A3.983 3.983 0 0 0 8 2a3.983 3.983 0 0 0-2.667 1.018A3.989 3.989 0 0 0 4 6h8z"></path>
			</svg> Application Error occurred!`),
		),
		h.Br(),
		h.P(h.Class("h5 errorText"), g.Text("The current request could not be processed and led to an error!")),
		h.Br(),
		h.P(h.Class("h5 errorDetails"), g.Textf("%v", err)),
		h.Br(),
		h.Hr(),
		g.If(env == config.Development, h.Div(
			h.P(h.Class("h5"), h.Strong(g.Text("Request-Data"))),
			h.Ul(
				h.Li(g.Rawf("Host: <strong>%s</strong>", r.Host)),
				h.Li(g.Rawf("Method: <strong>%s</strong>", r.Method)),
				h.Li(g.Rawf("Request: <strong>%s</strong>", r.RequestURI)),
				h.Li(g.Rawf("Remote: <strong>%s</strong>", r.RemoteAddr)),
				h.Li(g.Rawf("Referer: <strong>%s</strong>", r.Referer())),
				h.Li(
					g.Text("Headers,"),
					h.Ul(
						g.Map(headerKeys, func(key string) g.Node {
							return h.Li(g.Rawf("%s: <strong>%s</strong>", key, r.Header.Get(key)))
						}),
					),
				),
			),
		)),
		h.A(h.ID("go_home"), h.Href(startPage), h.Button(h.Type("button"), h.Class("btn btn-lg btn-primary"), g.Text("Go back to Start-Page"))),
	)

	return errorLayout(basePath, body)
}

func errorLayout(basePath string, body g.Node) g.Node {
	return c.HTML5(c.HTML5Props{
		Title:       "404 - not found / binggl.net",
		Description: "The specified resource could not be found",
		Head: []g.Node{
			h.Base(h.Href("/")),
			h.Meta(g.Attr("creator", "https://www.gomponents.com/")),
			h.Link(h.Rel("shortcut icon"), h.ID("site-favicon"), h.Type("image/x-icon"), h.Href("public/folder.svg")),
			h.Link(h.Href(basePath+"/bootstrap/css/bootstrap.min.css"), h.Rel("stylesheet")),
			h.Link(h.Href(basePath+"/fonts/local.css"), h.Rel("stylesheet")),
			errorStyles(),
		},
		Body: []g.Node{
			h.Body(
				h.Div(
					h.Class("container appcontent"),
					h.Style("height: 100%"),
					h.Div(
						h.Class("row contentArea"),
						h.Div(
							h.Class("col-main bodyContent"),
							body,
						),
					),
				),
			),
		},
	})
}

//go:embed error.css
var errorCSS string

func errorStyles() g.Node {
	return h.StyleEl(
		h.Type("text/css"),
		g.Raw(errorCSS),
	)
}
