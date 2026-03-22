# ── Build Stage ──────────────────────────────────────────────
FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /medconnect-server ./cmd/server

# ── Runtime Stage ────────────────────────────────────────────
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata
RUN adduser -D -g '' appuser

WORKDIR /app
COPY --from=builder /medconnect-server .

USER appuser
EXPOSE 3000
CMD ["./medconnect-server"]
