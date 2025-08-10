package common

import (
	"strings"

	"golang.binggl.net/monorepo/pkg/config"
	"golang.binggl.net/monorepo/pkg/handler/html"
	"golang.binggl.net/monorepo/pkg/security"
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

// AvailableApps statically defines the available Applications/Navigation-Pages
var AvailableApps = []html.NavItem{
	{
		DisplayName: "Bookmarks",
		Icon:        "<i class=\"bi bi-bookmark-star\"></i> ",
		URL:         "/bm",
	},
	{
		DisplayName: "Documents",
		Icon:        "<i class=\"bi bi-file-earmark-pdf\"></i> ",
		URL:         "/mydms",
	},
	{
		DisplayName: "Sites",
		Icon:        "<i class=\"bi bi-diagram-2\"></i> ",
		URL:         "/sites",
	},
	{
		DisplayName: "Encryption",
		Icon:        "<i class=\"bi bi-file-lock\"></i> ",
		URL:         "/crypter",
	},
}

// CreatePageModel provides the needed data for a page using the shared Layout
func CreatePageModel(pageURL, pageTitle, search, favicon, timeStamp, commit string, env config.Environment, user security.User) html.LayoutModel {
	appNav := make([]html.NavItem, 0)
	var title string
	for _, a := range AvailableApps {
		if a.URL == pageURL {
			a.Active = true
			title = a.DisplayName
		}
		appNav = append(appNav, html.NavItem{DisplayName: a.DisplayName, Icon: a.Icon, URL: a.URL, Active: a.Active})
	}
	if pageTitle == "" {
		pageTitle = title
	}
	model := html.LayoutModel{
		PageTitle:  pageTitle,
		Favicon:    favicon,
		TimeStamp:  timeStamp,
		Commit:     commit,
		User:       user,
		Search:     search,
		Navigation: appNav,
	}
	model.Env = env
	if model.Favicon == "" {
		model.Favicon = "/public/folder.svg"
	}
	return model
}

// HtmxIndicatorNode provides the gomponents code for a htmx indicator element
func HtmxIndicatorNode() g.Node {
	return h.Div(h.ID("indicator"), h.Class("htmx-indicator"),
		h.Div(h.Class("spinner-border text-light"), h.Role("status"),
			h.Span(h.Class("visually-hidden"), g.Text("Loading...")),
		),
	)
}

// Ellipsis cuts a string at a given length
func Ellipsis(entry string, length int, indicator string) string {
	if entry == "" {
		return ""
	}
	if len(entry) < length {
		return entry
	}
	return entry[:length] + indicator
}

// SubString cuts a given string at a length
func SubString(entry string, length int) string {
	if entry == "" {
		return ""
	}
	if len(entry) < length {
		return entry
	}
	return entry[:length]
}

// EnsureTrailingSlash checks the entry to end with a slash
func EnsureTrailingSlash(entry string) string {
	if strings.HasSuffix(entry, "/") {
		return entry
	}
	return entry + "/"
}

// ClassCond adds the conditional class as a string if the provided condition evaluates true
func ClassCond(starter, conditional string, condition bool) string {
	classes := make([]string, 1)
	classes = append(classes, starter)
	if condition {
		classes = append(classes, conditional)
	}
	return strings.Join(classes, " ")
}
