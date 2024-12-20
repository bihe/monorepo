package common

import (
	"net/http"

	"golang.binggl.net/monorepo/pkg/security"
)

func EnsureUser(r *http.Request) *security.User {
	user, ok := security.UserFromContext(r.Context())
	if !ok || user == nil {
		panic("the security context user is not available!")
	}
	return user
}
