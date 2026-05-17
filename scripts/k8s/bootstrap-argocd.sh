#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
TEMPLATE_PATH="$PROJECT_ROOT/infra/gitops/argocd/bootstrap/root-application.yaml.tpl"
RENDERED_PATH="${TMPDIR:-/tmp}/smart-parking-root-application.rendered.yaml"

require_command() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "Missing required command: $1" >&2
    exit 1
  fi
}

require_command git
require_command kubectl

REPO_URL="${1:-$(git -C "$PROJECT_ROOT" remote get-url origin)}"
TARGET_REVISION="${2:-$(git -C "$PROJECT_ROOT" branch --show-current)}"

sed \
  -e "s|__REPO_URL__|$REPO_URL|g" \
  -e "s|__TARGET_REVISION__|$TARGET_REVISION|g" \
  "$TEMPLATE_PATH" > "$RENDERED_PATH"

kubectl create namespace argocd --dry-run=client -o yaml | kubectl apply -f -
kubectl apply -n argocd --server-side --force-conflicts -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml
kubectl rollout status deployment/argocd-server -n argocd --timeout=300s
kubectl apply -f "$RENDERED_PATH"

echo "Argo CD bootstrap completed."
echo "Repo URL: $REPO_URL"
echo "Target revision: $TARGET_REVISION"
