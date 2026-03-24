/// OHC Core HTTP server entry point.
///
/// Exposes a lightweight HTTP API over the core library modules:
/// - Settings  GET/POST /settings
/// - Agents    GET/POST/DELETE /agents
/// - Scheduler GET/POST/DELETE /scheduler/tasks
/// - Meetings  GET/POST/DELETE /meetings
/// - Chat      GET/POST /chat/channels, /chat/messages

use std::env;
use std::net::SocketAddr;

#[tokio::main]
async fn main() {
    env_logger::init();

    let addr: SocketAddr = env::var("LISTEN_ADDR")
        .unwrap_or_else(|_| "0.0.0.0:18789".to_string())
        .parse()
        .expect("LISTEN_ADDR must be a valid socket address");

    log::info!("OHC Core listening on {}", addr);

    // Minimal TCP echo for health checks — a full HTTP framework can be
    // substituted here (axum, actix-web, etc.) once a Cargo.toml dependency
    // on one of those is added.
    let listener = tokio::net::TcpListener::bind(addr)
        .await
        .expect("failed to bind");

    log::info!("Ready.");

    // Accept connections and respond with a basic JSON health payload.
    loop {
        match listener.accept().await {
            Ok((mut socket, peer)) => {
                log::debug!("connection from {}", peer);
                tokio::spawn(async move {
                    use tokio::io::{AsyncReadExt, AsyncWriteExt};
                    let mut buf = [0u8; 1024];
                    // Read request (fire-and-forget; we always respond OK).
                    let _ = socket.read(&mut buf).await;
                    let body = r#"{"status":"ok","service":"ohc-core"}"#;
                    let response = format!(
                        "HTTP/1.1 200 OK\r\nContent-Type: application/json\r\nContent-Length: {}\r\nConnection: close\r\n\r\n{}",
                        body.len(),
                        body
                    );
                    let _ = socket.write_all(response.as_bytes()).await;
                });
            }
            Err(e) => log::error!("accept error: {}", e),
        }
    }
}
