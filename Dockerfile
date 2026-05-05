# ── Stage 1: Build Go backend ─────────────────────────────────────────────────
FROM golang:1.25-alpine AS api-builder

WORKDIR /app

COPY league-api/go.mod league-api/go.sum ./
RUN go mod download

COPY league-api/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /league-api ./cmd/server

# ── Stage 2: Build React frontend ─────────────────────────────────────────────
FROM node:22-alpine AS ui-builder

WORKDIR /app

COPY league-ui/package.json league-ui/package-lock.json ./
RUN npm ci

COPY league-ui/ ./
# VITE_API_URL left unset: defaults to /api/v1 (relative), proxied by nginx
RUN npm run build

# ── Stage 3: Runtime ──────────────────────────────────────────────────────────
FROM nginx:1.27-alpine

# Install supervisor to manage nginx + api processes
RUN apk add --no-cache supervisor

# Copy built artifacts
COPY --from=api-builder /league-api /usr/local/bin/league-api
COPY --from=ui-builder  /app/dist   /usr/share/nginx/html

# Copy migrations so the api can run them on startup
COPY --from=api-builder /app/migrations /migrations

# nginx config: serve UI, proxy /api/v1 and /ws to backend on :8080
COPY docker/nginx.conf /etc/nginx/conf.d/default.conf

# supervisord config
COPY docker/supervisord.conf /etc/supervisor/conf.d/supervisord.conf

EXPOSE 80

CMD ["/usr/bin/supervisord", "-c", "/etc/supervisor/conf.d/supervisord.conf"]
