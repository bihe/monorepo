## backend build-phase
## --------------------------------------------------------------------------
FROM golang:alpine AS BACKEND-BUILD

ARG buildtime_variable_version=3.0.0
ARG buildtime_variable_timestamp=20220101
ARG buildtime_variable_commit=dev
ARG buildtime_variable_arch=amd64

ENV VERSION=${buildtime_variable_version}
ENV TSTAMP=${buildtime_variable_timestamp}
ENV COMMIT=${buildtime_variable_commit}
ENV ARCH=${buildtime_variable_arch}

WORKDIR /backend-build
COPY ./cmd ./cmd
COPY ./go.mod ./
COPY ./go.sum ./
COPY ./internal/mydms ./internal/mydms
COPY ./internal/common  ./internal/common
COPY ./pkg ./pkg
COPY ./assets ./assets

##
## go build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${ARCH} go build -ldflags="-w -s -X main.Version=${TSTAMP} -X main.Build=${COMMIT}" -o mydms.api ./cmd/mydms/server/*.go

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
WORKDIR /opt/mydms
RUN mkdir -p /opt/mydms/uploads && mkdir -p /opt/mydms/etc && mkdir -p /opt/mydms/logs && mkdir -p /opt/mydms/db && mkdir -p /opt/mydms/assets

# Do not run as root user
## alpine specific user/group creation
RUN addgroup -g ${buildtime_variable_gid} -S ${buildtime_variable_groupname} && \
    adduser -u ${buildtime_variable_uid} -S ${buildtime_variable_username} -G ${buildtime_variable_groupname} -H -h /opt/mydms

COPY --chown=${buildtime_variable_uid}:${buildtime_variable_gid} --from=BACKEND-BUILD /backend-build/mydms.api /opt/mydms
COPY --chown=${buildtime_variable_uid}:${buildtime_variable_gid} --from=BACKEND-BUILD /backend-build/assets /opt/mydms/assets

RUN chown ${buildtime_variable_uid}:${buildtime_variable_gid} /opt/mydms/etc \
    && chown ${buildtime_variable_uid}:${buildtime_variable_gid} /opt/mydms/logs \
    &&  chown ${buildtime_variable_uid}:${buildtime_variable_gid} /opt/mydms/uploads \
    &&  chown ${buildtime_variable_uid}:${buildtime_variable_gid} /opt/mydms/db

USER ${buildtime_variable_username}

EXPOSE ${buildtime_variable_port}

CMD [ "/opt/mydms/mydms.api", "--basepath=/opt/mydms", "--hostname=0.0.0.0" ]
