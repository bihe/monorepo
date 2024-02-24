## runtime
## --------------------------------------------------------------------------
FROM caddy:latest
LABEL author="henrik@binggl.net"
WORKDIR /opt/frontend
RUN mkdir -p /opt/frontend/app && mkdir -p /opt/frontend/static
EXPOSE 443

# Do not run as root user
## alpine specific user/group creation
RUN addgroup -g 1000 -S frontend && \
    adduser -u 1000 -S frontend -G frontend
USER frontend
