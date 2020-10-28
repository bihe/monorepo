PROJECTNAME=$(shell basename "$(PWD)")

# Make is verbose in Linux. Make it silent.
MAKEFLAGS += --silent

VERSION="1.0.0-"
COMMIT=`git rev-parse HEAD | cut -c 1-8`
BUILD=`date -u +%Y%m%d.%H%M%S`

# ---------------------------------------------------------------------------
# application tasks
# ---------------------------------------------------------------------------

## common
## --------------------------------------------------------------------------
clean:
	@-$(MAKE) -s go-clean

mod-update:
	@-$(MAKE) -s go-update

protogen:
	@-$(MAKE) -s go-protogen

swagger:
	@-$(MAKE) -s go-swagger

test:
	@-$(MAKE) -s go-test

coverage:
	@-$(MAKE) -s go-coverage


## onefrontend
## --------------------------------------------------------------------------
onefrontend-build:
	@-$(MAKE) -s onefrontend_go_build

onefrontend-release:
	@-$(MAKE) -s onefrontend_go_build-release

onefrontend-ui:
	@-$(MAKE) -s onefrontend_do_angular_build

## crypter
## --------------------------------------------------------------------------
crypter-build:
	@-$(MAKE) -s crypter_go_build

crypter-release:
	@-$(MAKE) -s crypter_go_build-release

## login
## --------------------------------------------------------------------------
login-build:
	@-$(MAKE) -s login_go_build

login-release:
	@-$(MAKE) -s login_go_build-release

## bookmarks
## --------------------------------------------------------------------------
bookmarks-build:
	@-$(MAKE) -s bookmarks_go_build

bookmarks-release:
	@-$(MAKE) -s bookmarks_go_build-release

## mydms
## --------------------------------------------------------------------------
mydms-build:
	@-$(MAKE) -s mydms_go_build

mydms-release:
	@-$(MAKE) -s mydms_go_build-release


# ---------------------------------------------------------------------------
# docker tasks
# ---------------------------------------------------------------------------

docker-build-login:
	@-$(MAKE) -s __docker-build-login

docker-run-login:
	@-$(MAKE) -s __docker-run-login

docker-build-mydms:
	@-$(MAKE) -s __docker-build-mydms

docker-run-mydms:
	@-$(MAKE) -s __docker-run-mydms

docker-build-bookmarks:
	@-$(MAKE) -s __docker-build-bookmarks

docker-run-bookmarks:
	@-$(MAKE) -s __docker-run-bookmarks

docker-build-onefrontend:
	@-$(MAKE) -s __docker-build-onefrontend

docker-run-onefrontend:
	@-$(MAKE) -s __docker-run-onefrontend

docker-build-crypter:
	@-$(MAKE) -s __docker-build-crypter

docker-run-crypter:
	@-$(MAKE) -s __docker-run-crypter


## --------------------------------------------------------------------------
## common tasks
## --------------------------------------------------------------------------
go-clean:
	@echo "  >  Cleaning build cache"
	go clean ./...
	rm -f ./dist/onefrontend.api
	rm -f ./dist/crypter.api
	rm -f ./dist/login.api
	rm -f ./dist/mydms.api
	rm -f ./dist/bookmarks.api

go-update:
	@echo "  >  Go update dependencies ..."
	go get -u ./...
	go mod tidy

go-protogen:
	@echo "  >  Compiline protobuf files ..."
	rm -f ./proto/*pb*.go
	# https://developers.google.com/protocol-buffers/docs/gotutorial
	# https://grpc.io/docs/quickstart/go/
	# go get github.com/golang/protobuf/protoc-gen-go
	protoc --proto_path=./proto --go_out=plugins=grpc:./proto filecrypt.proto

go-swagger:
	# https://github.com/go-swagger/go-swagger
	./tools/swagger_linux_amd64 generate spec -o login/web/assets/swagger/swagger.json -m -w ./login/api
	./tools/swagger_linux_amd64 generate spec -o bookmarks/assets/swagger/swagger.json -m -w ./bookmarks/server/api

go-test:
	@echo "  >  Testing the monorepo ..."
	go test -race -count=1 ./...

go-coverage:
	@echo "  >  Testing the monorepo (coverage) ..."
	go test -race -coverprofile="coverage.txt" -covermode atomic -count=1 ./...

# ---------------------------------------------------------------------------
# login
# ---------------------------------------------------------------------------

login_go_build:
	@echo "  >  Building login ..."
	go build -o ./dist/login.api ./cmd/login/server/*.go

login_go_build-release:
	@echo "  >  Building login ..."
	GOOS=linux CGO_ENABLED=0 go build -ldflags="-s -w -X main.Version=${VERSION}${COMMIT} -X main.Build=${BUILD}" -o ./dist/login.api ./cmd/login/server/*.go

__docker-build-login:
	@echo " ... building 'login' docker image"
	docker build -t login -f ./login.Dockerfile .

__docker-run-login:
	@echo " ... running 'login' docker image"
	docker run -it -p 127.0.0.1:3000:3000 -v "$(PWD)/login-go":/opt/login/etc login

# ---------------------------------------------------------------------------
# mydms
# ---------------------------------------------------------------------------

mydms_go_build:
	@echo "  >  Building mydms ..."
	go build -o ./dist/mydms.api ./cmd/mydms/server/*.go

mydms_go_build-release:
	@echo "  >  Building mydms ..."
	GOOS=linux CGO_ENABLED=0 go build -ldflags="-s -w -X main.Version=${VERSION}${COMMIT} -X main.Build=${BUILD}" -o ./dist/mydms.api ./cmd/mydms/solid/*.go

__docker-build-mydms:
	@echo " ... building 'mydms' docker image"
	docker build -t mydms -f ./mydms.Dockerfile .

__docker-run-mydms:
	@echo " ... running 'mydms' docker image"
	docker run -it -p 127.0.0.1:3000:3000 --env-file=$(PWD)/internal/mydms/.env -v "$(PWD)/internal/mydms":/opt/mydms/etc mydms

# ---------------------------------------------------------------------------
# bookmarks
# ---------------------------------------------------------------------------

bookmarks_go_build:
	@echo "  >  Building bookmarks ..."
	go build -o ./dist/bookmarks.api ./cmd/bookmarks/server/*.go

bookmarks_go_build-release:
	@echo "  >  Building bookmarks ..."
	GOOS=linux CGO_ENABLED=0 go build -ldflags="-s -w -X main.Version=${VERSION}${COMMIT} -X main.Build=${BUILD}" -o ./dist/bookmarks.api ./cmd/bookmarks/server/*.go


__docker-build-bookmarks:
	@echo " ... building 'bookmarks' docker image"
	docker build -t bookmarks -f ./bookmarks.Dockerfile .

__docker-run-bookmarks:
	@echo " ... running 'bookmarks' docker image"
	docker run -it -p 127.0.0.1:3000:3000 -v "$(PWD)/bookmarks/_etc":/opt/bookmarks/etc bookmarks

# ---------------------------------------------------------------------------
# onefrontend
# ---------------------------------------------------------------------------

onefrontend_go_build:
	@echo "  >  Building onefrontend ..."
	go build -o ./dist/onefrontend.api ./cmd/onefrontend/server/*.go

onefrontend_go_build-release:
	@echo "  >  Building onefrontend ..."
	GOOS=linux CGO_ENABLED=0 go build -ldflags="-s -w -X main.Version=${VERSION}${COMMIT} -X main.Build=${BUILD}" -o ./dist/onefrontend.api ./cmd/onefrontend/server/*.go

onefrontend_do_angular_build:
	@echo "  >  Building angular frontend ..."
	cd ./onefrontend/web/angular.frontend;	npm install && npm run build -- --prod --base-href /ui/

__docker-build-onefrontend:
	@echo " ... building 'onefrontend' docker image"
	docker build -t onefrontend -f ./onefrontend.Dockerfile .

__docker-run-onefrontend:
	@echo " ... running 'onefrontend' docker image"
	docker run -it -p 127.0.0.1:3000:3000 -v "$(PWD)/onefrontend/etc":/opt/onefrontend/etc onefrontend

# ---------------------------------------------------------------------------
# crypter
# ---------------------------------------------------------------------------

crypter_go_build:
	@echo "  >  Building crypter ..."
	go build -o ./dist/crypter.api ./cmd/crypter/server/*.go

crypter_go_build-release:
	@echo "  >  Building crypter ..."
	GOOS=linux CGO_ENABLED=0 go build -ldflags="-s -w -X main.Version=${VERSION}${COMMIT} -X main.Build=${BUILD}" -o ./dist/crypter.api ./cmd/crypter/server/*.go

__docker-build-crypter:
	@echo " ... building 'crypter' docker image"
	docker build -t crypter -f ./crypter.Dockerfile .

__docker-run-crypter:
	@echo " ... running 'crypter' docker image"
	docker run -it -p 127.0.0.1:3001:3000 -v "$(PWD)/crypter/":/opt/crypter/etc crypter


# ---------------------------------------------------------------------------

.PHONY: docker-build-login docker-run-login
