param(
    [string]$AppName = "smart-parking-services",
    [string]$Namespace = "argocd"
)

$ErrorActionPreference = "Stop"

if (-not (Get-Command kubectl -ErrorAction SilentlyContinue)) {
    throw "Missing required command: kubectl"
}

kubectl annotate application $AppName -n $Namespace argocd.argoproj.io/refresh=hard --overwrite

if ($LASTEXITCODE -ne 0) {
    throw "Failed to refresh Argo CD application $AppName"
}

Write-Host "Requested hard refresh for Argo CD application $AppName."
