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

// EditBookmarkModel is used to create or edit a bookmark
type EditBookmarkModel struct {
	bookmarks.Bookmark
}
