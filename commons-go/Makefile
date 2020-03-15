PROJECTNAME=$(shell basename "$(PWD)")

# Make is verbose in Linux. Make it silent.
MAKEFLAGS += --silent

COMMIT=`git rev-parse HEAD | cut -c 1-8`
BUILD=`date -u +%Y%m%d.%H%M%S`

compile:
	@-$(MAKE) -s go-compile

test:
	@-$(MAKE) -s go-test

clean-test:
	@-$(MAKE) -s go-clean-test

coverage:
	@-$(MAKE) -s go-test-coverage

clean:
	@-$(MAKE) go-clean

go-compile: go-clean go-build

go-test:
	@echo "  >  Go test ..."
	go test -race ./...

go-clean-test:
	@echo "  >  Go test (no cache)..."
	go test -race -count=1 ./...

go-test-coverage:
	@echo "  >  Go test coverage ..."
	go test -race -coverprofile="coverage.txt" -covermode atomic ./...

go-build:
	@echo "  >  Building binary ..."
	go build ./...

go-clean:
	@echo "  >  Cleaning build cache"
	go clean ./...

.PHONY: compile release test run clean coverage
