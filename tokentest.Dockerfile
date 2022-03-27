## backend build-phase
## --------------------------------------------------------------------------
FROM golang:alpine AS BACKEND-BUILD

ARG buildtime_variable_arch=amd64
ENV ARCH=${buildtime_variable_arch}

WORKDIR /backend-build
COPY ./cmd ./cmd
COPY ./go.mod ./
COPY ./go.sum ./
COPY ./pkg ./pkg
RUN GOOS=linux GOARCH=${ARCH} go build -o tokentest.server ./cmd/login/tokentest/*.go
## --------------------------------------------------------------------------

## runtime
## --------------------------------------------------------------------------
FROM alpine:latest
LABEL author="henrik@binggl.net"
WORKDIR /opt/tokentest
COPY --from=BACKEND-BUILD /backend-build/tokentest.server /opt/tokentest
EXPOSE 3000

# Do not run as root user
## alpine specific user/group creation
RUN addgroup -g 1000 -S tokentest && \
    adduser -u 1000 -S tokentest -G tokentest

RUN chown -R tokentest:tokentest /opt/tokentest
USER tokentest

CMD ["/opt/tokentest/tokentest.server"]
