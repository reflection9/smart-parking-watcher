$ErrorActionPreference = "Stop"

$Namespace = if ($env:K8S_NAMESPACE) { $env:K8S_NAMESPACE } else { "smart-parking" }
$Email = "smoke-$([DateTimeOffset]::UtcNow.ToUnixTimeSeconds())@smartparking.local"
$RegisterPayload = "{""name"":""Smoke User"",""email"":""$Email"",""password"":""password123""}"

kubectl -n $Namespace get pods

Write-Host "Checking gateway health from inside cluster..."
kubectl -n $Namespace run spw-gateway-smoke --rm -i --restart=Never --image=curlimages/curl:8.10.1 -- `
    curl -fsS http://gateway:8080/health

if ($LASTEXITCODE -ne 0) {
    throw "Gateway health smoke check failed"
}

Write-Host "Checking user registration through gateway..."
kubectl -n $Namespace run spw-register-smoke --rm -i --restart=Never --image=curlimages/curl:8.10.1 -- `
    curl -fsS -X POST -H "Content-Type: application/json" -d $RegisterPayload http://gateway:8080/users/register

if ($LASTEXITCODE -ne 0) {
    throw "User registration smoke check failed"
}

Write-Host "Kubernetes application smoke test passed. Registered test email: $Email"
