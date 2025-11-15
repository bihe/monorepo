package bookmarks

import (
	"fmt"
	"time"

	"golang.binggl.net/monorepo/internal/bookmarks/app/store"
)

// --------------------------------------------------------------------------
// Models
// --------------------------------------------------------------------------

// NodeType specifies available node-types
type NodeType string

const (
	// Node is the standard item
	Node NodeType = "Node"
	// Folder is used to group nodes
	Folder NodeType = "Folder"
	// File is similar to node but has a binary payload
	FileItem NodeType = "File"
)

// File represents a binary payload
type File struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	MimeType string `json:"mimeType"`
	Payload  []byte `json:"payload"`
	Size     int    `json:"size"`
}

// Bookmark is the model provided via the REST API
type Bookmark struct {
	ID                 string     `json:"id"`
	Path               string     `json:"path"`
	DisplayName        string     `json:"displayName"`
	URL                string     `json:"url"`
	SortOrder          int        `json:"sortOrder"`
	Type               NodeType   `json:"type"`
	Created            time.Time  `json:"created"`
	Modified           *time.Time `json:"modified,omitempty"`
	ChildCount         int        `json:"childCount"`
	Highlight          int        `json:"highlight"`
	Favicon            string     `json:"favicon"`
	InvertFaviconColor int        `json:"invertFaviconColor"`
	File               *File      `json:"file,omitempty"`
}

// TStamp returns either the modified or created timestamp of the bookmark as unix time
func (b Bookmark) TStamp() string {
	var tstamp time.Time
	tstamp = b.Created
	if b.Modified != nil {
		tstamp = *b.Modified
	}
	return fmt.Sprintf("%d", tstamp.Unix())
}

// BookmarksSortOrder contains a sorting for a list of ids
type BookmarksSortOrder struct {
	IDs       []string `json:"ids"`
	SortOrder []int    `json:"sortOrder"`
}

// ObjectInfo is used to encapsulate information about a payload/file
type ObjectInfo struct {
	Payload  []byte
	Name     string
	Modified time.Time
}

// --------------------------------------------------------------------------
// convert entities to models
// --------------------------------------------------------------------------

func entityToModel(b store.Bookmark) *Bookmark {
	return &Bookmark{
		ID:                 b.ID,
		DisplayName:        b.DisplayName,
		Path:               b.Path,
		Type:               entityEnumToModel(b.Type),
		URL:                b.URL,
		SortOrder:          b.SortOrder,
		Created:            b.Created,
		Modified:           b.Modified,
		Highlight:          b.Highlight,
		ChildCount:         b.ChildCount,
		Favicon:            b.Favicon,
		InvertFaviconColor: b.InvertFaviconColor,
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
