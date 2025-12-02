# BUGVay

**Production-ready automated bug bounty platform** built with Go, Gin, Asynq, ClickHouse, and React.

---

## ✨ Features

- **Multi-scanner architecture**: XSS, SQLi, LFI, Open Redirect (pluggable)
- **Distributed task queue** powered by Asynq (Redis-backed)
- **High-performance analytics** with ClickHouse for billions of scan results
- **Rate-limited HTTP client** with exponential backoff
- **URL canonicalization & deduplication** for efficient scanning
- **RESTful API v1** with clean service layer architecture
- **Modern React dashboard** (coming soon)

---

## 🏗️ Architecture

```
BUGVay/
├── cmd/
│   ├── api/         # HTTP API server
│   └── worker/      # Asynq background workers
├── internal/
│   ├── api/         # Gin router & handlers
│   ├── config/      # Configuration management
│   ├── database/    # Postgres + ClickHouse clients
│   ├── queue/       # Asynq queue client
│   ├── scanners/    # Scanner modules (XSS, SQLi, etc.)
│   ├── services/    # Business logic layer
│   └── worker/      # Worker & HTTP client
└── migrations/      # Database migrations
```

---

## 🚀 Quick Start

### Prerequisites

- **Go 1.21+**
- **PostgreSQL 14+** (already running on port 5432)
- **Docker & Docker Compose** (for Redis + ClickHouse)

### 1. Clone & Setup

```bash
cd /home/kokuroshesh/tools/BUGVay
cp .env.example .env
```

### 2. Start Infrastructure

```bash
make dev
```

This starts:
- **Redis** (Asynq queue) on `localhost:6379`
- **ClickHouse** (analytics) on `localhost:9000`

### 3. Run Migrations

```bash
# Postgres (indexes + constraints)
make migrate-up

# ClickHouse (scan results table)
make migrate-clickhouse
```

### 4. Start Services

**Terminal 1 - API Server:**
```bash
make api
```

**Terminal 2 - Worker:**
```bash
make worker
```

---

## 📡 API Endpoints

Base URL: `http://localhost:8080/api/v1`

### Programs
- `GET /programs` - List programs
- `POST /programs` - Create program
- `GET /programs/:id` - Get program

### Endpoints
- `POST /endpoints/upload` - Upload endpoints.txt
- `GET /endpoints` - List endpoints
- `GET /endpoints/:id` - Get endpoint

### Scans
- `POST /scans` - Create scan job
- `GET /scans` - List scans
- `GET /scans/:id` - Get scan status

### Findings
- `GET /findings` - List findings (filterable)
- `GET /findings/:id` - Get finding
- `PATCH /findings/:id/triage` - Triage finding

### Jobs
- `GET /jobs` - List Asynq jobs
- `GET /jobs/:id` - Get job status

---

## 💻 Usage Example

### 1. Create a Program

```bash
curl -X POST http://localhost:8080/api/v1/programs \
  -H "Content-Type: application/json" \
  -d '{"name":"HackerOne"}'
```

### 2. Upload Endpoints

```bash
# Create endpoints.txt
cat > endpoints.txt <<EOF
https://example.com/search?q=test
https://example.com/api/user?id=123
EOF

# Upload
curl -X POST http://localhost:8080/api/v1/endpoints/upload \
  -F "asset_id=1" \
  -F "file=@endpoints.txt"
```

### 3. Trigger XSS Scan

```bash
curl -X POST http://localhost:8080/api/v1/scans \
  -H "Content-Type: application/json" \
  -d '{
    "endpoint_ids": [1, 2],
    "scanners": ["xss"],
    "concurrency": 5,
    "rate_limit": 10
  }'
```

### 4. Check Findings

```bash
curl http://localhost:8080/api/v1/findings?severity=medium
```

---

## 🔧 Configuration

Edit `.env` for your environment:

```env
# API
API_PORT=8080
API_HOST=0.0.0.0

# PostgreSQL (existing database)
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_DB=bugvay

# Redis (Asynq)
REDIS_HOST=localhost
REDIS_PORT=6379

# ClickHouse
CLICKHOUSE_HOST=localhost
CLICKHOUSE_PORT=9000
CLICKHOUSE_DB=bugvay

# Worker
WORKER_CONCURRENCY=10
WORKER_RATE_LIMIT=50

# Scanner
SCANNER_TIMEOUT=30
SCANNER_MAX_RETRIES=3
```

---

## 🧪 Development

### Build Binaries

```bash
make build
```

### Run Tests

```bash
make test
```

### Clean Up

```bash
make clean
```

---

## 📊 Database Schema

### PostgreSQL (Transactional)

- `programs` - Bug bounty programs
- `assets` - In-scope domains/URLs
- `endpoints` - Discovered endpoints (deduplicated)
- `findings` - Security vulnerabilities

### ClickHouse (Analytics)

- `scan_results` - Raw HTTP responses & evidence (billions of rows)
- `scan_stats_hourly` - Aggregated metrics
- `evidence_index` - Deduplication index

---

## 🎯 Scanners

### XSS Scanner (MVP)
- Reflection-based detection
- Context-aware payloads
- Parameter fuzzing
- Evidence collection

### Coming Soon
- SQL Injection (time-based + error-based)
- Local File Inclusion (path traversal)
- Open Redirect (header + meta)
- SSRF, IDOR, XXE

---

## 📦 Tech Stack

| Component | Technology |
|-----------|-----------|
| Backend | Go 1.21 + Gin |
| Task Queue | Asynq (Redis) |
| OLTP Database | PostgreSQL 18 |
| Analytics Database | ClickHouse 23.8 |
| Frontend | React + Vite + Tailwind (WIP) |
| Observability | Prometheus + Grafana (planned) |

---

## 🛡️ Production Checklist

- [x] Service layer architecture
- [x] API versioning (`/api/v1`)
- [x] Global error handler
- [x] Database indexes + constraints
- [x] ClickHouse optimizations (MergeTree + partitions)
- [x] Rate limiting & retry logic
- [x] URL canonicalization
- [ ] Authentication & RBAC
- [ ] WebSocket for real-time updates
- [ ] Comprehensive test suite
- [ ] CI/CD pipeline
- [ ] Docker production images
- [ ] Kubernetes manifests

---

## 📝 License

MIT

---

## 🙏 Credits

Built with:
- [Gin](https://github.com/gin-gonic/gin) - HTTP framework
- [Asynq](https://github.com/hibiken/asynq) - Distributed task queue
- [pgx](https://github.com/jackc/pgx) - PostgreSQL driver
- [ClickHouse Go](https://github.com/ClickHouse/clickhouse-go) - ClickHouse driver

---

**Happy Hunting!** 🐛
