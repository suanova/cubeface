# syntax=docker/dockerfile:1.7

# ---- frontend -------------------------------------------------------------
FROM oven/bun:1 AS web
WORKDIR /src/web
COPY web/package.json web/bun.lock ./
RUN bun install --frozen-lockfile
COPY web/ ./
RUN bun run build

# ---- backend --------------------------------------------------------------
FROM golang:1.27-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=web /src/web/dist ./web/dist
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/cubeface .

# ---- runtime --------------------------------------------------------------
FROM alpine:3.22
RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S -g 10001 cubeface \
    && adduser -S -u 10001 -G cubeface -h /home/cubeface cubeface \
    && mkdir -p /data /home/cubeface \
    && chown -R 10001:10001 /data /home/cubeface

COPY --from=build --chown=10001:10001 /out/cubeface /usr/local/bin/cubeface

ENV APP_ENV=production \
    HOME=/home/cubeface \
    SQLITE_PATH=/data/cubeface.db

USER 10001:10001
EXPOSE 3000
ENTRYPOINT ["/usr/local/bin/cubeface"]
