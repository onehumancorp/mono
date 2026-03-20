# Domain Module

## Identity
The Domain Module codifies the business logic that shapes how a company operates, defining structural hierarchy, roles, and AI operational parameters.

## Architecture
Core structs (`Organization`, `Member`, `RoleProfile`) define the structural boundaries, capabilities, and system prompts of agents and human workers. Preset factories (like `NewSoftwareCompany`) instantiate a full working organizational chart in memory.

## Quick Start
1. Ensure Bazel is active.
2. Build the module: `bazelisk build //srcs/domain/...`

## Developer Workflow
- Run domain unit tests: `bazelisk test //srcs/domain/...`
- Expand `RoleProfiles` carefully; prompt changes here directly alter the LLM behaviors of instantiated agent workers.

## Configuration
- Controlled via programmatic presets and initial states passed during application launch.
