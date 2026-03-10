# mono

This repository is initialized as a Bazel-based monorepo:

- application and library code lives under `/home/runner/work/mono/mono/srcs`
- documentation lives under `/home/runner/work/mono/mono/docs`
- contract definitions live under `/home/runner/work/mono/mono/srcs/proto`

## Initial roadmap slice

The first implemented slice focuses on the Phase 1 foundation from `docs/roadmap.md`:

- Software Company default organization schema
- in-memory agent orchestration and virtual meeting rooms
- model-aware token cost tracking
- a small CEO dashboard HTTP interface

## Running tests

```bash
bazelisk test //...
```
