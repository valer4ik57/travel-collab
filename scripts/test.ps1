$ErrorActionPreference = "Stop"

$Root = Split-Path -Parent $PSScriptRoot
$Backend = Join-Path $Root "backend"
$Frontend = Join-Path $Root "frontend"

function Run-Step($Title, $ScriptBlock) {
    Write-Host ""
    Write-Host $Title -ForegroundColor Cyan
    & $ScriptBlock
}

Write-Host "Travel-Collab verification" -ForegroundColor Green
Write-Host "Project root: $Root"

Run-Step "Running backend Go tests..." {
    Set-Location $Backend
    go test ./...
}

Run-Step "Installing frontend dependencies..." {
    Set-Location $Frontend
    npm install
}

Run-Step "Building frontend..." {
    Set-Location $Frontend
    npm run build
}

Set-Location $Root
Write-Host ""
Write-Host "All checks passed." -ForegroundColor Green
