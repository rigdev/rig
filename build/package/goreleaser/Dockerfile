FROM alpine AS builder

RUN apk add --no-cache ca-certificates && \
    addgroup -g 1000 nonroot && \
    adduser -u 1000 -G nonroot -D nonroot

FROM alpine:3
RUN apk add --no-cache curl

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY rig-operator /usr/local/bin/
COPY sidecar /usr/local/bin/plugin/
COPY init_container /usr/local/bin/plugin/
COPY object_template /usr/local/bin/plugin/
COPY annotations /usr/local/bin/plugin/
COPY datadog /usr/local/bin/plugin/
COPY google_sql_proxy /usr/local/bin/plugin/
COPY env_mapping /usr/local/bin/plugin/
COPY --from=builder /etc/passwd /etc/passwd

USER 1000

CMD ["rig-operator"]
