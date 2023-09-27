FROM golang:1.21.1-alpine3.18 AS builder

ENV CGO_ENABLED=0

WORKDIR /app

RUN apk add --no-cache ca-certificates make git && \
    addgroup -g 1000 nonroot && \
    adduser -u 1000 -G nonroot -D nonroot

COPY go.mod go.sum Makefile ./
COPY tools tools
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    make gen/go/rig/go.mod && \
    make buf protoc-gen-connect-go protoc-gen-go && \
    go mod download # && go mod verify

COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    make build-rig-operator

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/bin/rig-operator /usr/local/bin/
COPY --from=builder /etc/passwd /etc/passwd

USER 1000

CMD ["rig-operator"]
