FROM --platform=$BUILDPLATFORM node:21.7.1-alpine3.18 AS builder

WORKDIR /app

COPY docs/package.json docs/package-lock.json ./
RUN npm ci

COPY docs .
RUN npm run build

FROM nginx:1.25.4-alpine3.18

COPY build/package/docs/nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=builder /app/build /usr/share/nginx/html
