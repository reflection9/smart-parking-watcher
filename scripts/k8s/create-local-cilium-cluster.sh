#!/usr/bin/env bash
set -euo pipefail

CLUSTER_NAME="${CLUSTER_NAME:-spw-local}"
CILIUM_VERSION="${CILIUM_VERSION:-1.19.3}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
CLUSTER_CONFIG="${PROJECT_ROOT}/infra/k8s/local/k3d-cilium-cluster.yaml"
CILIUM_VALUES="${PROJECT_ROOT}/infra/k8s/local/cilium-values.yaml"

require_command() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "Missing required command: $1" >&2
    exit 1
  fi
}

require_command docker
require_command k3d
require_command kubectl
require_command helm

if k3d cluster list | awk 'NR > 1 {print $1}' | grep -qx "${CLUSTER_NAME}"; then
  echo "Cluster ${CLUSTER_NAME} already exists. Reusing it."
else
  echo "Creating k3d cluster ${CLUSTER_NAME}..."
  k3d cluster create --config "${CLUSTER_CONFIG}"
fi

kubectl config use-context "k3d-${CLUSTER_NAME}"

if ! kubectl get nodes >/dev/null 2>&1; then
  echo "kubectl cannot reach cluster k3d-${CLUSTER_NAME}" >&2
  exit 1
fi

echo "Installing Cilium ${CILIUM_VERSION}..."
helm upgrade --install cilium oci://quay.io/cilium/charts/cilium \
  --version "${CILIUM_VERSION}" \
  --namespace kube-system \
  --values "${CILIUM_VALUES}"

echo "Waiting for Cilium..."
kubectl -n kube-system rollout status daemonset/cilium --timeout=300s
kubectl -n kube-system rollout status deployment/cilium-operator --timeout=300s

if command -v cilium >/dev/null 2>&1; then
  cilium status --wait
else
  echo "Cilium CLI is optional and not installed. Skipping 'cilium status --wait'."
fi

echo "Cluster is ready."
kubectl get nodes -o wide
kubectl -n kube-system get pods -l k8s-app=cilium
