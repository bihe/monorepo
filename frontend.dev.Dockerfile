# the main purpose of this dockerfile is to run angular in dev-mode.
# the path /opt/onefrontend is mounted from the local filesystem - enabling development of the frontend logic
FROM node:lts-alpine
WORKDIR /opt/onefrontend
EXPOSE 4200
# angular is started so that the host is exposed, to be accessed externally, not only localhost
CMD ["yarn", "start", "--", "--host", "0.0.0.0", "--public-host", "https://dev.binggl.net"]
