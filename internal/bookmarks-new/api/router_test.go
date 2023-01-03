package api_test

import (
	"net/http"

	"golang.binggl.net/monorepo/internal/bookmarks-new/api"
	"golang.binggl.net/monorepo/internal/bookmarks-new/app/bookmarks"
	"golang.binggl.net/monorepo/internal/bookmarks-new/app/conf"
	"golang.binggl.net/monorepo/pkg/config"
	"golang.binggl.net/monorepo/pkg/logging"
)

const validToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE4MjcyMDE0NDUsImlhdCI6MTYyNjU5NjY0NSwiaXNzIjoiaXNzdWVyIiwianRpIjoiZmI1ZThjNDMtZDEwMi00MGU3LTljM2EtNTkwYzA3ODAzNzUwIiwic3ViIjoiaGVucmlrLmJpbmdnbEBnbWFpbC5jb20iLCJDbGFpbXMiOlsiQXxodHRwOi8vQXxBIl0sIkRpc3BsYXlOYW1lIjoiVXNlciBBIiwiRW1haWwiOiJ1c2VyQGEuY29tIiwiR2l2ZW5OYW1lIjoidXNlciIsIlN1cm5hbWUiOiJBIiwiVHlwZSI6ImxvZ2luLlVzZXIiLCJVc2VySWQiOiIxMiIsIlVzZXJOYW1lIjoidXNlckBhLmNvbSJ9.JcZ9-ImQieOWW1KaLJGR_Pqol2MQviFDdjqbIfhAlek"

// --------------------------------------------------------------------------
//   Boilerplate to make the tests happen
// --------------------------------------------------------------------------

var logger = logging.NewNop()

type handlerOps struct {
	version string
	build   string
	app     *bookmarks.Application
}

func handlerWith(ops *handlerOps) http.Handler {
	return api.MakeHTTPHandler(ops.app, logger, api.HTTPHandlerOptions{
		BasePath:  "./",
		ErrorPath: "/error",
		Config: conf.AppConfig{
			BaseConfig: config.BaseConfig{
				Cookies: config.ApplicationCookies{
					Prefix: "core",
				},
				Security: config.Security{
					CookieName: "jwt",
					JwtIssuer:  "issuer",
					JwtSecret:  "secret",
					Claim: config.Claim{
						Name:  "A",
						URL:   "http://A",
						Roles: []string{"A"},
					},
				},
				Logging: config.LogConfig{},
				Cors:    config.CorsSettings{},
				Assets: config.AssetSettings{
					AssetDir:    "./",
					AssetPrefix: "/static",
				},
			},
		},
		Version: ops.version,
		Build:   ops.build,
	})
}
