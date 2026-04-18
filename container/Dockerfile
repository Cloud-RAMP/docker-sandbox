# docker run -v /tmp/cloud_ramp_socket:/tmp/cloud_ramp_socket container-sandbox

FROM node:18-alpine

WORKDIR /app

COPY . .

ENTRYPOINT ["node", "index.js"]