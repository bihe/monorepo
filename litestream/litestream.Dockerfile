## build-phase
## --------------------------------------------------------------------------
FROM alpine:3 AS build

ARG buildtime_variable_arch=amd64
ARG buildtime_variable_litestream_ver=v0.3.13/litestream-v0.3.13-linux-amd64.tar.gz

ENV ARCH=${buildtime_variable_arch}
ENV LSV=${buildtime_variable_litestream_ver}

WORKDIR /build

##
## download the litestream binary
ADD https://github.com/benbjohnson/litestream/releases/download/${LSV} /build/litestream.tar.gz
RUN tar -C /build -xzf /build/litestream.tar.gz

RUN mkdir -p /build/opt/litestream && mkdir -p /build/opt/litestream/store

## --------------------------------------------------------------------------

## runtime
## --------------------------------------------------------------------------
FROM alpine:3

ARG buildtime_variable_username="containeruser"
ARG buildtime_variable_groupname="containergroup"
ARG buildtime_variable_uid="65532"
ARG buildtime_variable_gid="65532"


LABEL author="henrik@binggl.net"
WORKDIR /opt/litestream

RUN apk add bash

# Do not run as root user
## alpine specific user/group creation
RUN addgroup -g ${buildtime_variable_gid} -S ${buildtime_variable_groupname} && \
    adduser -u ${buildtime_variable_uid} -S ${buildtime_variable_username} -G ${buildtime_variable_groupname} -H -h /opt/litestream


COPY --chown=${buildtime_variable_uid}:${buildtime_variable_gid} --from=build /build/opt/litestream/ /opt/litestream
COPY --chown=${buildtime_variable_uid}:${buildtime_variable_gid} --from=build /build/opt/litestream/store /opt/litestream/store
COPY --chown=${buildtime_variable_uid}:${buildtime_variable_gid} --from=build /build/litestream /opt/litestream/litestream
COPY --chown=${buildtime_variable_uid}:${buildtime_variable_gid} ./restore_replicate.sh /opt/litestream
COPY --chown=${buildtime_variable_uid}:${buildtime_variable_gid} ./healthcheck.sh /opt/litestream

USER ${buildtime_variable_username}

CMD ["/opt/litestream/restore_replicate.sh"]
