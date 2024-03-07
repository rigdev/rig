FROM alpine AS builder

RUN apk add --no-cache ca-certificates && \
    addgroup -g 1000 nonroot && \
    adduser -u 1000 -G nonroot -D nonroot

FROM alpine:3
RUN apk add --no-cache curl

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY rig-operator /usr/local/bin/
COPY rigdev.sidecar /usr/local/bin/plugin/
COPY rigdev.init_container /usr/local/bin/plugin/
COPY rigdev.object_template /usr/local/bin/plugin/
COPY rigdev.annotations /usr/local/bin/plugin/
COPY rigdev.datadog /usr/local/bin/plugin/
COPY rigdev.google_sql_proxy /usr/local/bin/plugin/
COPY rigdev.env_mapping /usr/local/bin/plugin/
COPY rigdev.placement /usr/local/bin/plugin/
COPY --from=builder /etc/passwd /etc/passwd

USER 1000

CMD ["rig-operator"]
