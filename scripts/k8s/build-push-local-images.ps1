$ErrorActionPreference = "Stop"

$Registry = if ($env:LOCAL_REGISTRY) { $env:LOCAL_REGISTRY } else { "localhost:5001" }
$ImagePrefix = if ($env:IMAGE_PREFIX) { $env:IMAGE_PREFIX } else { "smart-parking" }
$ImageTag = if ($env:IMAGE_TAG) { $env:IMAGE_TAG } else { "local" }
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$ProjectRoot = Resolve-Path (Join-Path $ScriptDir "../..")
$Dockerfile = Join-Path $ProjectRoot "infra/docker/go-service.Dockerfile"
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

Push-Location $ProjectRoot
try {
    foreach ($Service in $Services) {
        $Image = "$Registry/$ImagePrefix/$Service`:$ImageTag"
        Write-Host "Building $Image ..."
        docker build `
            -f $Dockerfile `
            --build-arg SERVICE_NAME=$Service `
            -t $Image `
            .

        if ($LASTEXITCODE -ne 0) {
            throw "docker build failed for $Service"
        }

        Write-Host "Pushing $Image ..."
        docker push $Image

        if ($LASTEXITCODE -ne 0) {
            throw "docker push failed for $Service"
        }
    }
} finally {
    Pop-Location
}

Write-Host "All service images were pushed to $Registry."
