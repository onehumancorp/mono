//! # OHC Core
//!
//! Cross-platform core library for One Human Corp.
//!
//! Provides the fundamental building blocks shared between:
//! - **Single-docker deployments** — all logic runs in-process via the
//!   embedded Tokio runtime.
//! - **Cloud-native (Kubernetes) deployments** — the same library is used by
//!   the backend server; the storage backends swap to PostgreSQL/Redis.
//! - **Desktop / mobile apps** (macOS, Windows, Linux, iOS, Android) —
//!   called via FFI from the Flutter cross-platform app (`srcs/app`).
//!
//! ## Modules
//!
//! | Module | Responsibility |
//! |--------|----------------|
//! | [`settings`] | Load, persist, and watch application configuration |
//! | [`agents`] | Register and manage AI agent lifecycle |
//! | [`scheduler`] | Schedule agent tasks (once / interval / cron) |
//! | [`meeting`] | Create and manage virtual meeting rooms |
//! | [`chat`] | Unified chat integration (Centrifuge native, Chatwoot, Slack, Telegram, …) |

pub mod agents;
pub mod chat;
pub mod meeting;
pub mod scheduler;
pub mod settings;
