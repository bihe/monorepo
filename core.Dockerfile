## backend build-phase
## --------------------------------------------------------------------------
FROM golang:alpine AS BACKEND-BUILD

ARG buildtime_variable_version=2.0.0
ARG buildtime_variable_timestamp=20220101
ARG buildtime_variable_commit=dev
ARG buildtime_variable_runtime=golang
ARG buildtime_variable_arch=amd64
ARG buildtime_variable_litestream_ver=v0.3.9/litestream-v0.3.9-linux-amd64-static.tar.gz

ENV VERSION=${buildtime_variable_version}
ENV TSTAMP=${buildtime_variable_timestamp}
ENV COMMIT=${buildtime_variable_commit}
ENV ARCH=${buildtime_variable_arch}
ENV LSV=${buildtime_variable_litestream_ver}

WORKDIR /backend-build
COPY ./go.mod ./
COPY ./go.sum ./
COPY ./cmd/core/server/main.go ./cmd/core/server/main.go
COPY ./internal/core  ./internal/core
COPY ./internal/crypter  ./internal/crypter
COPY ./pkg ./pkg
COPY ./proto ./proto

# necessary to build sqlite3
RUN apk add build-base

RUN GOOS=linux GOARCH=${ARCH} go build -ldflags="-w -s -X main.Version=${TSTAMP} -X main.Build=${COMMIT}" -o core.api ./cmd/core/server/main.go

##
## include litestream into the image and use the litestream replication capabilities
ADD https://github.com/benbjohnson/litestream/releases/download/${LSV} /backend-build/litestream.tar.gz
RUN tar -C /backend-build -xzf /backend-build/litestream.tar.gz

## --------------------------------------------------------------------------

## runtime
## --------------------------------------------------------------------------
FROM alpine:latest
LABEL author="henrik@binggl.net"
WORKDIR /opt/core
RUN mkdir -p /opt/litestream && mkdir -p /opt/core/etc && mkdir -p /opt/core/logs && mkdir -p /opt/core/uploads && mkdir -p /opt/core/db
EXPOSE 3000

RUN apk add bash

# Do not run as root user
## alpine specific user/group creation
RUN addgroup -g 1000 -S coreapp && \
    adduser -u 1000 -S coreapp -G coreapp

COPY --chown=1000:1000 --from=BACKEND-BUILD /backend-build/core.api /opt/core
COPY --chown=1000:1000 --from=BACKEND-BUILD /backend-build/litestream /opt/litestream
COPY --chown=1000:1000 ./litestream/run_litestream.sh /opt/core
RUN chown coreapp:coreapp /opt/core/etc \
    && chown coreapp:coreapp /opt/core/logs \
    &&  chown coreapp:coreapp /opt/core/uploads \
    &&  chown coreapp:coreapp /opt/core/db
USER coreapp

CMD [ "/opt/core/run_litestream.sh" ]
