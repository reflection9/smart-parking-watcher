$ErrorActionPreference = "Stop"

$Registry = if ($env:CI_REGISTRY) { $env:CI_REGISTRY } else { "spw-registry:5000" }
$ImagePrefix = if ($env:IMAGE_PREFIX) { $env:IMAGE_PREFIX } else { "smart-parking" }
$ImageTag = if ($env:IMAGE_TAG) { $env:IMAGE_TAG } else { "local" }
$KanikoImage = if ($env:KANIKO_IMAGE) { $env:KANIKO_IMAGE } else { "gcr.io/kaniko-project/executor:v1.23.2-debug" }
$DockerNetwork = if ($env:DOCKER_NETWORK) { $env:DOCKER_NETWORK } else { "k3d-spw-local" }
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$ProjectRoot = Resolve-Path (Join-Path $ScriptDir "../..")
$Services = @(
    "user-service",
    "parking-service",
    "subscription-service",
    "reservation-service",
    "history-service",
    "notification-service"
)

if (-not (Get-Command docker -ErrorAction SilentlyContinue)) {
    throw "Missing required command: docker"
}

foreach ($Service in $Services) {
    $Destination = "$Registry/$ImagePrefix/$Service`:$ImageTag"
    Write-Host "Building and pushing $Destination with Kaniko..."

    docker run --rm `
        --network $DockerNetwork `
        -v "${ProjectRoot}:/workspace" `
        $KanikoImage `
        --context=dir:///workspace `
        --dockerfile=infra/docker/go-service.Dockerfile `
        --build-arg=SERVICE_NAME=$Service `
        --destination=$Destination `
        --insecure `
        --skip-tls-verify `
        --insecure-pull `
        --skip-tls-verify-pull

    if ($LASTEXITCODE -ne 0) {
        throw "Kaniko build failed for $Service"
    }
}

Write-Host "All service images were pushed to $Registry."
