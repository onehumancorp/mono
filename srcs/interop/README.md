# Interop Module

## Identity
The `interop` module handles B2B collaboration and ecosystem interoperability between federated One Human Corp clusters.

## Architecture
This module manages Cross-Org Collaboration via SPIFFE/SPIRE federation. It allows AI agents from "Company A" to securely negotiate and exchange data with agents from "Company B" across different Kubernetes clusters by verifying X.509 SVIDs and managing shared virtual meeting rooms.

## Developer Workflow
Testing this module often requires mocked SPIRE servers and multiple simulated trust domains.

- **Test Interop**: `bazel test //srcs/interop/...`
