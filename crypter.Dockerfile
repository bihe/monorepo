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
COPY ./cmd ./cmd
COPY ./go.mod ./
COPY ./internal/crypter  ./internal/crypter
COPY ./pkg ./pkg
COPY ./proto ./proto
RUN GOOS=linux GOARCH=amd64 go build -ldflags="-s -w -X main.Version=${VERSION}-${COMMIT} -X main.Build=${BUILD}" -o crypter.api ./cmd/crypter/server/*.go
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
