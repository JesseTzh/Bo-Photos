# syntax=docker/dockerfile:1

FROM node:22-bookworm-slim AS frontend-builder
WORKDIR /src/frontend
RUN corepack enable && corepack prepare pnpm@11.9.0 --activate
COPY frontend/package.json frontend/pnpm-lock.yaml frontend/pnpm-workspace.yaml ./
RUN pnpm install --frozen-lockfile
COPY frontend/ ./
RUN pnpm build

FROM golang:1.24-bookworm AS backend-builder
WORKDIR /src/backend
RUN apt-get update \
    && apt-get install -y --no-install-recommends build-essential \
    && rm -rf /var/lib/apt/lists/*
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/cmd ./cmd
COPY backend/internal ./internal
COPY backend/migrations ./migrations
RUN test -f ./cmd/server/main.go
RUN CGO_ENABLED=1 go build -trimpath -ldflags="-s -w" -o /out/bophotos ./cmd/server

FROM debian:bookworm-slim AS runtime
RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        ca-certificates \
        libheif1 \
        libimage-exiftool-perl \
        libraw20 \
        libvips42 \
        libvips-tools \
        tzdata \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY --from=backend-builder /out/bophotos /usr/local/bin/bophotos
COPY --from=frontend-builder /src/frontend/dist ./frontend/dist

ENV BOPHOTOS_ADDR=:8080 \
    BOPHOTOS_DATA_DIR=/data \
    BOPHOTOS_FRONTEND_DIR=/app/frontend/dist \
    BOPHOTOS_COOKIE_SECURE=false

VOLUME ["/data"]
EXPOSE 8080

CMD ["bophotos"]
