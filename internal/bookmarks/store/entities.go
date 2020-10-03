package store

import (
	"fmt"
	"time"
)

// NodeType identifies a node
type NodeType int

const (
	// Node is the std value - a node in a tree
	Node NodeType = iota
	// Folder defines a structure/grouping of nodes
	Folder
)

// Bookmark maps the database table to a struct
type Bookmark struct {
	ID          string     `gorm:"primary_key;TYPE:varchar(255);COLUMN:id"`
	Path        string     `gorm:"TYPE:varchar(255);COLUMN:path;NOT NULL;INDEX:IX_PATH;INDEX:IX_PATH_USER"`
	DisplayName string     `gorm:"TYPE:varchar(128);COLUMN:display_name;NOT NULL"`
	URL         string     `gorm:"TYPE:varchar(512);COLUMN:url;NOT NULL;INDEX:IX_SORT_ORDER"`
	SortOrder   int        `gorm:"COLUMN:sort_order;DEFAULT:0;NOT NULL"`
	Type        NodeType   `gorm:"COLUMN:type;DEFAULT:0;NOT NULL"`
	UserName    string     `gorm:"TYPE:varchar(128);COLUMN:user_name;NOT NULL;INDEX:IX_USER;INDEX:IX_PATH_USER"`
	Created     time.Time  `gorm:"COLUMN:created;NOT NULL"`
	Modified    *time.Time `gorm:"COLUMN:modified"`
	ChildCount  int        `gorm:"COLUMN:child_count;DEFAULT:0;NOT NULL"`
	AccessCount int        `gorm:"COLUMN:access_count;DEFAULT:0;NOT NULL"`
	Favicon     string     `gorm:"TYPE:varchar(128);COLUMN:favicon;NOT NULL"`
}

func (b Bookmark) String() string {
	return fmt.Sprintf("Bookmark: '%s, %s' (Id: %s, Type: %d)", b.Path, b.DisplayName, b.ID, b.Type)
}

// TableName specifies the name of the Table used
func (Bookmark) TableName() string {
	return "BOOKMARKS"
}

// NodeCount displays the number of child-elements for a given path (1 level)
type NodeCount struct {
	Path  string
	Count int
}
