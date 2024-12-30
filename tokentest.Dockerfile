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
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${ARCH} go build -o tokentest.server ./cmd/login/tokentest/*.go
## --------------------------------------------------------------------------

## runtime
## --------------------------------------------------------------------------
FROM gcr.io/distroless/static-debian12:nonroot

LABEL author="henrik@binggl.net"
WORKDIR /opt/tokentest
COPY --from=BACKEND-BUILD --chown=nonroot:nonroot /backend-build/tokentest.server /opt/tokentest

EXPOSE 3000

CMD ["/opt/tokentest/tokentest.server"]
