## fronted build-phase
## --------------------------------------------------------------------------
FROM node:lts-alpine AS FRONTEND-BUILD
WORKDIR /frontend-build
COPY ./web/angular.frontend .
RUN rm -f package-lock.json && yarn global add @angular/cli@latest && yarn install && yarn run build --prod --base-href /ui/
## --------------------------------------------------------------------------

## backend build-phase
## --------------------------------------------------------------------------
FROM golang:alpine AS BACKEND-BUILD

ARG buildtime_variable_version=1.0.0
ARG buildtime_variable_timestamp=20200217
ARG buildtime_variable_commit=b75038e5e9924b67db7bbf3b1147a8e3512b2acb

ENV VERSION=${buildtime_variable_version}
ENV BUILD=${buildtime_variable_timestamp}
ENV COMMIT=${buildtime_variable_commit}

WORKDIR /backend-build
COPY . .
RUN GOOS=linux GOARCH=amd64 go build -ldflags="-s -w -X main.Version=${VERSION}-${COMMIT} -X main.Build=${BUILD}" -tags prod -o onefrontend ./cmd/server/*.go
## --------------------------------------------------------------------------

## runtime
## --------------------------------------------------------------------------
FROM alpine:latest
LABEL author="henrik@binggl.net"
WORKDIR /opt/onefrontend
RUN mkdir -p /opt/onefrontend/etc && mkdir -p /opt/onefrontend/logs && mkdir -p /opt/onefrontend/templates && mkdir -p /opt/onefrontend/web/assets/ui
## required folders assets && templates
COPY --from=BACKEND-BUILD /backend-build/web/assets /opt/onefrontend/web/assets
COPY --from=BACKEND-BUILD /backend-build/templates /opt/onefrontend/templates
## the executable
COPY --from=BACKEND-BUILD /backend-build/onefrontend /opt/onefrontend
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
