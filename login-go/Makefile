PROJECTNAME=$(shell basename "$(PWD)")

# Make is verbose in Linux. Make it silent.
MAKEFLAGS += --silent

VERSION="2.0.0-"
COMMIT=`git rev-parse HEAD | cut -c 1-8`
BUILD=`date -u +%Y%m%d.%H%M%S`
RUNTIME=`go version | sed 's/.*version //' | sed 's/ .*//'`

compile: ## compile the application
	@-$(MAKE) -s go-compile

clean: ## clean
	@-$(MAKE) go-clean

update:
	@-$(MAKE) go-update

release:
	@-$(MAKE) -s go-compile-release

test:
	@-$(MAKE) -s go-test

clean-test:
	@-$(MAKE) -s go-clean-test

coverage:
	@-$(MAKE) -s go-test-coverage

swagger:
	@-$(MAKE) -s go-swagger

run:
	@-$(MAKE) -s go-compile go-run

docker-build:
	@-$(MAKE) -s __docker-build

docker-run:
	@-$(MAKE) -s __docker-run


go-compile: go-clean go-build

go-compile-release: go-clean go-build-release

go-run:
	@echo "  >  Running application ..."
	./login.api

go-test:
	@echo "  >  Go test ..."
	go test -race ./...

go-update:
	@echo "  >  Go update dependencies ..."
	go get -u ./...

go-clean-test:
	@echo "  >  Go test (no cache)..."
	go test -race -count=1 ./...

go-test-coverage:
	@echo "  >  Go test coverage ..."
	go test -race -coverprofile="coverage.txt" -covermode atomic ./...

go-build:
	@echo "  >  Building binary ..."
	go build -o login.api ./cmd/server/*.go

go-build-release:
	@echo "  >  Building binary..."
	GOOS=linux GOARCH=amd64 go build -ldflags="-w -s -X main.Version=${VERSION}${COMMIT} -X main.Build=${BUILD}" -tags prod -o login.api cmd/server/*.go

go-swagger:
	# https://github.com/go-swagger/go-swagger
	./tools/swagger_linux_amd64 generate spec -o web/assets/swagger/swagger.json -m -w ./internal/api

go-clean:
	@echo "  >  Cleaning build cache"
	go clean ./...
	rm -f ./login.api

__docker-build:
	@echo " ... building docker image"
	docker build -t login .

__docker-run:
	@echo " ... running docker image"
	docker run -it -p 127.0.0.1:3000:3000 -v "$(PWD)":/opt/login/etc login

.PHONY: compile release test run clean coverage

