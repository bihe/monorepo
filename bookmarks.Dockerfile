## backend build-phase
## --------------------------------------------------------------------------
FROM golang:alpine AS BACKEND-BUILD

ARG buildtime_variable_version=1.0.0
ARG buildtime_variable_timestamp=20220101
ARG buildtime_variable_commit=dev
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
COPY ./internal/bookmarks  ./internal/bookmarks
COPY ./pkg ./pkg

# necessary to build sqlite3
RUN apk add build-base

##
## go build
RUN GOOS=linux GOARCH=${ARCH} go build -ldflags="-s -w -X main.Version=${TSTAMP} -X main.Build=${COMMIT}" -o bookmarks.api ./internal/bookmarks/server.go

##
## include litestream into the image and use the litestream replication capabilities
ADD https://github.com/benbjohnson/litestream/releases/download/${LSV} /backend-build/litestream.tar.gz
RUN tar -C /backend-build -xzf /backend-build/litestream.tar.gz

## --------------------------------------------------------------------------

## runtime
## --------------------------------------------------------------------------
FROM alpine:latest
LABEL author="henrik@binggl.net"
WORKDIR /opt/bookmarks
RUN mkdir -p /opt/litestream && mkdir -p /opt/bookmarks/etc && mkdir -p /opt/bookmarks/logs && mkdir -p /opt/bookmarks/uploads && mkdir -p /opt/bookmarks/db
EXPOSE 3000

RUN apk add bash

# Do not run as root user
## alpine specific user/group creation
RUN addgroup -g 1000 -S bookmarks && \
    adduser -u 1000 -S bookmarks -G bookmarks
COPY --chown=1000:1000 --from=BACKEND-BUILD /backend-build/bookmarks.api /opt/bookmarks
COPY --chown=1000:1000 --from=BACKEND-BUILD /backend-build/litestream /opt/litestream
COPY --chown=1000:1000 ./litestream/run_litestream.sh /opt/bookmarks
RUN chown bookmarks:bookmarks /opt/bookmarks/etc \
    && chown bookmarks:bookmarks /opt/bookmarks/logs \
    && chown bookmarks:bookmarks /opt/bookmarks/uploads \
    && chown bookmarks:bookmarks /opt/bookmarks/db
USER bookmarks

CMD [ "/opt/bookmarks/run_litestream.sh" ]
