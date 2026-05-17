#!/usr/bin/env bash
set -euo pipefail

REGISTRY="${LOCAL_REGISTRY:-localhost:5001}"
IMAGE_PREFIX="${IMAGE_PREFIX:-smart-parking}"
IMAGE_TAG="${IMAGE_TAG:-local}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
DOCKERFILE="${PROJECT_ROOT}/infra/docker/go-service.Dockerfile"
SERVICES=(
  user-service
  parking-service
  subscription-service
  reservation-service
  history-service
  notification-service
)

if ! command -v docker >/dev/null 2>&1; then
  echo "Missing required command: docker" >&2
  exit 1
fi

cd "${PROJECT_ROOT}"
for service in "${SERVICES[@]}"; do
  image="${REGISTRY}/${IMAGE_PREFIX}/${service}:${IMAGE_TAG}"
  echo "Building ${image} ..."
  docker build \
    -f "${DOCKERFILE}" \
    --build-arg "SERVICE_NAME=${service}" \
    -t "${image}" \
    .

  echo "Pushing ${image} ..."
  docker push "${image}"
done

echo "All service images were pushed to ${REGISTRY}."
