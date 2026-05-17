#!/usr/bin/env bash
set -euo pipefail

NAMESPACE="${K8S_NAMESPACE:-smart-parking}"
kubectl delete namespace "${NAMESPACE}" --ignore-not-found
