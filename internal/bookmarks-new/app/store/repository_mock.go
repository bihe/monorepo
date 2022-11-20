package store

import "fmt"

// NewMockRepository prepares the in-memory storage to work with the mock repository
func NewMockRepository() Repository {
	return &mockRepo{
		store: make(map[string][]Bookmark),
	}
}

var _ Repository = &mockRepo{}

type mockRepo struct {
	store map[string][]Bookmark
}

func (m *mockRepo) InUnitOfWork(fn func(repo Repository) error) error {
	return nil
}

func (m *mockRepo) Create(item Bookmark) (Bookmark, error) {
	usrBookmarks := m.store[item.UserName]
	usrBookmarks = append(usrBookmarks, item)
	m.store[item.UserName] = usrBookmarks
	return item, nil
}

func (m *mockRepo) Update(item Bookmark) (Bookmark, error) {
	usrBookmarks := m.store[item.UserName]
	if len(usrBookmarks) == 0 {
		return Bookmark{}, fmt.Errorf("bookmark not found for user")
	}
	idx := -1
	for i, elem := range usrBookmarks {
		if elem.ID == item.ID {
			idx = i
			break
		}
	}
	if idx < 0 {
		return Bookmark{}, fmt.Errorf("bookmark not found")
	}
	usrBookmarks[idx] = item
	return usrBookmarks[idx], nil
}

func (m *mockRepo) Delete(item Bookmark) error {
	usrBookmarks := m.store[item.UserName]
	if len(usrBookmarks) == 0 {
		return fmt.Errorf("bookmark not found for user")
	}
	idx := -1
	for i, elem := range usrBookmarks {
		if elem.ID == item.ID {
			idx = i
			break
		}
	}
	if idx < 0 {
		return fmt.Errorf("bookmark not found")
	}

	copy(usrBookmarks[idx:], usrBookmarks[idx+1:])
	m.store[item.UserName] = usrBookmarks[:len(usrBookmarks)-1]

	return nil
}

func (m *mockRepo) DeletePath(path, username string) error {
	return nil
}

func (m *mockRepo) GetAllBookmarks(username string) ([]Bookmark, error) {
	usrBookmarks := m.store[username]
	if len(usrBookmarks) == 0 {
		return nil, fmt.Errorf("bookmark not found for user")
	}
	return usrBookmarks, nil
}

func (m *mockRepo) GetBookmarksByPath(path, username string) ([]Bookmark, error) {
	return make([]Bookmark, 0), nil
}

func (m *mockRepo) GetBookmarksByPathStart(path, username string) ([]Bookmark, error) {
	return make([]Bookmark, 0), nil
}

func (m *mockRepo) GetBookmarksByName(name, username string) ([]Bookmark, error) {
	return make([]Bookmark, 0), nil
}

func (m *mockRepo) GetPathChildCount(path, username string) ([]NodeCount, error) {
	return make([]NodeCount, 0), nil
}

func (m *mockRepo) GetAllPaths(username string) ([]string, error) {
	return make([]string, 0), nil
}

func (m *mockRepo) GetBookmarkByID(id, username string) (Bookmark, error) {
	usrBookmarks := m.store[username]
	if len(usrBookmarks) == 0 {
		return Bookmark{}, fmt.Errorf("bookmark not found for user")
	}
	idx := -1
	for i, elem := range usrBookmarks {
		if elem.ID == id {
			idx = i
			break
		}
	}
	if idx < 0 {
		return Bookmark{}, fmt.Errorf("bookmark not found")
	}

	return usrBookmarks[idx], nil
}

func (m *mockRepo) GetFolderByPath(path, username string) (Bookmark, error) {
	return Bookmark{}, nil
}
