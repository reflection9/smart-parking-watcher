#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
SMOKE_MANIFEST="${PROJECT_ROOT}/infra/k8s/local/cilium-policy-smoke.yaml"

kubectl get nodes
kubectl -n kube-system get pods -l k8s-app=cilium
kubectl -n kube-system get deployment cilium-operator

kubectl apply -f "${SMOKE_MANIFEST}"
kubectl -n spw-smoke rollout status deployment/echo-server --timeout=180s
kubectl -n spw-smoke rollout status deployment/client-allowed --timeout=180s
kubectl -n spw-smoke rollout status deployment/client-blocked --timeout=180s

echo "Checking allowed client -> echo-server..."
kubectl -n spw-smoke exec deploy/client-allowed -- curl -sS --max-time 5 http://echo-server >/dev/null

echo "Checking blocked client -> echo-server. This request should fail by timeout..."
if kubectl -n spw-smoke exec deploy/client-blocked -- curl -sS --max-time 5 http://echo-server >/dev/null; then
  echo "ERROR: blocked client reached echo-server, Cilium policy did not block traffic" >&2
  exit 1
fi

echo "Cilium network policy smoke test passed."
echo "Cleanup: kubectl delete namespace spw-smoke"
