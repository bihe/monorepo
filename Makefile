PROJECTNAME=$(shell basename "$(PWD)")

# Make is verbose in Linux. Make it silent.
MAKEFLAGS += --silent

VERSION="1.0.0-"
COMMIT=`git rev-parse HEAD | cut -c 1-8`
BUILD=`date -u +%Y%m%d.%H%M%S`

GREEN  := $(shell tput -Txterm setaf 2)
YELLOW := $(shell tput -Txterm setaf 3)
WHITE  := $(shell tput -Txterm setaf 7)
CYAN   := $(shell tput -Txterm setaf 6)
RESET  := $(shell tput -Txterm sgr0)


.PHONY: all clean mod-update proto build test coverage dev-frontend compose-dev compose-int

all: help

# ---------------------------------------------------------------------------
# application tasks
# ---------------------------------------------------------------------------

clean: ## clean caches and build output
	@-$(MAKE) -s go-clean

mod-update: ## update to latest compatible packages (yes golang!)
	@-$(MAKE) -s go-update

proto: ## generate protobuf code for grpc
	@-$(MAKE) -s go-protogen

build: ## compile the whole repo
	@-$(MAKE) -s go-build

test: ## unit-test the monorepo
	@-$(MAKE) -s go-test

coverage: ## print coverage results for the monorepo
	@-$(MAKE) -s go-coverage

dev-frontend: ## start the development angular-frontend
	@echo "  >  Starting angular frontend ..."
	cd ./frontend;	yarn install && yarn start -- --public-host https://dev.binggl.net

compose-dev: ## start the microservices for development of frontend
	@echo "  >  Starting docker containers for development..."
	@echo "  >  Remember to set the env-var ARCH. Linux=amd64, MacM1=arm64"
	docker compose -f compose-dev-frontend.yaml rm && docker compose -f compose-dev-frontend.yaml up --build

compose-int: ## start the whole application for integration testing
	@echo "  >  Starting docker containers for integration ..."
	@echo "  >  Remember to set the env-var ARCH. Linux=amd64, MacM1=arm64"
	docker compose -f compose-integration.yaml rm && docker compose -f compose-integration.yaml up --build


# internal tasks

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
	go get -d -u -t ./...
	go mod tidy -compat=1.17

go-proto:
	@echo "  >  Compiline protobuf files ..."
	rm -f ./proto/*pb.go
	## Protobuf
	# https://developers.google.com/protocol-buffers/docs/gotutorial
	# go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	#
	## GRPC
	# https://grpc.io/docs/languages/go/quickstart/
	# go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	protoc --go_out=./proto --go-grpc_out=./proto ./proto/filecrypt.proto

go-build:
	@echo "  >  Building the monorepo ..."
	go build ./...

go-test:
	@echo "  >  Testing the monorepo ..."
	go test -v -race -count=1 ./...

go-coverage:
	@echo "  >  Testing the monorepo (coverage) ..."
	go test -race -coverprofile="coverage.txt" -covermode atomic -count=1 ./...


## Help:
help: ## Show this help.
	@echo ''
	@echo 'Usage:'
	@echo '  ${YELLOW}make${RESET} ${GREEN}<target>${RESET}'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} { \
		if (/^[a-zA-Z_-]+:.*?##.*$$/) {printf "    ${YELLOW}%-20s${GREEN}%s${RESET}\n", $$1, $$2} \
		else if (/^## .*$$/) {printf "  ${CYAN}%s${RESET}\n", substr($$1,4)} \
		}' $(MAKEFILE_LIST)

