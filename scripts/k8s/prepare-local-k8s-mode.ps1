$ErrorActionPreference = "Stop"

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$ProjectRoot = Resolve-Path (Join-Path $ScriptDir "../..")
$ClusterName = if ($env:CLUSTER_NAME) { $env:CLUSTER_NAME } else { "spw-local" }

function Require-Command($Name) {
    if (-not (Get-Command $Name -ErrorAction SilentlyContinue)) {
        throw "Missing required command: $Name"
    }
}

Require-Command docker
Require-Command kubectl
Require-Command k3d

Push-Location $ProjectRoot
try {
    Write-Host "Preparing local Kubernetes mode for project '$ProjectRoot'..."
    Write-Host "Docker Desktop must stay running because k3d uses Docker."
    Write-Host "Docker Compose application stack will be stopped if it is running."

    $runningComposeServices = docker compose ps --services --status running 2>$null
    if ($LASTEXITCODE -eq 0 -and $runningComposeServices) {
        Write-Host "Stopping docker compose stack from project root..."
        docker compose down
        if ($LASTEXITCODE -ne 0) {
            throw "docker compose down failed"
        }
    } else {
        Write-Host "docker compose stack is already stopped."
    }

    $clusterExists = (& k3d cluster list | Select-String -Pattern "^$ClusterName\s") -ne $null
    if ($clusterExists) {
        Write-Host "k3d cluster '$ClusterName' exists."
    } else {
        Write-Host "k3d cluster '$ClusterName' does not exist yet."
        Write-Host "Create it with: .\\scripts\\k8s\\create-local-cilium-cluster.ps1"
    }

    Write-Host "Current kubectl context:"
    kubectl config current-context

    Write-Host "Available k3d clusters:"
    k3d cluster list

    Write-Host "Local Kubernetes mode is ready."
    Write-Host "Next steps:"
    Write-Host "  1. .\\scripts\\k8s\\create-local-cilium-cluster.ps1"
    Write-Host "  2. .\\scripts\\k8s\\validate-local-cilium-cluster.ps1"
} finally {
    Pop-Location
}
