# Orchestration Module

## Identity
The Orchestration Module is the central nervous system of the Agentic OS, managing the interaction protocols, virtual meetings, and asynchronous pub/sub execution of the AI swarm.

## Architecture
This Golang package exposes the `Hub`, the message router. Agents emit and consume structured `Message` objects within `MeetingRoom`s, preventing context bloat while supporting concurrent multi-agent decision making.

## Quick Start
1. Ensure Bazel is active.
2. Build the module: `bazelisk build //srcs/orchestration/...`

## Developer Workflow
- Execute tests locally via `bazelisk test //srcs/orchestration/...`
- Expand `Message` payloads and payload typing carefully, as all agents parse these signals to drive execution.

## Configuration
- Default settings initialize in memory, designed to be backed by persistent append-only logs in production.
