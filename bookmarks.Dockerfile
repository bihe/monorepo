## backend build-phase
## --------------------------------------------------------------------------
FROM golang:alpine AS BACKEND-BUILD

ARG buildtime_variable_version=1.0.0
ARG buildtime_variable_timestamp=20220101
ARG buildtime_variable_commit=dev

ENV VERSION=${buildtime_variable_version}
ENV TSTAMP=${buildtime_variable_timestamp}
ENV COMMIT=${buildtime_variable_commit}

WORKDIR /backend-build
COPY ./cmd ./cmd
COPY ./go.mod ./
COPY ./go.sum ./
COPY ./internal/bookmarks  ./internal/bookmarks
COPY ./pkg ./pkg

# necessary to build sqlite3
RUN apk add build-base

RUN GOOS=linux GOARCH=amd64 go build -ldflags="-s -w -X main.Version=${TSTAMP} -X main.Build=${COMMIT}" -o bookmarks.api ./cmd/bookmarks/server/*.go
## --------------------------------------------------------------------------

## runtime
## --------------------------------------------------------------------------
FROM alpine:latest
LABEL author="henrik@binggl.net"
WORKDIR /opt/bookmarks
RUN mkdir -p /opt/bookmarks/etc && mkdir -p /opt/bookmarks/logs && mkdir -p /opt/bookmarks/templates && mkdir -p /opt/bookmarks/uploads && mkdir -p /opt/bookmarks/db
## required folders assets && templates
COPY --from=BACKEND-BUILD /backend-build/internal/bookmarks/assets /opt/bookmarks/assets
COPY --from=BACKEND-BUILD /backend-build/internal/bookmarks/templates /opt/bookmarks/templates
## the executable
COPY --from=BACKEND-BUILD /backend-build/bookmarks.api /opt/bookmarks
EXPOSE 3000

# Do not run as root user
## alpine specific user/group creation
RUN addgroup -g 1000 -S bookmarks && \
    adduser -u 1000 -S bookmarks -G bookmarks

RUN chown -R bookmarks:bookmarks /opt/bookmarks
USER bookmarks

CMD ["/opt/bookmarks/bookmarks.api","--basepath=/opt/bookmarks","--port=3000", "--hostname=0.0.0.0"]
