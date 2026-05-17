param(
    [Parameter(Mandatory = $true)]
    [string]$ImageTag,
    [string]$ValuesPath
)

$ErrorActionPreference = "Stop"

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$ProjectRoot = Resolve-Path (Join-Path $ScriptDir "../..")

if (-not $ValuesPath) {
    $ValuesPath = Join-Path $ProjectRoot "helm/smart-parking/values.yaml"
}

$lines = Get-Content $ValuesPath
$insideGlobal = $false
$insideImage = $false
$updated = $false

for ($i = 0; $i -lt $lines.Count; $i++) {
    $line = $lines[$i]

    if ($line -match '^global:\s*$') {
        $insideGlobal = $true
        $insideImage = $false
        continue
    }

    if ($insideGlobal -and $line -match '^\S') {
        $insideGlobal = $false
        $insideImage = $false
    }

    if ($insideGlobal -and $line -match '^\s{2}image:\s*$') {
        $insideImage = $true
        continue
    }

    if ($insideImage -and $line -match '^\s{2}\S') {
        $insideImage = $false
    }

    if ($insideGlobal -and $insideImage -and $line -match '^\s{4}tag:\s*') {
        $lines[$i] = "    tag: $ImageTag"
        $updated = $true
        break
    }
}

if (-not $updated) {
    throw "Could not update global.image.tag in $ValuesPath"
}

Set-Content -Path $ValuesPath -Value $lines -Encoding ascii
Write-Host "Updated Helm values tag to $ImageTag."
