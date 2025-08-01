## backend build-phase
## --------------------------------------------------------------------------
FROM golang:alpine AS backend-build

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
COPY ./cmd/bookmarks/server ./cmd/bookmarks/server
COPY ./internal/bookmarks  ./internal/bookmarks
COPY ./internal/common  ./internal/common
COPY ./pkg ./pkg
COPY ./assets ./assets

##
## go build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${ARCH} go build -ldflags="-s -w -X main.Version=${TSTAMP} -X main.Build=${COMMIT}" -o bookmarks.api ./cmd/bookmarks/server/main.go

## --------------------------------------------------------------------------

## runtime
## --------------------------------------------------------------------------
FROM alpine:3

ARG buildtime_variable_username="containeruser"
ARG buildtime_variable_groupname="containergroup"
ARG buildtime_variable_uid="65532"
ARG buildtime_variable_gid="65532"
ARG buildtime_variable_port="3000"

LABEL author="henrik@binggl.net"
WORKDIR /opt/bookmarks
RUN mkdir -p /opt/bookmarks/etc && mkdir -p /opt/bookmarks/logs && mkdir -p /opt/bookmarks/uploads && mkdir -p /opt/bookmarks/db && mkdir -p /opt/bookmarks/assets

# Do not run as root user
## alpine specific user/group creation
RUN addgroup -g ${buildtime_variable_gid} -S ${buildtime_variable_groupname} && \
    adduser -u ${buildtime_variable_uid} -S ${buildtime_variable_username} -G ${buildtime_variable_groupname} -H -h /opt/bookmarks

COPY --chown=${buildtime_variable_uid}:${buildtime_variable_gid} --from=backend-build /backend-build/bookmarks.api /opt/bookmarks
COPY --chown=${buildtime_variable_uid}:${buildtime_variable_gid} --from=backend-build /backend-build/assets /opt/bookmarks/assets

RUN chown ${buildtime_variable_uid}:${buildtime_variable_gid} /opt/bookmarks/etc \
    && chown ${buildtime_variable_uid}:${buildtime_variable_gid} /opt/bookmarks/logs \
    && chown ${buildtime_variable_uid}:${buildtime_variable_gid} /opt/bookmarks/uploads \
    && chown ${buildtime_variable_uid}:${buildtime_variable_gid} /opt/bookmarks/db

USER ${buildtime_variable_username}

EXPOSE ${buildtime_variable_port}

CMD [ "/opt/bookmarks/bookmarks.api", "--basepath=/opt/bookmarks", "--hostname=0.0.0.0" ]
