PROJECTNAME=$(shell basename "$(PWD)")

# Make is verbose in Linux. Make it silent.
MAKEFLAGS += --silent

VERSION="1.0.0-"
COMMIT=`git rev-parse HEAD | cut -c 1-8`
BUILD=`date -u +%Y%m%d.%H%M%S`

all: compile

compile:
	@-$(MAKE) -s go-compile

test:
	@-$(MAKE) -s go-test

run:
	@-$(MAKE) -s go-run

clean:
	@-$(MAKE) go-clean

update:
	@-$(MAKE) go-update

proto:
	@-$(MAKE) go-protogen

coverage:
	@-$(MAKE) -s go-test-coverage


## --------------------------------------------------------------------------

go-compile: go-clean go-protogen go-build

go-clean:
	@echo "  >  Cleaning build cache"
	go clean ./...
	rm -f ./filecrypt.srv

go-test:
	@echo "  >  Go test..."
	go test -race -v ./...

go-build:
	@echo "  >  Building binary..."
	go build -o filecrypt.srv ./cmd/server/*.go

go-run:
	@echo "  >  Run grpc server..."
	go run ./cmd/server/*.go

go-update:
	@echo "  >  Go update dependencies ..."
	go get -u ./...

go-protogen:
	@echo "  >  Compiline protobuf files ..."
	rm -f ./proto/*pb*.go
	# https://developers.google.com/protocol-buffers/docs/gotutorial
	# https://grpc.io/docs/quickstart/go/
	# go get github.com/golang/protobuf/protoc-gen-go
	protoc --proto_path=./proto --go_out=plugins=grpc:./proto filecrypt.proto

go-test-coverage:
	@echo "  >  Go test coverage ..."
	go test -race -coverprofile="coverage.txt" -covermode atomic ./...

.PHONY: compile test clean update proto coverage
