## backend build-phase
## --------------------------------------------------------------------------
FROM golang:alpine AS BACKEND-BUILD

ARG buildtime_variable_version=2.0.0
ARG buildtime_variable_timestamp=YYYYMMDD
ARG buildtime_variable_commit=githash
ARG buildtime_variable_runtime=golang

ENV VERSION=${buildtime_variable_version}
ENV BUILD=${buildtime_variable_timestamp}
ENV RUNTIME=${buildtime_variable_runtime}
ENV COMMIT=${buildtime_variable_commit}

WORKDIR /backend-build
COPY ./cmd ./cmd
COPY ./go.mod ./
COPY ./internal/login  ./internal/login
COPY ./pkg ./pkg
RUN GOOS=linux GOARCH=amd64 go build -ldflags="-w -s -X main.Version=${VERSION}-${COMMIT} -X main.Build=${BUILD}" -o login.api ./cmd/login/server/*.go
## --------------------------------------------------------------------------

## runtime
## --------------------------------------------------------------------------
FROM alpine:latest
LABEL author="henrik@binggl.net"
WORKDIR /opt/login
RUN mkdir -p /opt/login/etc && mkdir -p /opt/login/logs && mkdir -p /opt/login/templates && mkdir -p /opt/login/web
## required folders assets && templates
COPY --from=BACKEND-BUILD /backend-build/internal/login/web /opt/login/web
COPY --from=BACKEND-BUILD /backend-build/internal/login/templates /opt/login/templates
## the executable
COPY --from=BACKEND-BUILD /backend-build/login.api /opt/login

EXPOSE 3000

# Do not run as root user
## alpine specific user/group creation
RUN addgroup -g 1000 -S loginapp && \
    adduser -u 1000 -S loginapp -G loginapp

RUN chown -R loginapp:loginapp /opt/login
USER loginapp

CMD ["/opt/login/login.api","--basepath=/opt/login","--port=3000", "--hostname=0.0.0.0"]
