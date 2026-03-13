# IFINU Radio API

High-performance Go API for internet radio streaming with robust buffering and resilience.

## Features

- 🎵 **Radio Directory Integration** - Syncs with Radio Browser API
- 🚀 **High Performance Streaming** - Handles 500+ concurrent listeners
- 🔄 **Automatic Retry** - Reconnects to unstable streams
- 💾 **Smart Caching** - In-memory cache for fast responses
- 📡 **Buffered Streaming** - 64KB buffer prevents stalls
- 🔍 **Full-Text Search** - Search by name, country, tags
- 🐳 **Docker Ready** - Complete Docker setup included

## Architecture

Clean Architecture with separated layers:

```
/cmd/api              - Application entrypoint
/internal/
  /handlers           - HTTP handlers
  /services           - Business logic
  /repository         - Data access layer
  /stream             - Streaming engine
  /models             - Domain models
  /config             - Configuration
  /cache              - Caching layer
```

## Quick Start

### Using Docker (Recommended)

```bash
# Start all services
docker-compose up -d

# Check logs
docker-compose logs -f api

# Stop services
docker-compose down
```

The API will be available at `http://localhost:8080`

### Manual Setup

1. **Install dependencies**

```bash
go mod download
```

2. **Setup PostgreSQL**

```bash
createdb radio_db
```

3. **Configure environment**

```bash
cp .env.example .env
# Edit .env with your settings
```

4. **Run the application**

```bash
go run cmd/api/main.go
```

## API Endpoints

### Health Check
```
GET /health
```

Response:
```json
{
  "status": "ok",
  "database": "ok",
  "total_radios": 1500,
  "version": "1.0.0"
}
```

### List Radios
```
GET /api/v1/radios?limit=20&offset=0
```

Response:
```json
{
  "sucesso": true,
  "dados": [...],
  "total": 1500
}
```

### Search Radios
```
GET /api/v1/radios/search?q=rock&limit=20
```

### Get Radio Details
```
GET /api/v1/radios/{id}
```

### Stream Radio
```
GET /api/v1/radios/{id}/stream
```

This endpoint proxies the radio stream with:
- Automatic buffering (64KB)
- Retry on connection loss (3 attempts)
- Exponential backoff
- Client disconnect detection

### Admin: Force Sync (Protected)
```
POST /api/v1/admin/sync
Header: X-API-Key: your_api_key_here
```

**⚠️ Requires authentication** - Set `ADMIN_API_KEY` environment variable.

Response:
```json
{
  "sucesso": true,
  "mensagem": "Sync completed successfully",
  "total_radios": 1500
}
```

## Streaming System

The streaming engine is designed for maximum stability:

### Buffer Strategy
- Uses `io.CopyBuffer` with 64KB buffer
- Prevents buffering stalls on slow networks
- Minimal memory usage per connection

### Resilience Features
- **Auto-reconnect**: Reconnects to upstream on failure
- **Retry logic**: 3 attempts with exponential backoff
- **Context support**: Graceful shutdown on client disconnect
- **Timeout handling**: Configurable read/write timeouts

### Performance Tuning

HTTP Client configuration:
- MaxIdleConns: 100
- MaxIdleConnsPerHost: 50
- IdleConnTimeout: 90s

Tested with:
- ✅ 500+ concurrent streams
- ✅ Unstable radio servers
- ✅ Long-running connections (hours)
- ✅ Various audio formats (MP3, AAC, OGG)

## Configuration

All configuration via environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| PORT | 8080 | Server port |
| ENV | development | Environment (development/production) |
| DB_HOST | localhost | PostgreSQL host |
| DB_PORT | 5432 | PostgreSQL port |
| DB_USER | postgres | Database user |
| DB_PASSWORD | postgres | Database password |
| DB_NAME | radio_db | Database name |
| STREAM_BUFFER_SIZE | 65536 | Stream buffer (bytes) |
| STREAM_RETRY_ATTEMPTS | 3 | Max retry attempts |
| SYNC_INTERVAL | 6h | Sync frequency |
| **ADMIN_API_KEY** | - | **API key for admin routes (required)** |
| **ALLOWED_ORIGINS** | * | **CORS allowed origins (comma-separated)** |

⚠️ **Security**: See [SECURITY.md](SECURITY.md) for security best practices.

## Database Schema

```sql
CREATE TABLE radios (
    id SERIAL PRIMARY KEY,
    uuid VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    stream_url VARCHAR(500) NOT NULL,
    country VARCHAR(100),
    language VARCHAR(100),
    tags TEXT,
    bitrate INTEGER,
    favicon VARCHAR(500),
    homepage VARCHAR(500),
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

CREATE INDEX idx_radios_name ON radios(name);
CREATE INDEX idx_radios_country ON radios(country);
CREATE INDEX idx_radios_uuid ON radios(uuid);
```

## Background Jobs

### Radio Synchronization

Automatically syncs radio stations from Radio Browser API:

- Runs on startup
- Periodic sync every 6 hours (configurable)
- Upserts stations (updates existing, creates new)
- Clears cache after sync

## Logging

Uses `zerolog` for structured logging:

```json
{
  "level": "info",
  "time": 1234567890,
  "message": "Starting stream proxy",
  "radio_id": 42,
  "radio_name": "Rock FM",
  "stream_url": "http://..."
}
```

## Production Deployment

### Recommended Setup

1. **Use Docker Compose** for easy orchestration
2. **Set ENV=production** to enable optimizations
3. **Configure reverse proxy** (Nginx) for:
   - SSL termination
   - Load balancing
   - Rate limiting
4. **Monitor** with health endpoint
5. **Scale horizontally** - API is stateless (except cache)

### Nginx Example

```nginx
upstream radio_api {
    server localhost:8080;
}

server {
    listen 80;
    server_name api.ifinu.io;

    location / {
        proxy_pass http://radio_api;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;

        # Important for streaming
        proxy_buffering off;
        proxy_request_buffering off;
    }
}
```

## Development

### Run tests
```bash
go test ./...
```

### Format code
```bash
go fmt ./...
```

### Linting
```bash
golangci-lint run
```

## License

MIT

## Author

IFINU.IO Team
