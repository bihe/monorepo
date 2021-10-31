## backend build-phase
## --------------------------------------------------------------------------
FROM golang:alpine AS BACKEND-BUILD

ARG buildtime_variable_version=2.0.0
ARG buildtime_variable_timestamp=YYYYMMDD
ARG buildtime_variable_commit=local
ARG buildtime_variable_runtime=golang

ENV VERSION=${buildtime_variable_version}
ENV BUILD=${buildtime_variable_timestamp}
ENV RUNTIME=${buildtime_variable_runtime}
ENV COMMIT=${buildtime_variable_commit}

WORKDIR /backend-build
COPY ./cmd/core/ ./cmd/core/
COPY ./go.mod ./
COPY ./go.sum ./
COPY ./internal/core  ./internal/core
COPY ./internal/crypter  ./internal/crypter
COPY ./pkg ./pkg
COPY ./proto ./proto
RUN GOOS=linux GOARCH=amd64 go build -ldflags="-w -s -X main.Version=${VERSION}-${COMMIT} -X main.Build=${BUILD}" -o core.api ./cmd/core/server/*.go
## --------------------------------------------------------------------------

## runtime
## --------------------------------------------------------------------------
FROM alpine:latest
LABEL author="henrik@binggl.net"
WORKDIR /opt/core
RUN mkdir -p /opt/core/etc && mkdir -p /opt/core/logs && mkdir -p /opt/core/uploads
COPY --from=BACKEND-BUILD /backend-build/core.api /opt/core
EXPOSE 3000

# Do not run as root user
## alpine specific user/group creation
RUN addgroup -g 1000 -S coreapp && \
    adduser -u 1000 -S coreapp -G coreapp

RUN chown -R coreapp:coreapp /opt/core
USER coreapp

CMD ["/opt/core/core.api","--basepath=/opt/core","--port=3000", "--hostname=0.0.0.0"]
