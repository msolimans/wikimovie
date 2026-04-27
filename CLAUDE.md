# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

WIKIMovie is a Go-based data lake service for movie information. It has two independently deployable binaries — an HTTP query **service** and a background **worker** — plus a **seed** utility. All three share packages under `pkg/`.

## Commands

### Build & Run

```bash
# Build all packages (used in CI)
go build -v ./...

# Run the API service (port 8080)
go run cmd/service/main.go

# Run the SQS worker
go run cmd/worker/main.go

# Load seed data into S3 (requires LocalStack running)
go run cmd/seed/main.go s3
```

### Testing

```bash
# Run all tests
go test ./...

# Run tests in a specific package
go test ./pkg/es/...
go test ./pkg/s3/...
go test ./pkg/sqs/...
go test ./pkg/config/...

# Run a single test by name
go test ./pkg/es/... -run TestESClient
```

### Local Development (LocalStack)

```bash
# Start LocalStack with Elasticsearch, S3, and SQS
docker-compose up

# Initialize LocalStack resources (queues, buckets, ES index)
# Run manually after localstack starts (init script unreliable)
./localstack.sh

# Verify SQS queues are up
AWS_REGION=us-east-1 aws --endpoint-url=http://localhost.localstack.cloud:4566 sqs list-queues

# Clean up all LocalStack resources
./clean.sh
```

> **Note:** Add `127.0.0.1 localhost.localstack.cloud` to `/etc/hosts` if DNS resolution fails for LocalStack endpoints.

### Docker

```bash
# Build image (builds both service and worker binaries)
docker build -t wikimovie .

# Run service (default CMD)
docker run -it --rm wikimovie

# Run worker
docker run -it --rm wikimovie worker
```

### Kubernetes / Skaffold

```bash
skaffold dev   # build + deploy to local K8s cluster (watches for changes)
skaffold run   # one-shot build + deploy
```

## Architecture

The project is split into three `cmd/` entry points and shared `pkg/` packages:

```
cmd/
  service/   → HTTP API server (Fiber, port 8080)
  worker/    → SQS long-poll daemon
  seed/      → One-shot data loader (fixtures → S3)
pkg/
  appconf/   → App-level config struct + env var bindings
  config/    → Viper YAML loader (merges defaults.yml + config_{ENV}.yml)
  es/        → Elasticsearch client wrapper
  s3/        → S3 client, manager (download/upload/move), downloader, uploader
  sqs/       → SQS client + worker (long-poll loop with retry backoff)
  routes/
    query/   → Movie search handlers (Fiber routes)
    monitor/ → Health check handlers
  handlers/  → Business logic: download from S3 → parse JSON → index to ES → move file
  utils/     → IHandler interface for route registration
  testutil/  → Shared mock helpers
```

### Data Flow

**Ingestion path:** Provider uploads a JSON file to S3 → S3 event notification goes to SQS → `worker` polls SQS → downloads file → parses movie documents → bulk-indexes to Elasticsearch → moves file to `processed/` prefix with a timestamp suffix.

**Query path:** Client sends HTTP request to `service` → Fiber route handler builds Elasticsearch DSL query → returns results as JSON.

### Configuration

Configuration uses a three-layer merge via Viper:

1. `defaults.yml` — base defaults for all environments
2. `config_{ENV}.yml` — environment overrides (e.g., `config_dev.yml` sets LocalStack endpoints)
3. Environment variables — mapped in `pkg/appconf/init.go`

Key env vars: `ENV`, `ELASTICSEARCH_URLS`, `AWS_REGION`, `AWS_ENDPOINT` (LocalStack only), `WORKER_QUEUE`, `BUCKET`. `AWS_ENDPOINT` should be left unset in production — the AWS SDK will use real endpoints automatically.

### Route Handler Pattern

Every route group implements the `IHandler` interface (`pkg/utils/fiber.go`). Handlers are constructed with their dependencies (e.g., `*es.ESClient`), then registered by calling `.Routes(app, router)`. See `pkg/routes/query/movie.go` and `pkg/routes/monitor/monitor.go` for examples.

### Elasticsearch Index

The movies index is hardcoded as `"my-index"` in `pkg/routes/query/movie.go:13`. Document IDs use the format `{year}-{title}` (spaces → dashes). Movie documents have fields: `title` (string), `year` (number), `cast` ([]string), `genres` ([]string), `href` (string, optional).

- Title/year searches use ES `match` queries.
- Genre/cast searches use ES `terms` queries.
- The `/movies/query` endpoint builds a `bool/must` query dynamically from whichever params are present.

### Known Issue

`getByCast` in `pkg/routes/query/movie.go:115` has a bug: it queries the `genres` field instead of `casts`. Fix by changing the query field to `"casts"`.

### Auth Middleware

Bearer token authentication is wired in the service but is a stub for demonstration — it needs to be integrated with a real auth provider (Okta, Firebase, etc.) before production use.

### SQS Worker Retry Logic

`pkg/sqs/worker.go` implements long polling with configurable retry intervals (`Worker.RetryIntervals` in config, default `[30, 60, 90]` seconds). Failed messages are not deleted so SQS visibility timeout eventually re-queues them.

## Infrastructure

- **K8s manifests:** `infra/k8s/` — includes HPA (min 2, max 10 replicas) and topology spread constraints.
- **Terraform:** `infra/aws/` — provisions AWS resources (EKS, S3, SQS, ES).
- **Skaffold:** `skaffold.yml` targets the K8s manifests and the Dockerfile.

## CI

GitHub Actions (`.github/workflows/go.yml`) runs `go build -v ./...` on pushes and PRs to `master`. There is no automated test step in CI yet.
