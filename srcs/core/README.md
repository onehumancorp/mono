# OHC Core — Rust Library

High-performance core library for One Human Corp.  Written in Rust for maximum
throughput in single-docker deployments and embedded as a sidecar in
cloud-native (Kubernetes) deployments.

## Modules

| Module | Purpose |
|--------|---------|
| [`settings`](src/settings/mod.rs) | Load, persist, and watch application settings |
| [`agents`](src/agents/mod.rs) | Register and manage AI agent lifecycle (hire / fire / status) |
| [`scheduler`](src/scheduler/mod.rs) | Schedule agent tasks — once, interval, or cron |
| [`meeting`](src/meeting/mod.rs) | Open, join, and close virtual meeting rooms |
| [`chat`](src/chat/mod.rs) | Unified chat integration (Chatwoot, Slack, Telegram, Discord, …) |

All data operations are **tenant-scoped**: every struct carries an
`organization_id` field, and every store implementation filters by it.

## Building

```bash
cargo build --release
```

## Testing

```bash
cargo test
```

## Running (HTTP server)

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

## Design Goals

- **Zero external runtime dependencies** — no JVM, no Python interpreter
- **Tenant isolation** — every data access is scoped to an `organization_id`
- **Pluggable storage** — in-memory for desktop/dev; swap in SQLite/PostgreSQL for production via trait objects
- **Async-first** — built on Tokio for high concurrency without blocking threads
