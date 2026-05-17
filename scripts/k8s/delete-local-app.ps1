$ErrorActionPreference = "Stop"

$Namespace = if ($env:K8S_NAMESPACE) { $env:K8S_NAMESPACE } else { "smart-parking" }
kubectl delete namespace $Namespace --ignore-not-found
