## backend build-phase
## --------------------------------------------------------------------------
FROM golang:alpine AS BACKEND-BUILD

ARG buildtime_variable_version=2.0.0
ARG buildtime_variable_timestamp=YYYYMMDD
ARG buildtime_variable_commit=b75038e5e9924b67db7bbf3b1147a8e3512b2acb

ENV VERSION=${buildtime_variable_version}
ENV BUILD=${buildtime_variable_timestamp}
ENV COMMIT=${buildtime_variable_commit}

WORKDIR /backend-build
COPY . .
RUN GOOS=linux GOARCH=amd64 go build -ldflags="-w -s -X main.Version=${VERSION}-${COMMIT} -X main.Build=${BUILD}" -tags prod -o mydms.api
#COPY --from=FRONTEND-BUILD /frontend-build/dist  ./ui
## --------------------------------------------------------------------------

## runtime
## --------------------------------------------------------------------------
FROM alpine:latest
LABEL author="henrik@binggl.net"
WORKDIR /opt/mydms
RUN mkdir -p /opt/mydms/uploads && mkdir -p /opt/mydms/etc && mkdir -p /opt/mydms/logs
COPY --from=BACKEND-BUILD /backend-build/mydms.api /opt/mydms
RUN ls -l /opt/mydms
RUN ls -l /opt/mydms/etc

EXPOSE 3000

CMD ["/opt/mydms/mydms.api","--c=/opt/mydms/etc/application.json","--port=3000", "--hostname=0.0.0.0"]
