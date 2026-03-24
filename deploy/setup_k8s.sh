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
    if ! command -v kind >/dev/null 2>&1; then
        echo "Kind is not installed. Attempting to install Kind..."
        # Install kind into a local bin directory if not present
        if [[ "$OSTYPE" == "linux-gnu"* ]]; then
            curl -Lo ./kind "https://kind.sigs.k8s.io/dl/v0.22.0/kind-linux-amd64"
            chmod +x ./kind
            sudo mv ./kind /usr/local/bin/kind
        elif [[ "$OSTYPE" == "darwin"* ]]; then
            # Assuming macos with arm or amd64
            ARCH=$(uname -m)
            if [ "$ARCH" = "arm64" ]; then
                curl -Lo ./kind "https://kind.sigs.k8s.io/dl/v0.22.0/kind-darwin-arm64"
            else
                curl -Lo ./kind "https://kind.sigs.k8s.io/dl/v0.22.0/kind-darwin-amd64"
            fi
            chmod +x ./kind
            sudo mv ./kind /usr/local/bin/kind
        else
            echo "Unsupported OS for automatic Kind installation. Please install Kind manually."
            exit 1
        fi
    fi

    echo "Attempting to create a new Kind cluster..."
    kind create cluster
    kubectl config use-context kind-kind
fi

echo "--- Local K8s Context Configured ---"
echo "You are now using context: $(kubectl config current-context)"
echo ""
echo "Run the application easily using: bazelisk run //:up"
