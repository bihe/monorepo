# login-go

Simple API which uses [OIDC](https://developers.google.com/identity/protocols/OpenIDConnect) via Google for authentication and has a simple mariaDB schema to do authorization for my applications.

[![codecov](https://codecov.io/gh/bihe/monorepo/branch/master/graph/badge.svg)](https://codecov.io/gh/bihe/monorepo)

## Technology

* REST backend: [chi](https://github.com/go-chi/chi) (v4.x), golang (1.x)
* mariadb: 10.x

## Build

Use the Makefile and call `make release`

Or the manual step by using go: `go build -o login.api ./cmd/server/*.go`

