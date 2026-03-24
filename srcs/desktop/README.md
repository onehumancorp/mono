# OpenClaw Manager Desktop

A Tauri 2.0 desktop application for managing the [OpenClaw](https://github.com/miaoxworld/openclaw-manager) AI agent platform.

## Identity
The `desktop` module serves as the primary visual interface for local deployment and management of the OpenClaw AI agent platform, providing real-time status, diagnostics, and skill configuration directly to the user's local operating system.

## Architecture
This application utilizes a hybrid architecture:
- **Frontend**: React 18 + TypeScript + TailwindCSS + Framer Motion + Lucide React, ensuring a highly responsive and styled UI.
- **Backend**: Rust + Tauri 2.0, providing lightweight, high-performance local system bindings.
- **Config**: Relies on local file system configurations at `~/.openclaw/openclaw.json` and `~/.openclaw/.env`.

## Quick Start
1. Ensure `npm` and `rustc` are installed on your system.
2. Install the OpenClaw service globally:
   ```bash
   npm install -g openclaw
   ```
3. Run the desktop application locally:
   ```bash
   npm run tauri dev
   ```

## Developer Workflow
Development requires Node.js 18+ and Rust 1.77+.
- **Install dependencies**: `npm install`
- **Start dev server**: `npm run tauri dev`
- **Build production release**: `npm run tauri build`

## Configuration
The `desktop` module manages configuration directly on the host system:
- **AI Model Config** — Supports 14+ AI providers (Anthropic, OpenAI, DeepSeek, etc.).
- **Message Channels** — Integrates with Telegram, Discord, Slack, and more.
- Security validations (IP exposure, port bindings) and local permissions are verified automatically.
- **Service Port**: The managed `openclaw` npm package runs on port **18789** by default.
