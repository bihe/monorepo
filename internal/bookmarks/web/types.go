package web

import "golang.binggl.net/monorepo/internal/bookmarks/app/bookmarks"

// PageModel combines common model data for pages
type PageModel struct {
	Development   bool
	Authenticated bool
	User          UserModel
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

// BookmarksSearchModel holds the search-results for bookmarks
type BookmarksSearchModel struct {
	PageModel
	Search        string
	Error         string
	VersionString string
	Bookmarks     []bookmarks.Bookmark
}
