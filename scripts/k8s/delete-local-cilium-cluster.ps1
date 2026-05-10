$ErrorActionPreference = "Stop"

$ClusterName = if ($env:CLUSTER_NAME) { $env:CLUSTER_NAME } else { "spw-local" }

if (-not (Get-Command k3d -ErrorAction SilentlyContinue)) {
    throw "Missing required command: k3d"
}

k3d cluster delete $ClusterName
