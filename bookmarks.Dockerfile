## backend build-phase
## --------------------------------------------------------------------------
FROM golang:alpine AS BACKEND-BUILD

ARG buildtime_variable_version=1.0.0
ARG buildtime_variable_timestamp=20220101
ARG buildtime_variable_commit=dev
ARG buildtime_variable_arch=amd64

ENV VERSION=${buildtime_variable_version}
ENV TSTAMP=${buildtime_variable_timestamp}
ENV COMMIT=${buildtime_variable_commit}
ENV ARCH=${buildtime_variable_arch}

WORKDIR /backend-build
COPY ./go.mod ./
COPY ./go.sum ./
COPY ./internal/bookmarks  ./internal/bookmarks
COPY ./pkg ./pkg

# necessary to build sqlite3
RUN apk add build-base

RUN GOOS=linux GOARCH=${ARCH} go build -ldflags="-s -w -X main.Version=${TSTAMP} -X main.Build=${COMMIT}" -o bookmarks.api ./internal/bookmarks/server.go
## --------------------------------------------------------------------------

## runtime
## --------------------------------------------------------------------------
FROM alpine:latest
LABEL author="henrik@binggl.net"
WORKDIR /opt/bookmarks
RUN mkdir -p /opt/bookmarks/etc && mkdir -p /opt/bookmarks/logs && mkdir -p /opt/bookmarks/uploads && mkdir -p /opt/bookmarks/db
EXPOSE 3000

# Do not run as root user
## alpine specific user/group creation
RUN addgroup -g 1000 -S bookmarks && \
    adduser -u 1000 -S bookmarks -G bookmarks
COPY --chown=1000:1000 --from=BACKEND-BUILD /backend-build/bookmarks.api /opt/bookmarks
RUN chown bookmarks:bookmarks /opt/bookmarks/etc \
    && chown bookmarks:bookmarks /opt/bookmarks/logs \
    && chown bookmarks:bookmarks /opt/bookmarks/uploads \
    && chown bookmarks:bookmarks /opt/bookmarks/db
USER bookmarks

CMD ["/opt/bookmarks/bookmarks.api","--basepath=/opt/bookmarks","--port=3000", "--hostname=0.0.0.0"]
