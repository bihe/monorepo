PROJECTNAME=$(shell basename "$(PWD)")

# Make is verbose in Linux. Make it silent.
MAKEFLAGS += --silent

VERSION="1.0.0-"
COMMIT=`git rev-parse HEAD | cut -c 1-8`
BUILD=`date -u +%Y%m%d.%H%M%S`

compile:
	@-$(MAKE) -s go-compile

release:
	@-$(MAKE) -s go-compile-release

test:
	@-$(MAKE) -s go-test

run:
	@-$(MAKE) -s go-compile go-run

clean:
	@-$(MAKE) go-clean

swagger:
	@-$(MAKE) -s go-swagger

update:
	@-$(MAKE) go-update

coverage:
	@-$(MAKE) -s go-test-coverage


go-compile: go-clean go-build

go-compile-release: go-clean go-build-release

go-run:
	@echo "  >  Run..."
	./bookmarks.api

go-test:
	@echo "  >  Go test..."
	go test -race -v ./...

go-build:
	@echo "  >  Building binary..."
	go build -o bookmarks.api ./cmd/server/*.go

go-build-release:
	@echo "  >  Building binary..."
	GOOS=linux CGO_ENABLED=0 go build -ldflags="-s -w -X main.Version=${VERSION}${COMMIT} -X main.Build=${BUILD}" -tags prod -o bookmarks.api ./cmd/server/*.go

go-swagger:
	# https://github.com/go-swagger/go-swagger
	swagger generate spec -o assets/swagger/swagger.json -m -w ./internal/server/api

go-clean:
	@echo "  >  Cleaning build cache"
	go clean ./...
	rm -f ./bookmarks.api

go-update:
	@echo "  >  Go update dependencies ..."
	go get -u ./...

go-test-coverage:
	@echo "  >  Go test coverage ..."
	go test -race -coverprofile="coverage.txt" -covermode atomic ./...

.PHONY: compile release test run clean coverage
