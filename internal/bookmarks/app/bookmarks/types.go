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

// Bookmark is the model provided via the REST API
type Bookmark struct {
	ID                 string       `json:"id"`
	Path               string       `json:"path"`
	DisplayName        string       `json:"displayName"`
	URL                string       `json:"url"`
	SortOrder          int          `json:"sortOrder"`
	Type               NodeType     `json:"type"`
	Created            time.Time    `json:"created"`
	Modified           *time.Time   `json:"modified,omitempty"`
	ChildCount         int          `json:"childCount"`
	Highlight          int          `json:"highlight"`
	Favicon            string       `json:"favicon"`
	InvertFaviconColor int          `json:"invertFaviconColor"`
	FileID             string       `json:"file,omitempty"`
	FilePayload        *FilePayload `json:"file_payload,omitempty"`
}

// A FilePayload represents a saved file used with a bookmark
type FilePayload struct {
	ID       string    `json:"id,omitempty"`
	Name     string    `json:"name,omitempty"`
	MimeType string    `json:"mime_type,omitempty"`
	Payload  []byte    `json:"payload,omitempty"`
	Size     int       `json:"size,omitempty"`
	Modified time.Time `json:"modified,omitempty"`
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
	var (
		fileID      string
		filePayload *FilePayload
	)
	if b.FileID != nil {
		fileID = *b.FileID
	}
	if b.File != nil {
		file := FilePayload{
			ID:       b.File.ID,
			Name:     b.File.Name,
			MimeType: b.File.MimeType,
			Payload:  b.File.Payload,
			Size:     b.File.Size,
		}
		filePayload = &file
	}
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
		FileID:             fileID,
		FilePayload:        filePayload,
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
	var nodeType NodeType

	switch t {
	case store.Node:
		nodeType = Node
	case store.Folder:
		nodeType = Folder
	case store.FileItem:
		nodeType = FileItem
	}
	return nodeType
}
