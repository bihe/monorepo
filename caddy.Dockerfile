## runtime
## --------------------------------------------------------------------------
FROM caddy:latest

ARG buildtime_variable_username="containeruser"
ARG buildtime_variable_groupname="containergroup"
ARG buildtime_variable_uid="65532"
ARG buildtime_variable_gid="65532"

LABEL author="henrik@binggl.net"
WORKDIR /opt/frontend
RUN mkdir -p /opt/frontend/app && mkdir -p /opt/frontend/static
EXPOSE 443

# Do not run as root user
## alpine specific user/group creation
RUN addgroup -g ${buildtime_variable_gid} -S ${buildtime_variable_groupname} && \
    adduser -u ${buildtime_variable_uid} -S ${buildtime_variable_username} -G ${buildtime_variable_groupname} -H -h /opt/frontend

RUN chown -R ${buildtime_variable_uid}:${buildtime_variable_gid} /opt/frontend

USER ${buildtime_variable_username}
