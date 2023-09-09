package web

import "golang.binggl.net/monorepo/internal/bookmarks/app/bookmarks"

// PageModel combines common model data for pages
type PageModel struct {
	Development   bool
	Authenticated bool
	User          UserModel
	VersionString string
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
	Path string
}

// BookmarksSearchModel holds the search-results for bookmarks
type BookmarksSearchModel struct {
	PageModel
	BookmarkResultModel
	Search string
}
