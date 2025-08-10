FROM node:lts-slim AS bundler

ARG buildtime_variable_commit=dev
ENV COMMIT=${buildtime_variable_commit}

RUN npm install -g esbuild

# start the bundeling
WORKDIR /bundler
COPY ./assets ./
RUN rm -rf ./bundle
RUN ./bundle.sh auto ${COMMIT}

# cleanup


## runtime
## --------------------------------------------------------------------------
FROM caddy:latest AS frontend

ARG buildtime_variable_username="containeruser"
ARG buildtime_variable_groupname="containergroup"
ARG buildtime_variable_uid="65532"
ARG buildtime_variable_gid="65532"

LABEL author="henrik@binggl.net"
WORKDIR /opt/frontend
RUN mkdir -p /opt/frontend/app && mkdir -p /opt/frontend/static && mkdir -p /opt/frontend/assets
EXPOSE 443

COPY --chown=${buildtime_variable_uid}:${buildtime_variable_gid} --from=bundler /bundler/bundle /opt/frontend/assets/bundle
COPY --chown=${buildtime_variable_uid}:${buildtime_variable_gid} --from=bundler /bundler/*.png /opt/frontend/assets
COPY --chown=${buildtime_variable_uid}:${buildtime_variable_gid} --from=bundler /bundler/*.svg /opt/frontend/assets
COPY --chown=${buildtime_variable_uid}:${buildtime_variable_gid} --from=bundler /bundler/js/tags /opt/frontend/assets/js/tags

# Do not run as root user
## alpine specific user/group creation
RUN addgroup -g ${buildtime_variable_gid} -S ${buildtime_variable_groupname} && \
    adduser -u ${buildtime_variable_uid} -S ${buildtime_variable_username} -G ${buildtime_variable_groupname} -H -h /opt/frontend

RUN chown -R ${buildtime_variable_uid}:${buildtime_variable_gid} /opt/frontend

USER ${buildtime_variable_username}
