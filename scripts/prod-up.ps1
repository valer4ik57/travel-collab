$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $PSScriptRoot
$envFile = Join-Path $root ".env.production"

if (!(Test-Path $envFile)) {
    Write-Host ".env.production not found. Creating local demo env..." -ForegroundColor Yellow
    & (Join-Path $PSScriptRoot "init-production-env.ps1") -Local
}

Set-Location $root
docker compose --env-file .env.production -f docker-compose.prod.yml up --build
