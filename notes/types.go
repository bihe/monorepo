package notes

import "time"

// Note defines the data of a entry
type Note struct {
	ID       string
	User     string
	Created  time.Time
	Modified *time.Time
	Title    string
	Data     string
	Tags     []string
}
