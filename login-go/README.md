# login-go
Simple API which uses [OIDC](https://developers.google.com/identity/protocols/OpenIDConnect) via Google for authentication and has a really, really simple mariaDB to do authorization for my applications.

[![codecov](https://codecov.io/gh/bihe/login-go/branch/master/graph/badge.svg)](https://codecov.io/gh/bihe/login-go)
[![Build Status](https://dev.azure.com/henrikbinggl/login-go/_apis/build/status/bihe.login-go?branchName=master)](https://dev.azure.com/henrikbinggl/login-go/_build/latest?definitionId=7&branchName=master)

## Technology

* REST backend: [chi](https://github.com/go-chi/chi) (v4.0.x), golang (1.1x)
* frontend angular (8.x.x)
* mariadb: 10.x

## Build

The REST Api and the UI can be built separately.

### UI

`npm run build -- --prod --base-href /ui/`

### Api

Use the Makefile and call `make release` or create a docker image by `make docker-build`

Or the manual step by using go: `go build`

