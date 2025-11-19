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
	// FileItem is similar to a Node, but is defined by a file
	FileItem
)

// NodeCount displays the number of child-elements for a given path (1 level)
type NodeCount struct {
	Path  string
	Count int
}

// Bookmark maps the database table to a struct
type Bookmark struct {
	ID                 string     `gorm:"primary_key;TYPE:varchar(255);COLUMN:id"`
	Path               string     `gorm:"TYPE:varchar(255);COLUMN:path;NOT NULL;INDEX:IX_PATH;INDEX:IX_PATH_USER"`
	DisplayName        string     `gorm:"TYPE:varchar(128);COLUMN:display_name;NOT NULL"`
	URL                string     `gorm:"TYPE:varchar(512);COLUMN:url;NOT NULL;INDEX:IX_SORT_ORDER"`
	SortOrder          int        `gorm:"COLUMN:sort_order;DEFAULT:0;NOT NULL"`
	Type               NodeType   `gorm:"COLUMN:type;DEFAULT:0;NOT NULL"`
	UserName           string     `gorm:"TYPE:varchar(128);COLUMN:user_name;NOT NULL;INDEX:IX_USER;INDEX:IX_PATH_USER"`
	Created            time.Time  `gorm:"COLUMN:created;NOT NULL"`
	Modified           *time.Time `gorm:"COLUMN:modified"`
	ChildCount         int        `gorm:"COLUMN:child_count;DEFAULT:0;NOT NULL"`
	Highlight          int        `gorm:"COLUMN:highlight;DEFAULT:0;NOT NULL"`
	Favicon            string     `gorm:"TYPE:varchar(128);COLUMN:favicon;"`
	InvertFaviconColor int        `gorm:"COLUMN:invert_favicon_color;DEFAULT:0;NOT NULL"`
	File               *File      `gorm:"foreignKey:FileID;default:SET NULL;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	FileID             *string
}

func (b Bookmark) String() string {
	return fmt.Sprintf("Bookmark: '%s, %s' (Id: %s, Type: %d)", b.Path, b.DisplayName, b.ID, b.Type)
}

// TableName specifies the name of the Table used
func (Bookmark) TableName() string {
	return "BOOKMARKS"
}

// Favicon stores the fetched favicons as a binary payload
type Favicon struct {
	ID           string    `gorm:"primary_key;TYPE:varchar(128);COLUMN:id;NOT NULL"`
	Payload      []byte    `gorm:"COLUMN:payload;TYPE:bytes;NOT NULL"`
	LastModified time.Time `gorm:"COLUMN:modified;NOT NULL"`
}

func (Favicon) TableName() string {
	return "FAVICONS"
}

// A File represents a binary ot text file (defined by the mime-type)
type File struct {
	ID           string      `gorm:"primary_key;TYPE:varchar(128);COLUMN:id;NOT NULL"`
	Name         string      `gorm:"TYPE:varchar(128);COLUMN:name;NOT NULL"`
	MimeType     string      `gorm:"TYPE:varchar(128);COLUMN:mime_type;NOT NULL"`
	Size         int         `gorm:"COLUMN:size;DEFAULT:0;NOT NULL"`
	Modified     time.Time   `gorm:"COLUMN:modified;NOT NULL"`
	FileObject   *FileObject `gorm:"foreignKey:FileObjectID;default:SET NULL;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	FileObjectID *string
}

func (File) TableName() string {
	return "FILES"
}

// A FileObject holds the actual file payload. It is used to separate the file meta-data from the payload.
// As a result Bookmark and File can be happily joined, without fetching the full payload.
type FileObject struct {
	ID      string `gorm:"primary_key;TYPE:varchar(128);COLUMN:id;NOT NULL"`
	Payload []byte `gorm:"COLUMN:payload;TYPE:bytes;NOT NULL"`
}

func (FileObject) TableName() string {
	return "FILEOBJECTS"
}
