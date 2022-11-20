package bookmarks_test

import (
	"testing"

	"golang.binggl.net/monorepo/internal/bookmarks-new/app/bookmarks"
	"golang.binggl.net/monorepo/internal/bookmarks-new/app/store"
	"golang.binggl.net/monorepo/pkg/logging"
	"golang.binggl.net/monorepo/pkg/security"
)

var repo = store.NewMockRepository()

func Test_GetBookmark(t *testing.T) {
	svc := bookmarks.Service{
		Store:  repo,
		Logger: logging.NewNop(),
	}

	_, err := svc.GetBookmark("", security.User{})
	if err != nil {
		t.Errorf("error expected if no id supplied")
	}

}
