# Billing Module

## Identity
The Billing Module accurately tracks model-aware token pricing and execution costs, giving the CEO total financial visibility.

## Architecture
Built around a concurrent `Tracker` instance mapping agent usage telemetry to baseline `Price` catalogs containing Anthropic, OpenAI, and Google models. Intercepts inference requests globally to sum up `costUsd` natively in Go memory structures.

## Quick Start
1. Ensure Bazel is installed.
2. Build the module: `bazelisk build //srcs/billing/...`

## Developer Workflow
- Run unit tests and benchmark token computation logic: `bazelisk test //srcs/billing/...`
- Use Golang style and hermetic mocks when testing.

## Configuration
- Currently initializes against a static `DefaultCatalog`. In the future, this maps to environment variables controlling cost constraints or budget limits.
