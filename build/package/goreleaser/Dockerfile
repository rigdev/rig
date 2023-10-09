FROM alpine AS builder

RUN apk add --no-cache ca-certificates && \
    addgroup -g 1000 nonroot && \
    adduser -u 1000 -G nonroot -D nonroot

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY rig-operator /usr/local/bin/
COPY --from=builder /etc/passwd /etc/passwd

USER 1000

CMD ["rig-operator"]
