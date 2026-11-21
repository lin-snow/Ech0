
FROM node:26-alpine AS frontend-builder

WORKDIR /web

COPY web/package.json web/pnpm-lock.yaml web/pnpm-workspace.yaml ./

RUN npm install -g pnpm@12.16.1

RUN pnpm install --frozen-lockfile

COPY web/ .

RUN pnpm run build --mode production

FROM golang:1.27.4-alpine AS backend-builder

RUN apk add --no-cache git ca-certificates gcc musl-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download \
    && go mod download github.com/mattn/go-sqlite3 \
    && mkdir -p /usr/local/include/sqlite-h \
    && cp "$(go list -m -f '{{.Dir}}' github.com/mattn/go-sqlite3)/sqlite3-binding.h" /usr/local/include/sqlite-h/sqlite3.h

# sqlite-vec's cgo binding includes <sqlite3.h>, which nothing provides for the
# pure-Go driver: go-sqlite3 ships only sqlite3-binding.h, and musl headers lack
# the BSD u_int*_t aliases it uses. Mirrors the release workflow's build step.
ENV CGO_CFLAGS="-I/usr/local/include/sqlite-h -Du_int8_t=uint8_t -Du_int16_t=uint16_t -Du_int64_t=uint64_t -Wno-typedef-redefinition"

COPY . .
COPY --from=frontend-builder /template/dist /app/template/dist

ARG TARGETOS
ARG TARGETARCH
ARG GIT_COMMIT
ARG BUILD_TIME

RUN COMMIT="${GIT_COMMIT:-$(git rev-parse --short HEAD 2>/dev/null || echo unknown)}" \
    && BUILD="${BUILD_TIME:-$(date -u +%Y-%m-%dT%H:%M:%SZ)}" \
    && CGO_ENABLED=1 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build \
    -tags netgo \
    -ldflags="-linkmode external -extldflags '-static' -X github.com/lin-snow/ech0/internal/version.Commit=${COMMIT} -X github.com/lin-snow/ech0/internal/version.BuildTime=${BUILD}" \
    -o ech0 ./cmd/ech0/main.go

FROM alpine:latest

WORKDIR /app

# The server only writes under /app/data (SQLite, logs, uploads, migration temp
# dirs) and binds the unprivileged port 6277, so it never needs root.
RUN apk add --no-cache tzdata su-exec \
    && addgroup -S -g 10001 ech0 \
    && adduser -S -D -H -G ech0 -h /app -u 10001 ech0 \
    && mkdir -p /app/data \
    && chown 10001:10001 /app/data

COPY --chown=10001:10001 --from=backend-builder /app/ech0 /app/ech0
COPY --chmod=0755 --chown=0:0 docker/entrypoint.sh /app/entrypoint.sh

RUN chmod +x /app/ech0

EXPOSE 6277

# Root only until entrypoint.sh has handed /app/data to uid 10001; the server
# itself always runs as 10001 (su-exec replaces this process).
USER root

ENTRYPOINT ["/app/entrypoint.sh"]

CMD ["serve"]
