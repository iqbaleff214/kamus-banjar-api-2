# syntax=docker/dockerfile:1

# ─── base ────────────────────────────────────────────────────────────────────
FROM golang:1.26-alpine AS base
WORKDIR /app
RUN apk add --no-cache git ca-certificates tzdata

# ─── dev (hot-reload via Air) ─────────────────────────────────────────────────
FROM base AS dev
RUN go install github.com/air-verse/air@latest
COPY go.mod go.sum* ./
RUN go mod download || true
EXPOSE 3000
CMD ["air", "-c", ".air.toml"]

# ─── builder ─────────────────────────────────────────────────────────────────
FROM base AS builder
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /bin/api ./cmd/api

# ─── seeder ──────────────────────────────────────────────────────────────────
FROM base AS seeder-builder
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /bin/seeder ./scripts/seed

FROM scratch AS seeder
COPY --from=seeder-builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=seeder-builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=seeder-builder /bin/seeder /seeder
COPY --from=seeder-builder /app/scripts/seed/seed_data.json /scripts/seed/seed_data.json
ENTRYPOINT ["/seeder"]

# ─── prod ────────────────────────────────────────────────────────────────────
FROM scratch AS prod
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /bin/api /api
EXPOSE 3000
USER 65534:65534
ENTRYPOINT ["/api"]
