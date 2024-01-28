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
		URL:         "/ui",
	},
	{
		DisplayName: "Sites",
		Icon:        "<i class=\"bi bi-diagram-2\"></i> ",
		URL:         "/sites",
	},
}
