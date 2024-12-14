package common

import (
	"golang.binggl.net/monorepo/pkg/config"
	"golang.binggl.net/monorepo/pkg/handler/html"
	"golang.binggl.net/monorepo/pkg/handler/templates"
	"golang.binggl.net/monorepo/pkg/security"
)

// statically define the available Applications/Navigation-Pages
var AvailableApps = []templates.NavItem{
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
}

func CreatePageModel(pageURL, pageTitle, search, favicon, version string, env config.Environment, user security.User) html.LayoutModel {
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
		Version:    version,
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
