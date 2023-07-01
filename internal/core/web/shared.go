package web

import (
	"net/http"

	"golang.binggl.net/monorepo/pkg/security"
)

func ensureUser(r *http.Request) *security.User {
	user, ok := security.UserFromContext(r.Context())
	if !ok || user == nil {
		panic("the sucurity context user is not available!")
	}
	return user
}
