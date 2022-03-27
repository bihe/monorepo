## backend build-phase
## --------------------------------------------------------------------------
FROM golang:alpine AS BACKEND-BUILD

ARG buildtime_variable_version=2.0.0
ARG buildtime_variable_timestamp=20220101
ARG buildtime_variable_commit=dev
ARG buildtime_variable_runtime=golang
ARG buildtime_variable_arch=amd64

ENV VERSION=${buildtime_variable_version}
ENV TSTAMP=${buildtime_variable_timestamp}
ENV COMMIT=${buildtime_variable_commit}
ENV ARCH=${buildtime_variable_arch}

WORKDIR /backend-build
COPY ./go.mod ./
COPY ./go.sum ./
COPY ./internal/core  ./internal/core
COPY ./internal/crypter  ./internal/crypter
COPY ./pkg ./pkg
COPY ./proto ./proto

# necessary to build sqlite3
RUN apk add build-base

RUN GOOS=linux GOARCH=${ARCH} go build -ldflags="-w -s -X main.Version=${TSTAMP} -X main.Build=${COMMIT}" -o core.api ./internal/core/server.go
## --------------------------------------------------------------------------

## runtime
## --------------------------------------------------------------------------
FROM alpine:latest
LABEL author="henrik@binggl.net"
WORKDIR /opt/core
RUN mkdir -p /opt/core/etc && mkdir -p /opt/core/logs && mkdir -p /opt/core/uploads && mkdir -p /opt/core/db
COPY --from=BACKEND-BUILD /backend-build/core.api /opt/core
EXPOSE 3000

# Do not run as root user
## alpine specific user/group creation
RUN addgroup -g 1000 -S coreapp && \
    adduser -u 1000 -S coreapp -G coreapp

RUN chown -R coreapp:coreapp /opt/core
USER coreapp

CMD ["/opt/core/core.api","--basepath=/opt/core","--port=3000", "--hostname=0.0.0.0"]
