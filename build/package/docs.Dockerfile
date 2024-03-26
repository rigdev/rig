FROM --platform=$BUILDPLATFORM node:21.7.1-alpine3.18 AS builder

WORKDIR /app

COPY package.json package-lock.json ./
RUN npm ci

COPY . .
RUN npm run build

FROM nginx:1.25.4-alpine3.18

COPY --from=builder /app/build /usr/share/nginx/html
