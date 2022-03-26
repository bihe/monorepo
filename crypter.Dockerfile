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
COPY ./cmd ./cmd
COPY ./go.mod ./
COPY ./go.sum ./
COPY ./internal/crypter  ./internal/crypter
COPY ./pkg ./pkg
COPY ./proto ./proto
RUN GOOS=linux GOARCH=${ARCH} go build -ldflags="-s -w -X main.Version=${TSTAMP} -X main.Build=${COMMIT}" -o crypter.api ./cmd/crypter/server/*.go
## --------------------------------------------------------------------------

## runtime
## --------------------------------------------------------------------------
FROM alpine:latest
LABEL author="henrik@binggl.net"
WORKDIR /opt/crypter
RUN mkdir -p /opt/crypter/etc && mkdir -p /opt/crypter/logs
COPY --from=BACKEND-BUILD /backend-build/crypter.api /opt/crypter
EXPOSE 3000

# Do not run as root user
## alpine specific user/group creation
RUN addgroup -g 1000 -S crypter && \
    adduser -u 1000 -S crypter -G crypter

RUN chown -R crypter:crypter /opt/crypter
USER crypter

CMD ["/opt/crypter/crypter.api","--basepath=/opt/crypter","--port=3000", "--hostname=0.0.0.0"]
