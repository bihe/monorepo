## backend build-phase
## --------------------------------------------------------------------------
FROM golang:alpine AS BACKEND-BUILD

ARG buildtime_variable_version=3.0.0
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
COPY ./cmd ./cmd
COPY ./go.mod ./
COPY ./go.sum ./
COPY ./internal/mydms ./internal/mydms
COPY ./pkg ./pkg
COPY ./tools ./tools
RUN go generate ./...

# necessary to build sqlite3
RUN apk add build-base

##
## go build
RUN GOOS=linux GOARCH=${ARCH} go build -ldflags="-w -s -X main.Version=${TSTAMP} -X main.Build=${COMMIT}" -o mydms.api ./cmd/mydms/server/*.go

##
## include litestream into the image and use the litestream replication capabilities
ADD https://github.com/benbjohnson/litestream/releases/download/${LSV} /backend-build/litestream.tar.gz
RUN tar -C /backend-build -xzf /backend-build/litestream.tar.gz

## --------------------------------------------------------------------------

## runtime
## --------------------------------------------------------------------------
FROM alpine:latest
LABEL author="henrik@binggl.net"
WORKDIR /opt/mydms
RUN mkdir -p /opt/litestream && mkdir -p /opt/mydms/uploads && mkdir -p /opt/mydms/etc && mkdir -p /opt/mydms/logs && mkdir -p /opt/mydms/uploads && mkdir -p /opt/mydms/db
EXPOSE 3000

RUN apk add bash

# Do not run as root user
## alpine specific user/group creation
RUN addgroup -g 1000 -S mydms && \
    adduser -u 1000 -S mydms -G mydms

COPY --chown=1000:1000 --from=BACKEND-BUILD /backend-build/mydms.api /opt/mydms
COPY --chown=1000:1000 --from=BACKEND-BUILD /backend-build/internal/mydms/assets /opt/mydms/assets
COPY --chown=1000:1000 --from=BACKEND-BUILD /backend-build/litestream /opt/litestream
COPY --chown=1000:1000 ./litestream/run_litestream.sh /opt/mydms
RUN chown mydms:mydms /opt/mydms/etc \
    && chown mydms:mydms /opt/mydms/logs \
    &&  chown mydms:mydms /opt/mydms/uploads \
    &&  chown mydms:mydms /opt/mydms/db

USER mydms

CMD [ "/opt/mydms/run_litestream.sh" ]
