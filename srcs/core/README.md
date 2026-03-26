# OHC Core — Rust Library

## Identity
The `core` module is a high-performance Rust library for One Human Corp, providing maximum throughput for performance-critical logic such as agent scheduling, meeting rooms, and chat integrations.

## Architecture

It is embedded as a sidecar in cloud-native (Kubernetes) deployments or acts as the main handler in single-docker deployments.

### Modules

| Module | Purpose |
|--------|---------|
| [`settings`](src/settings/mod.rs) | Load, persist, and watch application settings |
| [`agents`](src/agents/mod.rs) | Register and manage AI agent lifecycle (hire / fire / status) |
| [`scheduler`](src/scheduler/mod.rs) | Schedule agent tasks — once, interval, or cron |
| [`meeting`](src/meeting/mod.rs) | Open, join, and close virtual meeting rooms |
| [`chat`](src/chat/mod.rs) | Unified chat integration (Chatwoot, Slack, Telegram, Discord, …) |

All data operations are **tenant-scoped**: every struct carries an
`organization_id` field, and every store implementation filters by it.

### Design Goals

- **Zero external runtime dependencies** — no JVM, no Python interpreter
- **Tenant isolation** — every data access is scoped to an `organization_id`
- **Pluggable storage** — in-memory for desktop/dev; swap in SQLite/PostgreSQL for production via trait objects
- **Async-first** — built on Tokio for high concurrency without blocking threads

## Quick Start

The `ohc-core` binary exposes a minimal HTTP health endpoint on `$LISTEN_ADDR`
(default `0.0.0.0:18789`):

```bash
LISTEN_ADDR=0.0.0.0:18789 cargo run --release --bin ohc-core
```

### Docker

```bash
docker build -f deploy/docker/ohc-core/Dockerfile -t ohc-core .
docker run -p 18789:18789 ohc-core
```

## Developer Workflow

- **Build**: `cargo build --release`
- **Test**: `cargo test`
- **Lint**: `cargo clippy`

## Configuration

The core library uses the following environment variables:
- `LISTEN_ADDR`: The address and port to listen on (e.g., `0.0.0.0:18789`).
