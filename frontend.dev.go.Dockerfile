# the main purpose of this dockerfile is to enable development of the frontend part (core) and in parallel have the other services available for access
# the path /opt/core is mounted from the local filesystem - enabling development of the frontend logic
FROM golang:alpine
WORKDIR /opt/core

ARG buildtime_variable_arch=amd64
ARG buildtime_variable_version=2.0.0
ARG buildtime_variable_timestamp=20220101
ARG buildtime_variable_commit=dev

ENV VERSION=${buildtime_variable_version}
ENV TSTAMP=${buildtime_variable_timestamp}
ENV COMMIT=${buildtime_variable_commit}
ENV ARCH=${buildtime_variable_arch}

COPY ./go.mod ./
COPY ./go.sum ./
COPY ./cmd/core/server ./cmd/core/server
COPY ./internal/core  ./internal/core
COPY ./internal/crypter  ./internal/crypter
COPY ./pkg ./pkg
COPY ./proto ./proto
COPY ./.air.toml ./
COPY ./tools/air ./

RUN apk add build-base && apk add bash && apk add curl

RUN GOOS=linux GOARCH=${ARCH} go build -ldflags="-w -s -X main.Version=${TSTAMP} -X main.Build=${COMMIT}" -o /opt/core/tmp/app.dev /opt/core/cmd/core/server/main.go

EXPOSE 3000

CMD ["/opt/core/air"]
#CMD ["/opt/core/tmp/app.dev", "--port", "3000", "--basepath", "/opt/core/internal/core"]
