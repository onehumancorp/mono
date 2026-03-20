# Design Doc: Virtual Meeting Rooms

**Author(s):** Antigravity
**Status:** In Review
**Last Updated:** 2026-03-17

## Overview
Virtual Meeting Rooms are collaborative spaces where multiple AI agents (and human managers) gather to discuss, debate, and resolve complex tasks. They provide a shared context window and a persistent transcript for all participants.

## Goals
- Facilitate synchronous multi-agent collaboration.
- Maintain a persistent, auditable transcript of all discussions.
- Provide a "Whiteboard" for shared artifacts and state.

## Proposed Design
A Meeting Room is a resource managed by the `Hub`. It uses a WebSocket/Server-Sent Events (SSE) connection to provide real-time updates to all participants.

### Shared Context
The "Whiteboard" or shared context window is a structured document that agents can read and write to. It ensures that all participants have the same "Ground Truth" during a discussion.

## Alternatives Considered
- **Direct Agent-to-Agent Messaging**: Rejected for complex tasks because it lacks a unified state and makes it difficult for a human manager to oversee the collaboration.

## Implementation Details
- **Architecture**: A Go 1.26 `Hub` orchestrates the room lifecycle, managing WebSocket/SSE connections pushing real-time events to the React/Next.js UI.
- **State Storage**: The meeting "Whiteboard" and transcript are fully persisted into Postgres as an append-only `events.jsonl` stream, ensuring no data loss on pod failure.
- **Concurrency Control**: Updates to the shared transcript are strictly serialized by the Hub using a synchronized mutex or transactional database row locking to prevent interleaving race conditions.

## Edge Cases
- **Disconnects**: If an agent's network connection to the Hub drops, they enter a `RECONNECTING` state and must re-fetch missed transcript history before contributing further.
- **Context Bloat**: Long-lived meeting rooms aggressively summarize early history; agents joining late only receive the summarized "Current State" and recent chat lines to prevent immediate context exhaustion.
- **Spamming**: A malfunctioning agent spamming messages is detected by rate-limiters on the Hub, which forcefully "evicts" the agent and alerts the CEO for manual intervention.
