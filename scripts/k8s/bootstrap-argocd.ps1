param(
    [string]$RepoUrl,
    [string]$TargetRevision
)

$ErrorActionPreference = "Stop"

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$ProjectRoot = Resolve-Path (Join-Path $ScriptDir "../..")
$TemplatePath = Join-Path $ProjectRoot "infra/gitops/argocd/bootstrap/root-application.yaml.tpl"
$RenderedPath = Join-Path ([System.IO.Path]::GetTempPath()) "smart-parking-root-application.rendered.yaml"

function Require-Command($Name) {
    if (-not (Get-Command $Name -ErrorAction SilentlyContinue)) {
        throw "Missing required command: $Name"
    }
}

Require-Command git
Require-Command kubectl

if (-not $RepoUrl) {
    $RepoUrl = (git -C $ProjectRoot remote get-url origin).Trim()
}

if (-not $TargetRevision) {
    $TargetRevision = (git -C $ProjectRoot branch --show-current).Trim()
}

$template = Get-Content $TemplatePath -Raw
$rendered = $template.Replace("__REPO_URL__", $RepoUrl).Replace("__TARGET_REVISION__", $TargetRevision)
Set-Content -Path $RenderedPath -Value $rendered -Encoding ascii

kubectl create namespace argocd --dry-run=client -o yaml | kubectl apply -f -
kubectl apply -n argocd --server-side --force-conflicts -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml
kubectl rollout status deployment/argocd-server -n argocd --timeout=300s
kubectl apply -f $RenderedPath

Write-Host "Argo CD bootstrap completed."
Write-Host "Repo URL: $RepoUrl"
Write-Host "Target revision: $TargetRevision"
