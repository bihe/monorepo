## fronted build-phase
## --------------------------------------------------------------------------
FROM node:lts-alpine AS FRONTEND-BUILD

ARG FRONTEND_MODE=prod
ENV FRONTEND_MODE=${FRONTEND_MODE}

WORKDIR /frontend-build
COPY ./onefrontend/web/angular.frontend .
RUN echo ${FRONTEND_MODE}
RUN rm -f package-lock.json && yarn global add @angular/cli@latest && yarn install && yarn run ${FRONTEND_MODE} --base-href /ui/
## --------------------------------------------------------------------------

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
COPY ./onefrontend ./onefrontend
COPY ./commons-go ./commons-go
WORKDIR /backend-build/onefrontend
RUN GOOS=linux GOARCH=amd64 go build -ldflags="-s -w -X main.Version=${VERSION}-${COMMIT} -X main.Build=${BUILD}" -tags prod -o onefrontend ./cmd/server/*.go
## --------------------------------------------------------------------------

## runtime
## --------------------------------------------------------------------------
FROM alpine:latest
LABEL author="henrik@binggl.net"
WORKDIR /opt/onefrontend
RUN mkdir -p /opt/onefrontend/etc && mkdir -p /opt/onefrontend/logs && mkdir -p /opt/onefrontend/templates && mkdir -p /opt/onefrontend/web/assets/ui
## required folders assets && templates
COPY --from=BACKEND-BUILD /backend-build/onefrontend/web/assets /opt/onefrontend/web/assets
COPY --from=BACKEND-BUILD /backend-build/onefrontend/templates /opt/onefrontend/templates
## the executable
COPY --from=BACKEND-BUILD /backend-build/onefrontend/onefrontend /opt/onefrontend
## the SPA frontend
COPY --from=FRONTEND-BUILD /frontend-build/dist/onefrontend-ui /opt/onefrontend/web/ui

EXPOSE 3000

# Do not run as root user
## alpine specific user/group creation
RUN addgroup -g 1000 -S onefrontend && \
    adduser -u 1000 -S onefrontend -G onefrontend

RUN chown -R onefrontend:onefrontend /opt/onefrontend
USER onefrontend

CMD ["/opt/onefrontend/onefrontend","--c=/opt/onefrontend/etc/application.yaml","--port=3000", "--hostname=0.0.0.0", "--e=Production"]
