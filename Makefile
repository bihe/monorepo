PROJECTNAME=$(shell basename "$(PWD)")

# Make is verbose in Linux. Make it silent.
MAKEFLAGS += --silent

VERSION="1.0.0-"
COMMIT=`git rev-parse HEAD | cut -c 1-8`
BUILD=`date -u +%Y%m%d.%H%M%S`

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

# ---------------------------------------------------------------------------
# login
# ---------------------------------------------------------------------------
__docker-build-login:
	@echo " ... building 'login' docker image"
	docker build -t login -f ./login.Dockerfile .

__docker-run-login:
	@echo " ... running 'login' docker image"
	docker run -it -p 127.0.0.1:3000:3000 -v "$(PWD)/login-go":/opt/login/etc login

# ---------------------------------------------------------------------------
# mydms
# ---------------------------------------------------------------------------
__docker-build-mydms:
	@echo " ... building 'mydms' docker image"
	docker build -t mydms -f ./mydms.Dockerfile .

__docker-run-mydms:
	@echo " ... running 'mydms' docker image"
	docker run -it -p 127.0.0.1:3000:3000 -v "$(PWD)/mydms-go":/opt/mydms/etc mydms

# ---------------------------------------------------------------------------
# bookmarks
# ---------------------------------------------------------------------------
__docker-build-bookmarks:
	@echo " ... building 'bookmarks' docker image"
	docker build -t bookmarks -f ./bookmarks.Dockerfile .

__docker-run-bookmarks:
	@echo " ... running 'bookmarks' docker image"
	docker run -it -p 127.0.0.1:3000:3000 -v "$(PWD)/bookmarks/_etc":/opt/bookmarks/etc bookmarks

# ---------------------------------------------------------------------------
# onefrontend
# ---------------------------------------------------------------------------
__docker-build-onefrontend:
	@echo " ... building 'onefrontend' docker image"
	docker build -t onefrontend -f ./onefrontend.Dockerfile .

__docker-run-onefrontend:
	@echo " ... running 'onefrontend' docker image"
	docker run -it -p 127.0.0.1:3000:3000 -v "$(PWD)/onefrontend/etc":/opt/onefrontend/etc onefrontend

# ---------------------------------------------------------------------------

.PHONY: docker-build-login docker-run-login
