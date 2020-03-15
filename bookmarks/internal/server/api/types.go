package api

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/bihe/bookmarks/internal/store"
	"github.com/go-chi/render"
)

// --------------------------------------------------------------------------
// Models
// --------------------------------------------------------------------------

// available node-types
// swagger:enum NodeType
type NodeType string

const (
	Node   NodeType = "Node"
	Folder NodeType = "Folder"
)

// Bookmark is the model provided via the REST API
// swagger:model
type Bookmark struct {
	ID          string     `json:"id"`
	Path        string     `json:"path"`
	DisplayName string     `json:"displayName"`
	URL         string     `json:"url"`
	SortOrder   int        `json:"sortOrder"`
	Type        NodeType   `json:"type"`
	Created     time.Time  `json:"created"`
	Modified    *time.Time `json:"modified,omitempty"`
	ChildCount  int        `json:"childCount"`
	AccessCount int        `json:"accessCount"`
	Favicon     string     `json:"favicon"`
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
		AccessCount: b.AccessCount,
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

// --------------------------------------------------------------------------
// Swagger specific definitions
// --------------------------------------------------------------------------

// swagger:parameters CreateBookmark UpdateBookmark
type BookMarkRequestSwagger struct {
	// In: body
	Body Bookmark
}

// swagger:parameters UpdateSortOrder
type BookmarksSortOrderRequestSwagger struct {
	// In: body
	Body BookmarksSortOrder
}

// --------------------------------------------------------------------------
// BookmarkRequest
// --------------------------------------------------------------------------

// BookmarkRequest is the request payload for Bookmark data model.
type BookmarkRequest struct {
	*Bookmark
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
	*BookmarksSortOrder
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

// --------------------------------------------------------------------------
// BookmarkResponse
// --------------------------------------------------------------------------

// BookmarkResponse is the response payload for the Bookmark data model.
type BookmarkResponse struct {
	*Bookmark
}

// Render the specific response
func (b BookmarkResponse) Render(w http.ResponseWriter, r *http.Request) error {
	// Pre-processing before a response is marshalled and sent across the wire
	return nil
}

// --------------------------------------------------------------------------
// BookmarkListResponse
// --------------------------------------------------------------------------

// BookmarkListResponse returns a list of Bookmark Items
type BookmarkListResponse struct {
	*BookmarkList
}

// Render the specific response
func (b BookmarkListResponse) Render(w http.ResponseWriter, r *http.Request) error {
	// Pre-processing before a response is marshalled and sent across the wire
	return nil
}

// --------------------------------------------------------------------------
// BookmarResultResponse
// --------------------------------------------------------------------------

// BookmarResultResponse returns BookmarResult
type BookmarResultResponse struct {
	*BookmarkResult
}

// Render the specific response
func (b BookmarResultResponse) Render(w http.ResponseWriter, r *http.Request) error {
	// Pre-processing before a response is marshalled and sent across the wire
	return nil
}

// --------------------------------------------------------------------------
// BookmarksPathsResponse
// --------------------------------------------------------------------------

// BookmarksPathsResponse returns available Paths
// swagger:model
type BookmarksPathsResponse struct {
	Paths []string `json:"paths"`
	Count int      `json:"count"`
}

// Render the specific response
func (b BookmarksPathsResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

// --------------------------------------------------------------------------
// ResultResponse
// --------------------------------------------------------------------------

// ResultResponse returns Result
type ResultResponse struct {
	*Result
	Status int `json:"-"` // ignore this
}

// Render the specific response
func (b ResultResponse) Render(w http.ResponseWriter, r *http.Request) error {
	if b.Status == 0 {
		render.Status(r, http.StatusOK)
	} else {
		render.Status(r, b.Status)
	}
	return nil
}
