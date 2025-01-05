## backend build-phase
## --------------------------------------------------------------------------
FROM golang:alpine AS BACKEND-BUILD

ARG buildtime_variable_version="2.0.0"
ARG buildtime_variable_timestamp="20220101"
ARG buildtime_variable_commit="dev"
ARG buildtime_variable_arch="amd64"

ENV VERSION=${buildtime_variable_version}
ENV TSTAMP=${buildtime_variable_timestamp}
ENV COMMIT=${buildtime_variable_commit}
ENV ARCH=${buildtime_variable_arch}

WORKDIR /backend-build
COPY ./go.mod ./
COPY ./go.sum ./
COPY ./cmd/core/server/main.go ./cmd/core/server/main.go
COPY ./internal/core  ./internal/core
COPY ./internal/common  ./internal/common
COPY ./pkg ./pkg
COPY ./assets ./assets

RUN CGO_ENABLED=0 GOOS=linux GOARCH=${ARCH} go build -ldflags="-w -s -X main.Version=${TSTAMP} -X main.Build=${COMMIT}" -o core.api ./cmd/core/server/main.go

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
WORKDIR /opt/core

RUN mkdir -p /opt/core/etc && mkdir -p /opt/core/logs && mkdir -p /opt/core/db

# Do not run as root user
## alpine specific user/group creation
RUN addgroup -g ${buildtime_variable_gid} -S ${buildtime_variable_groupname} && \
    adduser -u ${buildtime_variable_uid} -S ${buildtime_variable_username} -G ${buildtime_variable_groupname} -H -h /opt/core

COPY --chown=${buildtime_variable_uid}:${buildtime_variable_gid} --from=BACKEND-BUILD /backend-build/core.api /opt/core
COPY --chown=${buildtime_variable_uid}:${buildtime_variable_gid} --from=BACKEND-BUILD /backend-build/assets /opt/core/assets

RUN chown ${buildtime_variable_uid}:${buildtime_variable_gid} /opt/core/etc \
    && chown ${buildtime_variable_uid}:${buildtime_variable_gid} /opt/core/logs \
    &&  chown ${buildtime_variable_uid}:${buildtime_variable_gid} /opt/core/db

USER ${buildtime_variable_username}

EXPOSE ${buildtime_variable_port}

CMD [ "/opt/core/core.api", "--basepath=/opt/core", "--hostname=0.0.0.0" ]
