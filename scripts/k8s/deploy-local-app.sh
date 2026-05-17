#!/usr/bin/env bash
set -euo pipefail

NAMESPACE="${K8S_NAMESPACE:-smart-parking}"
BUILD_IMAGES="false"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
APP_ROOT="${PROJECT_ROOT}/infra/k8s/app"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --build-images)
      BUILD_IMAGES="true"
      shift
      ;;
    *)
      echo "Unknown argument: $1" >&2
      exit 1
      ;;
  esac
done

if ! command -v kubectl >/dev/null 2>&1; then
  echo "Missing required command: kubectl" >&2
  exit 1
fi

if [[ "${BUILD_IMAGES}" == "true" ]]; then
  "${SCRIPT_DIR}/build-push-local-images.sh"
fi

echo "Applying infrastructure manifests..."
kubectl apply -f "${APP_ROOT}/infra"

echo "Waiting for infrastructure..."
kubectl -n "${NAMESPACE}" rollout status statefulset/postgres --timeout=300s
kubectl -n "${NAMESPACE}" rollout status statefulset/mongo --timeout=300s
kubectl -n "${NAMESPACE}" rollout status deployment/redis --timeout=180s
kubectl -n "${NAMESPACE}" rollout status statefulset/minio --timeout=300s
kubectl -n "${NAMESPACE}" rollout status statefulset/kafka --timeout=420s

echo "Recreating migration jobs..."
MIGRATION_JOBS=(
  user-db-migrate
  parking-db-migrate
  subscription-db-migrate
  reservation-db-migrate
  notification-db-migrate
)

for job in "${MIGRATION_JOBS[@]}"; do
  kubectl -n "${NAMESPACE}" delete job "${job}" --ignore-not-found
done

kubectl apply -f "${APP_ROOT}/migrations"

for job in "${MIGRATION_JOBS[@]}"; do
  kubectl -n "${NAMESPACE}" wait --for=condition=complete "job/${job}" --timeout=240s
done

echo "Applying application manifests..."
kubectl apply -f "${APP_ROOT}/services"

echo "Waiting for application deployments..."
DEPLOYMENTS=(
  user-service
  parking-service
  subscription-service
  reservation-service
  history-service
  notification-service
  gateway
)

for deployment in "${DEPLOYMENTS[@]}"; do
  kubectl -n "${NAMESPACE}" rollout status "deployment/${deployment}" --timeout=300s
done

echo "Application is deployed."
kubectl -n "${NAMESPACE}" get pods -o wide
kubectl -n "${NAMESPACE}" get svc
echo "Gateway inside cluster: http://gateway:8080"
echo "For local browser access: kubectl -n ${NAMESPACE} port-forward svc/gateway 8080:8080"
