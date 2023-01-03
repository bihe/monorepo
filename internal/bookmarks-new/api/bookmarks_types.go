package api

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"golang.binggl.net/monorepo/internal/bookmarks-new/app/bookmarks"
)

// BookmarkFolderResult has additional information about a Bookmark
type BookmarkFolderResult struct {
	Success bool               `json:"success"`
	Message string             `json:"message"`
	Value   bookmarks.Bookmark `json:"value"`
}

// BookmarkList is a collection of Bookmarks
type BookmarkList struct {
	Success bool                 `json:"success"`
	Count   int                  `json:"count"`
	Message string               `json:"message"`
	Value   []bookmarks.Bookmark `json:"value"`
}

// Result is a generic response with a string value
type Result struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Value   string `json:"value"`
}

// BookmarksPaths encapsulates the available bookmarks paths
type BookmarksPaths struct {
	Paths []string `json:"paths"`
	Count int      `json:"count"`
}

// --------------------------------------------------------------------------
// BookmarkRequest
// --------------------------------------------------------------------------

// BookmarkRequest is the request payload for Bookmark data model.
type BookmarkRequest struct {
	*bookmarks.Bookmark
}

// Bind assigns the the provided data to a BookmarkRequest
func (b *BookmarkRequest) Bind(r *http.Request) error {
	// a Bookmark is nil if no Bookmark fields are sent in the request. Return an
	// error to avoid a nil pointer dereference.
	if b.Bookmark == nil {
		return fmt.Errorf("missing required Bookmarks fields")
	}
	return nil
}

// String returns a string representation of a BookmarkRequest
func (b *BookmarkRequest) String() string {
	return fmt.Sprintf("Bookmark: '%s, %s' (Id: %s, Type: %s)", b.Path, b.DisplayName, b.ID, b.Type)
}

// --------------------------------------------------------------------------
// BookmarksSortOrderRequest
// --------------------------------------------------------------------------

// BookmarksSortOrderRequest is the request payload for Bookmark data model.
type BookmarksSortOrderRequest struct {
	*bookmarks.BookmarksSortOrder
}

// Bind assigns the the provided data to a BookmarksSortOrderRequest
func (b *BookmarksSortOrderRequest) Bind(r *http.Request) error {
	if b.BookmarksSortOrder == nil {
		return fmt.Errorf("missing required BookmarksSortOrder fields")
	}
	return nil
}

// String returns a string representation of a BookmarkRequest
func (b *BookmarksSortOrderRequest) String() string {
	var order []string
	for i := range b.SortOrder {
		order = append(order, strconv.Itoa(i))
	}
	return fmt.Sprintf("IDs: '%s', SortOrder: %s", strings.Join(b.IDs, ","), strings.Join(order, ","))
}
