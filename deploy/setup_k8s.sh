#!/bin/bash
# Automates the K8s context setup for the local developer experience.
set -euo pipefail

echo "--- One Human Corp: Local K8s Setup ---"

# Detect if Docker or minikube is running and switch context
if command -v kind >/dev/null 2>&1 && kind get clusters | grep -q "kind"; then
    echo "Found local Kind cluster. Switching context..."
    kubectl config use-context kind-kind
elif kubectl config get-contexts -o name | grep -q "docker-desktop"; then
    echo "Found docker-desktop context. Switching context..."
    kubectl config use-context docker-desktop
elif kubectl config get-contexts -o name | grep -q "minikube"; then
    echo "Found minikube context. Switching context..."
    kubectl config use-context minikube
else
    echo "No standard local K8s context (kind, docker-desktop, minikube) found."
    echo "Please ensure you have a local Kubernetes cluster running."
    return 1 2>/dev/null || exit 1
fi

echo "--- Local K8s Context Configured ---"
echo "You are now using context: $(kubectl config current-context)"
echo ""
echo "Run the application easily using: bazelisk run //:up"
