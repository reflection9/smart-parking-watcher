#!/usr/bin/env bash
set -euo pipefail

NAMESPACE="${K8S_NAMESPACE:-smart-parking}"
EMAIL="smoke-$(date +%s)@smartparking.local"
REGISTER_PAYLOAD="{\"name\":\"Smoke User\",\"email\":\"${EMAIL}\",\"password\":\"password123\"}"

kubectl -n "${NAMESPACE}" get pods

echo "Checking gateway health from inside cluster..."
kubectl -n "${NAMESPACE}" run spw-gateway-smoke --rm -i --restart=Never --image=curlimages/curl:8.10.1 -- \
  curl -fsS http://gateway:8080/health

echo "Checking user registration through gateway..."
kubectl -n "${NAMESPACE}" run spw-register-smoke --rm -i --restart=Never --image=curlimages/curl:8.10.1 -- \
  curl -fsS -X POST -H "Content-Type: application/json" -d "${REGISTER_PAYLOAD}" http://gateway:8080/users/register

echo "Kubernetes application smoke test passed. Registered test email: ${EMAIL}"
