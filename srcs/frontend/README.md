# Frontend

## Identity
This module contains the Next.js React frontend for the One Human Corp Agentic OS, providing human oversight and operational dashboards.

## Architecture
Built with Next.js and React, it communicates with the Go backend via REST/JSON. It adheres to 'Apple Standard' aesthetics, requiring high contrast, subtle borders, and glassmorphism.

## Quick Start
1. Install dependencies: `npm install`
2. Run the development server: `npm run dev`

## Developer Workflow
Use Bazel to build and test:
`bazelisk build //srcs/frontend/...`
`bazelisk test //srcs/frontend/...`

## Configuration
Requires environment variables such as `NEXT_PUBLIC_API_URL` to point to the correct backend orchestration service.
