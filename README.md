# mono

This repository is initialized as a Bazel-based monorepo:

- application and library code lives under `srcs/`
- documentation lives under `docs/`
- contract definitions live under `srcs/proto/`
- Bazel is pinned to `9.0.0`
- Go modules target `1.25`

## Initial roadmap slice

The first implemented slice focuses on the Phase 1 foundation from `docs/roadmap.md`:

- Software Company default organization schema
- in-memory agent orchestration and virtual meeting rooms
- model-aware token cost tracking
- a small CEO dashboard HTTP interface with a message form for user interaction

## Running tests

```bash
bazelisk test //...
```
