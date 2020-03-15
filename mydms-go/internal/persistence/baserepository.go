package persistence

// BaseRepository defines the common methods of all implementations
type BaseRepository interface {
	// CreateAtomic returns a new atomic object
	CreateAtomic() (Atomic, error)
}
