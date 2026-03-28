# Settings Module

## Identity
The `settings` module manages the configuration state, environment variables, and organizational preferences for the One Human Corp platform.

## Architecture
It centralizes the loading, validation, and storage of global platform settings (like default LLM providers, budget limits, and UI themes). It exposes these configurations to other backend modules via a thread-safe registry.

## Developer Workflow
Ensure all new environment variables or global toggles are registered and validated within this module.

- **Test Settings**: `bazel test //srcs/settings/...`
