# OpenClaw Manager Desktop

## Identity
A Tauri 2.0 desktop application for managing the [OpenClaw](https://github.com/miaoxworld/openclaw-manager) AI agent platform. It provides a visual interface for managing AI agents, channels, skills, and platform configuration.

## Architecture
The OpenClaw Manager Desktop application utilizes a modern, hybrid architecture:
- **Frontend**: React 18 + TypeScript + TailwindCSS + Framer Motion + Lucide React for a responsive, accessible user interface.
- **Backend**: Rust + Tauri 2.0, providing native performance and deep OS integration.
- **Config**: Relies on `~/.openclaw/openclaw.json` and `~/.openclaw/.env` for configuration state.
- **Service Layer**: Manages the `openclaw` npm package which runs as a background service on port **18789** by default.

## Quick Start
1. Ensure you have the prerequisites installed: Node.js 18+, Rust 1.77+, and `npm install -g openclaw`.
2. Install dependencies:
   ```bash
   npm install
   ```
3. Start the development server:
   ```bash
   npm run tauri dev
   ```

## Developer Workflow
This project uses standard Node.js and Rust workflows managed via npm scripts.
- **Run dev server**: `npm run tauri dev`
- **Build application**: `npm run tauri build`
- **Format code**: Use standard Prettier and rustfmt conventions.

## Configuration
The application relies on the following configurations:
- Service runs on port `18789` by default.
- Local configuration files are stored at `~/.openclaw/openclaw.json` and `~/.openclaw/.env`.
- Integrates with 14+ AI providers (Anthropic, OpenAI, DeepSeek, Moonshot, Gemini, Azure, Groq, Mistral, Cohere, xAI, Ollama, LM Studio, Qwen, Zhipu).
