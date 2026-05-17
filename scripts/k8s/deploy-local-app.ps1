param(
    [switch]$BuildImages
)

$ErrorActionPreference = "Stop"

$Namespace = if ($env:K8S_NAMESPACE) { $env:K8S_NAMESPACE } else { "smart-parking" }
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$ProjectRoot = Resolve-Path (Join-Path $ScriptDir "../..")
$AppRoot = Join-Path $ProjectRoot "infra/k8s/app"

function Require-Command($Name) {
    if (-not (Get-Command $Name -ErrorAction SilentlyContinue)) {
        throw "Missing required command: $Name"
    }
}

Require-Command kubectl

if ($BuildImages) {
    & (Join-Path $ScriptDir "build-push-local-images.ps1")
}

Write-Host "Applying infrastructure manifests..."
kubectl apply -f (Join-Path $AppRoot "infra")

Write-Host "Waiting for infrastructure..."
kubectl -n $Namespace rollout status statefulset/postgres --timeout=300s
kubectl -n $Namespace rollout status statefulset/mongo --timeout=300s
kubectl -n $Namespace rollout status deployment/redis --timeout=180s
kubectl -n $Namespace rollout status statefulset/minio --timeout=300s
kubectl -n $Namespace rollout status statefulset/kafka --timeout=420s

Write-Host "Recreating migration jobs..."
$MigrationJobs = @(
    "user-db-migrate",
    "parking-db-migrate",
    "subscription-db-migrate",
    "reservation-db-migrate",
    "notification-db-migrate"
)

foreach ($Job in $MigrationJobs) {
    kubectl -n $Namespace delete job $Job --ignore-not-found
}

kubectl apply -f (Join-Path $AppRoot "migrations")

foreach ($Job in $MigrationJobs) {
    kubectl -n $Namespace wait --for=condition=complete job/$Job --timeout=240s
}

Write-Host "Applying application manifests..."
kubectl apply -f (Join-Path $AppRoot "services")

Write-Host "Waiting for application deployments..."
$Deployments = @(
    "user-service",
    "parking-service",
    "subscription-service",
    "reservation-service",
    "history-service",
    "notification-service",
    "gateway"
)

foreach ($Deployment in $Deployments) {
    kubectl -n $Namespace rollout status deployment/$Deployment --timeout=300s
}

Write-Host "Application is deployed."
kubectl -n $Namespace get pods -o wide
kubectl -n $Namespace get svc
Write-Host "Gateway inside cluster: http://gateway:8080"
Write-Host "For local browser access: kubectl -n $Namespace port-forward svc/gateway 8080:8080"
