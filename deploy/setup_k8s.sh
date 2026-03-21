#!/bin/bash
# Automates the K8s context setup for the local developer experience.
set -euo pipefail

echo "--- One Human Corp: Local K8s Setup ---"

# Detect if Docker or minikube is running and switch context
if command -v kind >/dev/null 2>&1 && kind get clusters 2>/dev/null | grep -q "kind"; then
    echo "Found local Kind cluster. Switching context..."
    kubectl config use-context kind-kind
elif command -v kubectl >/dev/null 2>&1 && kubectl config get-contexts -o name 2>/dev/null | grep -q "docker-desktop"; then
    echo "Found docker-desktop context. Switching context..."
    kubectl config use-context docker-desktop
elif command -v kubectl >/dev/null 2>&1 && kubectl config get-contexts -o name 2>/dev/null | grep -q "minikube"; then
    echo "Found minikube context. Switching context..."
    kubectl config use-context minikube
else
    echo "No standard local K8s context (kind, docker-desktop, minikube) found."
    if command -v kind >/dev/null 2>&1; then
        echo "Attempting to create a new Kind cluster..."
        kind create cluster
        kubectl config use-context kind-kind
    else
        echo "Kind is not installed. Please install Kind or ensure you have a local Kubernetes cluster running."
        exit 1
    fi
fi

echo "--- Local K8s Context Configured ---"
echo "You are now using context: $(kubectl config current-context)"
echo ""
echo "Run the application easily using: bazelisk run //:up"
