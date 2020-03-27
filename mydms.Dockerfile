## backend build-phase
## --------------------------------------------------------------------------
FROM golang:alpine AS BACKEND-BUILD

ARG buildtime_variable_version=2.0.0
ARG buildtime_variable_timestamp=YYYYMMDD
ARG buildtime_variable_commit=githash

ENV VERSION=${buildtime_variable_version}
ENV BUILD=${buildtime_variable_timestamp}
ENV COMMIT=${buildtime_variable_commit}

WORKDIR /backend-build
COPY ./mydms-go ./mydms-go
COPY ./commons-go ./commons-go
WORKDIR /backend-build/mydms-go
RUN GOOS=linux GOARCH=amd64 go build -ldflags="-w -s -X main.Version=${VERSION}-${COMMIT} -X main.Build=${BUILD}" -tags prod -o mydms.api
## --------------------------------------------------------------------------

## runtime
## --------------------------------------------------------------------------
FROM alpine:latest
LABEL author="henrik@binggl.net"
WORKDIR /opt/mydms
RUN mkdir -p /opt/mydms/uploads && mkdir -p /opt/mydms/etc && mkdir -p /opt/mydms/logs
COPY --from=BACKEND-BUILD /backend-build/mydms-go/mydms.api /opt/mydms

EXPOSE 3000

# Do not run as root user
## alpine specific user/group creation
RUN addgroup -g 1000 -S mydms && \
    adduser -u 1000 -S mydms -G mydms

RUN chown -R mydms:mydms /opt/mydms
USER mydms

CMD ["/opt/mydms/mydms.api","--basepath=/opt/mydms","--port=3000", "--hostname=0.0.0.0"]
