# OpenClaw Manager Desktop

A Tauri 2.0 desktop application for managing the [OpenClaw](https://github.com/miaoxworld/openclaw-manager) AI agent platform.

## Features

- **Dashboard** — Real-time service status, start/stop/restart/diagnose, live logs
- **AI Model Config** — 14+ AI providers (Anthropic, OpenAI, DeepSeek, Moonshot, Gemini, Azure, Groq, Mistral, Cohere, xAI, Ollama, LM Studio, Qwen, Zhipu)
- **Message Channels** — Telegram, Discord, Slack, Feishu, WeChat, iMessage, DingTalk, QQ, WhatsApp, LINE
- **Agent Management** — Create/edit/delete agents with channel binding and tool permissions
- **Skills Library** — Browse, install, configure skill plugins
- **Security** — IP exposure detection, port binding check, token auth, file permissions, one-click fix
- **Diagnostics** — System env check, AI connection test, channel connectivity test
- **Service Logs** — Real-time log viewer with search and export
- **Settings** — Language, startup options, update management
- **Theme Toggle** — Light/dark mode

## Prerequisites

- Node.js 18+
- Rust 1.77+
- `npm install -g openclaw`

## Development

```bash
# Install dependencies
npm install

# Start dev server
npm run tauri dev
```

## Build

```bash
npm run tauri build
```

## Architecture

- **Frontend**: React 18 + TypeScript + TailwindCSS + Framer Motion + Lucide React
- **Backend**: Rust + Tauri 2.0
- **Config**: `~/.openclaw/openclaw.json` and `~/.openclaw/.env`

## Service

The managed service is the `openclaw` npm package which runs on port **18789** by default.
