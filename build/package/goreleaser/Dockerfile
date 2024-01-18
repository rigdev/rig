FROM alpine AS builder

RUN apk add --no-cache ca-certificates && \
    addgroup -g 1000 nonroot && \
    adduser -u 1000 -G nonroot -D nonroot

FROM alpine:3
RUN apk add --no-cache curl

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY rig-operator /usr/local/bin/
COPY --from=builder /etc/passwd /etc/passwd

USER 1000

CMD ["rig-operator"]
