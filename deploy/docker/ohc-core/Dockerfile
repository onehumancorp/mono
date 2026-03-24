# Build stage — compile the ohc-core binary.
FROM rust:1.82-slim AS builder

WORKDIR /app

# Cache dependencies by copying manifests first.
COPY srcs/core/Cargo.toml srcs/core/Cargo.lock* ./srcs/core/
RUN mkdir -p srcs/core/src && echo "fn main() {}" > srcs/core/src/main.rs
WORKDIR /app/srcs/core
RUN cargo build --release 2>/dev/null || true

# Copy real source and rebuild (incremental).
COPY srcs/core/src ./src/
RUN touch src/lib.rs && cargo build --release

# ─── Runtime stage ────────────────────────────────────────────────────────────
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
      ca-certificates \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/srcs/core/target/release/ohc-core /usr/local/bin/ohc-core

ENV LISTEN_ADDR="0.0.0.0:18789"
EXPOSE 18789

ENTRYPOINT ["/usr/local/bin/ohc-core"]
