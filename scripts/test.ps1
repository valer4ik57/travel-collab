param(
    [switch]$SkipFrontendInstall
)

$ErrorActionPreference = "Stop"

$Root = Split-Path -Parent $PSScriptRoot
$Backend = Join-Path $Root "backend"
$Frontend = Join-Path $Root "frontend"

function Run-Step($Title, $ScriptBlock) {
    Write-Host ""
    Write-Host $Title -ForegroundColor Cyan
    & $ScriptBlock
}

$OriginalLocation = Get-Location
try {
    Write-Host "Travel-Collab verification" -ForegroundColor Green
    Write-Host "Project root: $Root"

    Run-Step "Running backend Go tests..." {
        Push-Location $Backend
        try {
            go test ./...
        } finally {
            Pop-Location
        }
    }

    if (-not $SkipFrontendInstall) {
        Run-Step "Installing frontend dependencies with npm ci..." {
            Push-Location $Frontend
            try {
                npm ci
            } finally {
                Pop-Location
            }
        }
    } else {
        Write-Host ""
        Write-Host "Skipping frontend dependency installation." -ForegroundColor Yellow
    }

    Run-Step "Building frontend..." {
        Push-Location $Frontend
        try {
            npm run build
        } finally {
            Pop-Location
        }
    }

    Write-Host ""
    Write-Host "All checks passed." -ForegroundColor Green
} finally {
    Set-Location $OriginalLocation
}
