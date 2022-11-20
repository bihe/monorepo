package api

import "golang.binggl.net/monorepo/pkg/logging"

// BookmarksHandlers implements the API used for bookmarks
type BookmarksHandler struct {
	Logger logging.Logger
}

// func (b *BookmarksAPI) GetBookmarkByID(user security.User, w http.ResponseWriter, r *http.Request) error {
// 	id := chi.URLParam(r, "id")

// 	if id == "" {
// 		return errors.BadRequestError{Err: fmt.Errorf("missing id parameter"), Request: r}
// 	}
// 	b.Log.InfoRequest(fmt.Sprintf("try to get bookmark by ID: '%s' for user: '%s'", id, user.Username), r)
// 	bookmark, err := b.Repository.GetBookmarkByID(id, user.Username)
// 	if err != nil {
// 		b.Log.InfoRequest(fmt.Sprintf("try to get bookmark by ID: '%s' for user: '%s'", id, user.Username), r)
// 		return errors.NotFoundError{Err: fmt.Errorf("no bookmark with ID '%s' avaliable", id), Request: r}
// 	}

// 	return render.Render(w, r, BookmarkResponse{Bookmark: entityToModel(bookmark)})
// }
