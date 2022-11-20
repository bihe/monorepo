package bookmarks

import (
	"time"

	"golang.binggl.net/monorepo/internal/bookmarks-new/app/store"
)

// --------------------------------------------------------------------------
// Models
// --------------------------------------------------------------------------

// NodeType specifies available node-types
type NodeType string

const (
	// Node is the standard item
	Node NodeType = "Node"
	// Folder is ised to group nodes
	Folder NodeType = "Folder"
)

// Bookmark is the model provided via the REST API
// swagger:model
type Bookmark struct {
	ID            string     `json:"id"`
	Path          string     `json:"path"`
	DisplayName   string     `json:"displayName"`
	URL           string     `json:"url"`
	SortOrder     int        `json:"sortOrder"`
	Type          NodeType   `json:"type"`
	Created       time.Time  `json:"created"`
	Modified      *time.Time `json:"modified,omitempty"`
	ChildCount    int        `json:"childCount"`
	Highlight     int        `json:"highlight"`
	Favicon       string     `json:"favicon"`
	CustomFavicon string     `json:"customFavicon"`
}

// BookmarkList is a collection of Bookmarks
// swagger:model
type BookmarkList struct {
	Success bool       `json:"success"`
	Count   int        `json:"count"`
	Message string     `json:"message"`
	Value   []Bookmark `json:"value"`
}

// BookmarkResult has additional information about a Bookmark
// swagger:model
type BookmarkResult struct {
	Success bool     `json:"success"`
	Message string   `json:"message"`
	Value   Bookmark `json:"value"`
}

// Result is a generic response with a string value
// swagger:model
type Result struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Value   string `json:"value"`
}

// BookmarksSortOrder contains a sorting for a list of ids
// swagger:model
type BookmarksSortOrder struct {
	IDs       []string `json:"ids"`
	SortOrder []int    `json:"sortOrder"`
}

// --------------------------------------------------------------------------
// convert entities to models
// --------------------------------------------------------------------------

func entityToModel(b store.Bookmark) *Bookmark {
	return &Bookmark{
		ID:          b.ID,
		DisplayName: b.DisplayName,
		Path:        b.Path,
		Type:        entityEnumToModel(b.Type),
		URL:         b.URL,
		SortOrder:   b.SortOrder,
		Created:     b.Created,
		Modified:    b.Modified,
		Highlight:   b.Highlight,
		ChildCount:  b.ChildCount,
		Favicon:     b.Favicon,
	}
}

func entityListToModel(bms []store.Bookmark) []Bookmark {
	var model = make([]Bookmark, 0)
	for _, b := range bms {
		model = append(model, *entityToModel(b))
	}
	return model
}

func entityEnumToModel(t store.NodeType) NodeType {
	if t == store.Folder {
		return Folder
	}
	return Node
}
