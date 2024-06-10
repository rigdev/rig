FROM golang:1.22.1-alpine3.19 AS builder

ENV CGO_ENABLED=0

WORKDIR /app

RUN echo "http://dl-cdn.alpinelinux.org/alpine/edge/community" >> /etc/apk/repositories && \
    apk add --no-cache ca-certificates go-task git clang-extra-tools && \
    addgroup -g 1000 nonroot && \
    adduser -u 1000 -G nonroot -D nonroot

COPY go.mod go.sum taskfile.yaml ./
COPY tools tools
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go-task tools:buf \
         tools:protoc-gen-go \
         tools:protoc-gen-connect-go \
         tools:protoc-gen-go-grpc \
         tools:protoc-gen-doc && \
    go mod download # && go mod verify

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    GOFLAGS='-ldflags=-s' go-task proto build-rig

FROM alpine:3

RUN apk add --no-cache curl
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/bin/rig /usr/local/bin/
COPY --from=builder /etc/passwd /etc/passwd

USER 1000

CMD ["rig"]
