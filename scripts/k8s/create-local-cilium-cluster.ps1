$ErrorActionPreference = "Stop"

$ClusterName = if ($env:CLUSTER_NAME) { $env:CLUSTER_NAME } else { "spw-local" }
$CiliumVersion = if ($env:CILIUM_VERSION) { $env:CILIUM_VERSION } else { "1.19.3" }
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$ProjectRoot = Resolve-Path (Join-Path $ScriptDir "../..")
$ClusterConfig = Join-Path $ProjectRoot "infra/k8s/local/k3d-cilium-cluster.yaml"
$CiliumValues = Join-Path $ProjectRoot "infra/k8s/local/cilium-values.yaml"

function Require-Command($Name) {
    if (-not (Get-Command $Name -ErrorAction SilentlyContinue)) {
        throw "Missing required command: $Name"
    }
}

Require-Command docker
Require-Command k3d
Require-Command kubectl
Require-Command helm

$clusterExists = (& k3d cluster list | Select-String -Pattern "^$ClusterName\s") -ne $null

if ($clusterExists) {
    Write-Host "Cluster $ClusterName already exists. Reusing it."
} else {
    Write-Host "Creating k3d cluster $ClusterName..."
    k3d cluster create --config $ClusterConfig
}

kubectl config use-context "k3d-$ClusterName"

Write-Host "Installing Cilium $CiliumVersion..."
helm upgrade --install cilium oci://quay.io/cilium/charts/cilium `
    --version $CiliumVersion `
    --namespace kube-system `
    --values $CiliumValues

Write-Host "Waiting for Cilium..."
kubectl -n kube-system rollout status daemonset/cilium --timeout=300s
kubectl -n kube-system rollout status deployment/cilium-operator --timeout=300s

if (Get-Command cilium -ErrorAction SilentlyContinue) {
    cilium status --wait
} else {
    Write-Host "Cilium CLI is optional and not installed. Skipping 'cilium status --wait'."
}

Write-Host "Cluster is ready."
kubectl get nodes -o wide
kubectl -n kube-system get pods -l k8s-app=cilium
