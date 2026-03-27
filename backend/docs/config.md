# Configuration — Environment & Deployment

> **Sources:**
> - [`backend/cmd/server/main.go`](../cmd/server/main.go) — Environment loading & server setup
> - [`backend/Dockerfile`](../Dockerfile) — Container build
> - [`docker-compose.yml`](../../docker-compose.yml) — Orchestration

---

## Overview

MedConnect backend uses environment variables for configuration, loaded via `godotenv` from a `.env` file in the working directory. All sensitive values (keys, secrets) are required — no insecure defaults.

---

## Environment Variables

### Required

| Variable     | Description                          | Format              | Example                                |
| ------------ | ------------------------------------ | ------------------- | -------------------------------------- |
| `AES_KEY`    | AES-256-GCM encryption key           | 64-char hex string  | `0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef` |
| `JWT_SECRET` | JWT signing secret                   | Min 32 chars        | `my-super-secret-jwt-key-at-least-32`  |

### Server

| Variable         | Default    | Description                 |
| ---------------- | ---------- | --------------------------- |
| `SERVER_PORT`    | `3000`     | HTTP server listen port     |
| `ALLOWED_ORIGINS`| `http://localhost:5173,http://localhost:3000` | CORS allowed origins |

### Database

| Variable  | Default                                                                 | Description        |
| --------- | ----------------------------------------------------------------------- | ------------------ |
| `DB_DSN`  | `host=localhost user=medadmin password=securepass123 dbname=medconnect port=5432 sslmode=disable` | PostgreSQL DSN |

### AI (Ollama)

| Variable       | Default                  | Description           |
| -------------- | ------------------------ | --------------------- |
| `OLLAMA_URL`   | `http://localhost:11434` | Ollama API base URL   |
| `OLLAMA_MODEL` | `llama3`                 | Model for inference   |

### WhatsApp (Evolution API)

| Variable      | Default                  | Description                |
| ------------- | ------------------------ | -------------------------- |
| `WA_URL`      | `http://localhost:8080`  | Evolution API base URL     |
| `WA_TOKEN`    | *(empty — disabled)*     | API authentication token   |
| `WA_INSTANCE` | `medconnect`             | Instance name              |

> **Note:** If `WA_TOKEN` is empty, WhatsApp notifications are silently disabled.

---

## Docker Configuration

### Dockerfile

Multi-stage build for production:

```dockerfile
# Build stage
FROM golang:latest AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /server ./cmd/server

# Runtime stage
FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
RUN adduser -D -g '' appuser
COPY --from=builder /server /server
USER appuser
EXPOSE 3000
ENTRYPOINT ["/server"]
```

**Build flags:**
- `-ldflags="-s -w"` strips debug symbols for smaller binary
- `CGO_ENABLED=0` for static linking
- Non-root `appuser` for security

### Docker Compose

```yaml
services:
  db:
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: medadmin
      POSTGRES_PASSWORD: securepass123
      POSTGRES_DB: medconnect
    volumes:
      - pgdata:/var/lib/postgresql/data
    ports:
      - "5432:5432"

  ollama:
    image: ollama/ollama:latest
    volumes:
      - ollama_data:/root/.ollama
    ports:
      - "11434:11434"

  backend:
    build: ./backend
    ports:
      - "3000:3000"
    environment:
      DB_DSN: "host=db user=medadmin password=securepass123 dbname=medconnect port=5432 sslmode=disable"
      AES_KEY: "${AES_KEY}"
      JWT_SECRET: "${JWT_SECRET}"
      OLLAMA_URL: "http://ollama:11434"
    depends_on:
      - db
      - ollama

  frontend:
    build: ./frontend
    ports:
      - "5173:80"
    depends_on:
      - backend

volumes:
  pgdata:
  ollama_data:
```

---

## Run Script

The `run.sh` script at project root provides convenience commands:

```bash
./run.sh start    # docker-compose up -d
./run.sh stop     # docker-compose down
./run.sh logs     # docker-compose logs -f
./run.sh rebuild  # docker-compose build --no-cache
```

---

## Security Considerations

1. **AES_KEY** must be a cryptographically random 64-character hex string (256 bits)
   ```bash
   openssl rand -hex 32
   ```

2. **JWT_SECRET** should be at least 32 characters of high-entropy random data
   ```bash
   openssl rand -base64 48
   ```

3. **DB_DSN** should use strong passwords in production
4. **ALLOWED_ORIGINS** should be restricted to production frontend URL
5. **WA_TOKEN** should only be set when WhatsApp integration is deployed

---

## .env Template

```env
# Required
AES_KEY=your-64-char-hex-aes-key-here
JWT_SECRET=your-jwt-secret-min-32-chars-here

# Server
SERVER_PORT=3000
ALLOWED_ORIGINS=http://localhost:5173

# Database
DB_DSN=host=localhost user=medadmin password=securepass123 dbname=medconnect port=5432 sslmode=disable

# AI
OLLAMA_URL=http://localhost:11434
OLLAMA_MODEL=llama3

# WhatsApp (optional)
WA_URL=http://localhost:8080
WA_TOKEN=
WA_INSTANCE=medconnect
```
