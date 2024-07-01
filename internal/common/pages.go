package common

import "golang.binggl.net/monorepo/pkg/handler/templates"

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
		Icon:        "<i class=\"bi bi-shield-check\"></i> ",
		URL:         "/sites",
	},
	{
		DisplayName: "Tools",
		Icon:        "<i class=\"bi bi-wrench-adjustable-circle\"></i> ",
		URL:         "/tools",
	},
}
