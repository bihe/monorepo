package web

import (
	"html/template"

	"golang.binggl.net/monorepo/internal/bookmarks/app/bookmarks"
)

// PageModel combines common model data for pages
type PageModel struct {
	Development        bool
	Authenticated      bool
	User               UserModel
	VersionString      string
	PageReloadClientJS template.HTML
}

// UserModel provides user information for pages
type UserModel struct {
	Username    string
	Email       string
	UserID      string
	DisplayName string
	ProfileURL  string
	Token       string
}

// BookmarkResultModel holds results from a bookmark operation
type BookmarkResultModel struct {
	Error     string
	Bookmarks []bookmarks.Bookmark
}

// BookmarkPathModel is used to get bookmarks for a given path
type BookmarkPathModel struct {
	PageModel
	BookmarkResultModel
	Path []BookmarkPathEntry
}

// BookmarkPathEntry is used to display the path hierarchy
type BookmarkPathEntry struct {
	UrlPath     string
	DisplayName string
	LastItem    bool
}

// BookmarksSearchModel holds the search-results for bookmarks
type BookmarksSearchModel struct {
	PageModel
	BookmarkResultModel
	Search string
}

// ConfirmDeleteModel is used for confirm dialogs
type ConfirmDeleteModel struct {
	Item string
	Path string
}
