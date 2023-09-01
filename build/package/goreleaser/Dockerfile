FROM alpine AS builder

RUN apk add --no-cache ca-certificates

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY rig-server /usr/local/bin/
COPY rig-proxy /usr/local/bin/
COPY rig-admin /usr/local/bin/

CMD ["rig-server"]
