FROM alpine:3.19.1 AS builder

RUN apk add --no-cache ca-certificates && \
    addgroup -g 1000 nonroot && \
    adduser -u 1000 -G nonroot -D nonroot

FROM alpine:3.19.1

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
COPY rig /usr/local/bin/

USER 1000

CMD ["rig"]
