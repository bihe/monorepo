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


## set the default architecture should work for most Linux systems
ARCH := amd64
## we use litestream to sync databases, specify the version to use
LITESTREAM_V := v0.3.9

UNAME_M := $(shell uname -m)
ifeq ($(UNAME_M), x86_64)
	ARCH = amd64
endif
ifeq ($(UNAME_M), arm64)
	ARCH = arm64
endif


.PHONY: all clean mod-update proto build test coverage compose-int integration

all: help

# ---------------------------------------------------------------------------
# application tasks
# ---------------------------------------------------------------------------

clean: ## clean caches and build output
	@-$(MAKE) -s go-clean

mod-update: ## update to latest compatible packages (yes golang!)
	@-$(MAKE) -s go-update

build: ## compile the whole repo
	@-$(MAKE) -s go-build

test: ## unit-test the monorepo
	@-$(MAKE) -s go-test

coverage: ## print coverage results for the monorepo
	@-$(MAKE) -s go-coverage

compose-int: ## start the whole application for integration testing
	@echo "  >  Starting docker containers for integration ..."
	@echo "  >  Remember to set the hostname **dev.binggl.net** locally or via DNS"
	ARCH=${ARCH} LSV="${LITESTREAM_V}/litestream-${LITESTREAM_V}-linux-${ARCH}-static.tar.gz" docker compose -f compose-integration.yaml rm && ARCH=${ARCH} LSV="${LITESTREAM_V}/litestream-${LITESTREAM_V}-linux-${ARCH}-static.tar.gz" docker compose -f compose-integration.yaml up --build

integration: ## run the integration test with playwright. NOTE: the compose setup needs to be running
	@echo "  >  Starting integration tests ..."
	python ./testdata/integration/run.py

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
	go mod tidy -compat=1.22

go-build:
	@echo "  >  Building the monorepo ..."
	go build ./...

go-test:
	@echo "  >  Testing the monorepo ..."
	# tparse: https://github.com/mfridman/tparse
	go test -v -race -count=1 -json ./... | tparse -all

go-coverage:
	@echo "  >  Testing the monorepo (coverage) ..."
	# tparse: https://github.com/mfridman/tparse
	go test -race -coverprofile="coverage.txt" -covermode atomic -count=1 -json ./... | tparse -all


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

