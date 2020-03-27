## backend build-phase
## --------------------------------------------------------------------------
FROM golang:alpine AS BACKEND-BUILD

ARG buildtime_variable_version=1.0.0
ARG buildtime_variable_timestamp=YYYYMMDD
ARG buildtime_variable_commit=githash

ENV VERSION=${buildtime_variable_version}
ENV BUILD=${buildtime_variable_timestamp}
ENV COMMIT=${buildtime_variable_commit}

WORKDIR /backend-build
COPY ./bookmarks ./bookmarks
COPY ./commons-go ./commons-go
WORKDIR /backend-build/bookmarks
RUN GOOS=linux GOARCH=amd64 go build -ldflags="-s -w -X main.Version=${VERSION}-${COMMIT} -X main.Build=${BUILD}" -tags prod -o bookmarks.api ./cmd/server/*.go
## --------------------------------------------------------------------------

## runtime
## --------------------------------------------------------------------------
FROM alpine:latest
LABEL author="henrik@binggl.net"
WORKDIR /opt/bookmarks
RUN mkdir -p /opt/bookmarks/etc && mkdir -p /opt/bookmarks/logs && mkdir -p /opt/bookmarks/templates && mkdir -p /opt/bookmarks/uploads
## required folders assets && templates
COPY --from=BACKEND-BUILD /backend-build/bookmarks/assets /opt/bookmarks/assets
COPY --from=BACKEND-BUILD /backend-build/bookmarks/templates /opt/bookmarks/templates
## the executable
COPY --from=BACKEND-BUILD /backend-build/bookmarks/bookmarks.api /opt/bookmarks
EXPOSE 3000

# Do not run as root user
## alpine specific user/group creation
RUN addgroup -g 1000 -S bookmarks && \
    adduser -u 1000 -S bookmarks -G bookmarks

RUN chown -R bookmarks:bookmarks /opt/bookmarks
USER bookmarks

CMD ["/opt/bookmarks/bookmarks.api","--basepath=/opt/bookmarks","--port=3000", "--hostname=0.0.0.0"]
