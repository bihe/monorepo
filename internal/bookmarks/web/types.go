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

const (
	MessageTypeInfo    = "info"
	MessageTypeError   = "error"
	MessageTypeSuccess = "success"
)

// MessageModel is used to provide information to the UI
type MessageModel struct {
	Title string
	Text  string
	Type  string
	Show  bool
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
	Message   MessageModel
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

// ConfirmDeleteModel is used for confirm dialogs
type ConfirmDeleteModel struct {
	Item string
	ID   string
}

// EditBookmarkModel is used to create or edit a bookmark
type EditBookmarkModel struct {
	bookmarks.Bookmark
}
