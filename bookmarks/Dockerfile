## backend build-phase
## --------------------------------------------------------------------------
FROM golang:alpine AS BACKEND-BUILD

ARG buildtime_variable_version=1.0.0
ARG buildtime_variable_timestamp=20200217
ARG buildtime_variable_commit=b75038e5e9924b67db7bbf3b1147a8e3512b2acb

ENV VERSION=${buildtime_variable_version}
ENV BUILD=${buildtime_variable_timestamp}
ENV COMMIT=${buildtime_variable_commit}

WORKDIR /backend-build
COPY . .
RUN GOOS=linux GOARCH=amd64 go build -ldflags="-s -w -X main.Version=${VERSION}-${COMMIT} -X main.Build=${BUILD}" -tags prod -o bookmarks.api ./cmd/server/*.go
## --------------------------------------------------------------------------

## runtime
## --------------------------------------------------------------------------
FROM alpine:latest
LABEL author="henrik@binggl.net"
WORKDIR /opt/bookmarks
RUN mkdir -p /opt/bookmarks/etc && mkdir -p /opt/bookmarks/logs && mkdir -p /opt/bookmarks/templates && mkdir -p /opt/bookmarks/uploads
## required folders assets && templates
COPY --from=BACKEND-BUILD /backend-build/assets /opt/bookmarks/assets
COPY --from=BACKEND-BUILD /backend-build/templates /opt/bookmarks/templates
## the executable
COPY --from=BACKEND-BUILD /backend-build/bookmarks.api /opt/bookmarks
EXPOSE 3000

# Do not run as root user
## alpine specific user/group creation
RUN addgroup -g 1000 -S bookmarks && \
    adduser -u 1000 -S bookmarks -G bookmarks

RUN chown -R bookmarks:bookmarks /opt/bookmarks
USER bookmarks

CMD ["/opt/bookmarks/bookmarks.api","--c=/opt/bookmarks/etc/application.yaml","--port=3000", "--hostname=0.0.0.0", "--e=Production"]
