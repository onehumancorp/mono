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
