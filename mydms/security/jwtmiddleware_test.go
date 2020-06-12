package security

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"

	sec "golang.binggl.net/monorepo/pkg/security"
	"golang.binggl.net/monorepo/mydms/errors"
)

const authHeader = "Authorization"
const bearer = "Bearer "

func TestJwtMiddleware(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	config := JwtOptions{
		JwtSecret:  "secret",
		JwtIssuer:  "test",
		CookieName: "test",
		RequiredClaim: sec.Claim{
			Name:  "claim",
			URL:   "http://localhost",
			Roles: []string{"roleA"},
		},
		RedirectURL:   "http://localhost?redirect",
		CacheDuration: "10s",
	}

	h := JwtWithConfig(config)(func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	})

	assert := assert.New(t)

	// no token / no cookie
	re := h(c).(errors.RedirectError)
	assert.Equal(re.URL, config.RedirectURL)

	// no token / invalid cookie
	cookie := &http.Cookie{
		Name:   "any_cookie",
		Domain: "localhost",
		Path:   "/",
	}
	req.AddCookie(cookie)
	re = h(c).(errors.RedirectError)
	assert.Equal(re.URL, config.RedirectURL)

	// invalid token / no cookie
	cookie.Expires = time.Unix(0, 0)
	req.Header.Set(authHeader, "Bearer token")
	re = h(c).(errors.RedirectError)
	assert.Equal(re.URL, config.RedirectURL)

	// valid token, missing claim
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJqdGkiOiI3NzAyNmYxMWM2ODQ0YzVlOWEyYmM1YWNkNGQxYjljNyIsImlhdCI6MTU0MDcyODAxMiwiaXNzIjoidGVzdCIsImV4cCI6MTc0MTMzMjgxMiwic3ViIjoiaGVucmlrQGEuYi5jLmRlIiwiVHlwZSI6ImxvZ2luLlVzZXIiLCJVc2VyTmFtZSI6InVzZXIubmFtZSIsIkVtYWlsIjoiaGVucmlrQGEuYi5jLmRlIiwiQ2xhaW1zIjpbImNsYWltfGh0dHA6Ly9sb2NhbGhvc3R8QUEiXSwiVXNlcklkIjoiMTExIiwiRGlzcGxheU5hbWUiOiJVc2VyMSBOYW1lIiwiU3VybmFtZSI6Ik5hbWUiLCJHaXZlbk5hbWUiOiJVc2VyMSJ9.oMcQdAipJFJL9Q2ACh1Rav_I9FyDWBbc2vnI4rAkhKY"
	req.Header.Set(authHeader, bearer+token)
	re = h(c).(errors.RedirectError)
	assert.Equal(re.URL, config.RedirectURL)

	// valid token
	token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJqdGkiOiI3NzAyNmYxMWM2ODQ0YzVlOWEyYmM1YWNkNGQxYjljNyIsImlhdCI6MTU0MDcyODAxMiwiaXNzIjoidGVzdCIsImV4cCI6MTc0MTMzMjgxMiwic3ViIjoiaGVucmlrQGEuYi5jLmRlIiwiVHlwZSI6ImxvZ2luLlVzZXIiLCJVc2VyTmFtZSI6InVzZXIubmFtZSIsIkVtYWlsIjoiaGVucmlrQGEuYi5jLmRlIiwiQ2xhaW1zIjpbImNsYWltfGh0dHA6Ly9sb2NhbGhvc3R8cm9sZUEiXSwiVXNlcklkIjoiMTExIiwiRGlzcGxheU5hbWUiOiJVc2VyMSBOYW1lIiwiU3VybmFtZSI6Ik5hbWUiLCJHaXZlbk5hbWUiOiJVc2VyMSJ9._gNI2zMSop__flAeUguOJbbwiP8IDzFMi0VltVvm13o"
	req.Header.Set(authHeader, bearer+token)
	assert.NoError(h(c))

	// call again to check cache
	token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJqdGkiOiI3NzAyNmYxMWM2ODQ0YzVlOWEyYmM1YWNkNGQxYjljNyIsImlhdCI6MTU0MDcyODAxMiwiaXNzIjoidGVzdCIsImV4cCI6MTc0MTMzMjgxMiwic3ViIjoiaGVucmlrQGEuYi5jLmRlIiwiVHlwZSI6ImxvZ2luLlVzZXIiLCJVc2VyTmFtZSI6InVzZXIubmFtZSIsIkVtYWlsIjoiaGVucmlrQGEuYi5jLmRlIiwiQ2xhaW1zIjpbImNsYWltfGh0dHA6Ly9sb2NhbGhvc3R8cm9sZUEiXSwiVXNlcklkIjoiMTExIiwiRGlzcGxheU5hbWUiOiJVc2VyMSBOYW1lIiwiU3VybmFtZSI6Ik5hbWUiLCJHaXZlbk5hbWUiOiJVc2VyMSJ9._gNI2zMSop__flAeUguOJbbwiP8IDzFMi0VltVvm13o"
	req.Header.Set(authHeader, bearer+token)
	assert.NoError(h(c))

	// valid token via cookie
	req.Header.Del(authHeader)
	cookie = &http.Cookie{
		Name:   "test",
		Domain: "localhost",
		Path:   "/",
		Value:  token,
	}
	req.AddCookie(cookie)
	assert.NoError(h(c))
}
