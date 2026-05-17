#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
CLUSTER_NAME="${CLUSTER_NAME:-spw-local}"

require_command() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "Missing required command: $1" >&2
    exit 1
  fi
}

require_command docker
require_command kubectl
require_command k3d

cd "$PROJECT_ROOT"

echo "Preparing local Kubernetes mode for project '$PROJECT_ROOT'..."
echo "Docker Desktop must stay running because k3d uses Docker."
echo "Docker Compose application stack will be stopped if it is running."

if running_compose_services="$(docker compose ps --services --status running 2>/dev/null)" && [[ -n "$running_compose_services" ]]; then
  echo "Stopping docker compose stack from project root..."
  docker compose down
else
  echo "docker compose stack is already stopped."
fi

if k3d cluster list | grep -q "^${CLUSTER_NAME}[[:space:]]"; then
  echo "k3d cluster '$CLUSTER_NAME' exists."
else
  echo "k3d cluster '$CLUSTER_NAME' does not exist yet."
  echo "Create it with: ./scripts/k8s/create-local-cilium-cluster.sh"
fi

echo "Current kubectl context:"
kubectl config current-context

echo "Available k3d clusters:"
k3d cluster list

echo "Local Kubernetes mode is ready."
echo "Next steps:"
echo "  1. ./scripts/k8s/create-local-cilium-cluster.sh"
echo "  2. ./scripts/k8s/validate-local-cilium-cluster.sh"
