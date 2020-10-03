package shared

import "golang.binggl.net/monorepo/pkg/persistence"

// BaseRepository defines the common methods of all implementations
type BaseRepository interface {
	// CreateAtomic returns a new atomic object
	CreateAtomic() (persistence.Atomic, error)
}
