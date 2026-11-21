FROM alpine:latest AS builder

WORKDIR /app

ARG TARGETOS
ARG TARGETARCH

RUN mkdir -p /app/data

COPY --chmod=0755 backend-artifacts/ech0-${TARGETOS}-${TARGETARCH} /app/ech0

FROM alpine:latest

# The server only writes under /app/data (SQLite, logs, uploads, migration temp
# dirs) and binds the unprivileged port 6277, so it never needs root.
RUN addgroup -S -g 10001 ech0 \
    && adduser -S -D -H -G ech0 -h /app -u 10001 ech0 \
    && apk add --no-cache su-exec

WORKDIR /app

COPY --chown=10001:10001 --from=builder /app /app
COPY --chmod=0755 --chown=0:0 docker/entrypoint.sh /app/entrypoint.sh

EXPOSE 6277

# Root only until entrypoint.sh has handed /app/data to uid 10001; the server
# itself always runs as 10001 (su-exec replaces this process).
USER root

ENTRYPOINT ["/app/entrypoint.sh"]

CMD ["serve"]
