# Deployment

## Identity
This module contains Kubernetes manifests, Helm charts, and CI/CD pipelines to deploy One Human Corp.

## Architecture
Targets a cloud-native K8s environment, applying strict security policies (read-only root filesystems, minimal egress, SPIFFE/SPIRE).

## Quick Start
Apply manifests using standard kubectl:
`kubectl apply -k ./deploy/overlays/dev`

## Developer Workflow
To test deployment configurations:
`bazelisk test //deploy/...`

## Configuration
Requires a valid Kubernetes cluster, KUBECONFIG, and the presence of SPIRE for identity management.
