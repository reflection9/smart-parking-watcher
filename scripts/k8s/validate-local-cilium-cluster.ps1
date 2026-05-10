$ErrorActionPreference = "Stop"

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$ProjectRoot = Resolve-Path (Join-Path $ScriptDir "../..")
$SmokeManifest = Join-Path $ProjectRoot "infra/k8s/local/cilium-policy-smoke.yaml"

kubectl get nodes
kubectl -n kube-system get pods -l k8s-app=cilium
kubectl -n kube-system get deployment cilium-operator

kubectl apply -f $SmokeManifest
kubectl -n spw-smoke rollout status deployment/echo-server --timeout=180s
kubectl -n spw-smoke rollout status deployment/client-allowed --timeout=180s
kubectl -n spw-smoke rollout status deployment/client-blocked --timeout=180s

Write-Host "Checking allowed client -> echo-server..."
& kubectl -n spw-smoke exec deploy/client-allowed -- curl -sS --max-time 5 http://echo-server | Out-Null

if ($LASTEXITCODE -ne 0) {
    throw "Allowed client could not reach echo-server"
}

Write-Host "Checking blocked client -> echo-server. This request should fail by timeout..."
& kubectl -n spw-smoke exec deploy/client-blocked -- curl -sS --max-time 5 http://echo-server | Out-Null

if ($LASTEXITCODE -eq 0) {
    throw "Blocked client reached echo-server, Cilium policy did not block traffic"
}

Write-Host "Blocked client request failed as expected."
Write-Host "Cilium network policy smoke test passed."
Write-Host "Cleanup: kubectl delete namespace spw-smoke"