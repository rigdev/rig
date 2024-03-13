FROM alpine:3.19.1 AS builder

RUN apk add --no-cache ca-certificates && \
    addgroup -g 1000 nonroot && \
    adduser -u 1000 -G nonroot -D nonroot

COPY rig /usr/local/bin/

USER 1000

CMD ["rig"]
