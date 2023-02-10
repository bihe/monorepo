## fronted build-phase
## --------------------------------------------------------------------------
FROM node:lts-alpine AS FRONTEND-BUILD

ARG FRONTEND_MODE=build
ENV FRONTEND_MODE=${FRONTEND_MODE}

WORKDIR /frontend-build
COPY ./frontend .
RUN echo ${FRONTEND_MODE}
RUN rm -f package-lock.json && yarn global add @angular/cli@latest && yarn install && yarn run ${FRONTEND_MODE}
## --------------------------------------------------------------------------


## runtime
## --------------------------------------------------------------------------
FROM caddy:latest
LABEL author="henrik@binggl.net"
WORKDIR /opt/frontend
RUN mkdir -p /opt/frontend/app && mkdir -p /opt/frontend/static
COPY --from=FRONTEND-BUILD /frontend-build/dist/onefrontend-ui /opt/frontend/app
EXPOSE 443

# Do not run as root user
## alpine specific user/group creation
RUN addgroup -g 1000 -S frontend && \
    adduser -u 1000 -S frontend -G frontend
RUN chown -R frontend:frontend /opt/frontend
USER frontend
