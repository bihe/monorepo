package api_test

import "golang.binggl.net/monorepo/internal/bookmarks/app/store"

// --------------------------------------------------------------------------
// mock repository implementation
// --------------------------------------------------------------------------

// create a mock implementation of the bookmark repository
// tests can use the mock to implement their desired behavior
type mockBookmarkRepository struct{}

var _ store.BookmarkRepository = (*mockBookmarkRepository)(nil)

func (m *mockBookmarkRepository) CheckStoreConnectivity(timeOut uint) (err error) {
	return nil
}

func (m *mockBookmarkRepository) InUnitOfWork(fn func(repo store.BookmarkRepository) error) error {
	return nil
}

func (m *mockBookmarkRepository) Create(item store.Bookmark) (store.Bookmark, error) {
	return store.Bookmark{}, nil
}

func (m *mockBookmarkRepository) Update(item store.Bookmark) (store.Bookmark, error) {
	return store.Bookmark{}, nil
}

func (m *mockBookmarkRepository) Delete(item store.Bookmark) error {
	return nil
}

func (m *mockBookmarkRepository) DeletePath(path, username string) error {
	return nil
}

func (m *mockBookmarkRepository) GetAllBookmarks(username string) ([]store.Bookmark, error) {
	return nil, nil
}

func (m *mockBookmarkRepository) GetBookmarksByPath(path, username string) ([]store.Bookmark, error) {
	return nil, nil
}

func (m *mockBookmarkRepository) GetBookmarksByPathStart(path, username string) ([]store.Bookmark, error) {
	return nil, nil
}

func (m *mockBookmarkRepository) GetBookmarksByName(name, username string) ([]store.Bookmark, error) {
	return nil, nil
}

func (m *mockBookmarkRepository) GetMostRecentBookmarks(username string, limit int) ([]store.Bookmark, error) {
	return nil, nil
}

func (m *mockBookmarkRepository) GetPathChildCount(path, username string) ([]store.NodeCount, error) {
	return nil, nil
}

func (m *mockBookmarkRepository) GetBookmarkByID(id, username string) (store.Bookmark, error) {
	return store.Bookmark{}, nil
}

func (m *mockBookmarkRepository) GetFolderByPath(path, username string) (store.Bookmark, error) {
	return store.Bookmark{}, nil
}

func (m *mockBookmarkRepository) GetAllPaths(username string) ([]string, error) {
	return nil, nil
}

// --------------------------------------------------------------------------

// create a mock implementation of the favicon repository
// tests can use the mock to implement their desired behavior
type mockFaviconRepository struct{}

var _ store.FaviconRepository = (*mockFaviconRepository)(nil)

func (m *mockFaviconRepository) InUnitOfWork(fn func(repo store.FaviconRepository) error) error {
	return nil
}

func (m *mockFaviconRepository) Get(fid string) (store.Favicon, error) {
	return store.Favicon{}, nil
}

func (m *mockFaviconRepository) Save(item store.Favicon) (store.Favicon, error) {
	return store.Favicon{}, nil
}

func (m *mockFaviconRepository) Delete(item store.Favicon) error {
	return nil
}
