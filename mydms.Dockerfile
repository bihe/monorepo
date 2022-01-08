## backend build-phase
## --------------------------------------------------------------------------
FROM golang:alpine AS BACKEND-BUILD

ARG buildtime_variable_version=3.0.0
ARG buildtime_variable_timestamp=20220101
ARG buildtime_variable_commit=dev

ENV VERSION=${buildtime_variable_version}
ENV TSTAMP=${buildtime_variable_timestamp}
ENV COMMIT=${buildtime_variable_commit}

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

RUN GOOS=linux GOARCH=amd64 go build -ldflags="-w -s -X main.Version=${TSTAMP} -X main.Build=${COMMIT}" -o mydms.api ./cmd/mydms/server/*.go
## --------------------------------------------------------------------------

## runtime
## --------------------------------------------------------------------------
FROM alpine:latest
LABEL author="henrik@binggl.net"
WORKDIR /opt/mydms
RUN mkdir -p /opt/mydms/uploads && mkdir -p /opt/mydms/etc && mkdir -p /opt/mydms/logs && mkdir -p /opt/mydms/db
COPY --from=BACKEND-BUILD /backend-build/mydms.api /opt/mydms
COPY --from=BACKEND-BUILD /backend-build/internal/mydms/assets /opt/mydms/assets

EXPOSE 3000

# Do not run as root user
## alpine specific user/group creation
RUN addgroup -g 1000 -S mydms && \
    adduser -u 1000 -S mydms -G mydms

RUN chown -R mydms:mydms /opt/mydms
USER mydms

CMD ["/opt/mydms/mydms.api","--basepath=/opt/mydms","--port=3000", "--hostname=0.0.0.0"]
