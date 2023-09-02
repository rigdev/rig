FROM golang:1.20.6-alpine3.18 AS builder

ENV CGO_ENABLED=0
ENV GOBIN=/app/tools/bin

WORKDIR /app

RUN apk add --no-cache ca-certificates make git

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
    make proto build-rig-server build-rig-proxy build-rig-admin

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/bin/rig-server /usr/local/bin/
COPY --from=builder /app/bin/rig-proxy /usr/local/bin/
COPY --from=builder /app/bin/rig-admin /usr/local/bin/

CMD ["rig-server"]
