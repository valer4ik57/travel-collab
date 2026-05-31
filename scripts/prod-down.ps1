param(
    [switch]$Volumes
)

$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $PSScriptRoot
Set-Location $root

if (-not (Test-Path ".env.production")) {
    Write-Host ".env.production not found. Nothing to stop with production env." -ForegroundColor Yellow
    exit 0
}

if ($Volumes) {
    docker compose --env-file .env.production -f docker-compose.prod.yml down -v
} else {
    docker compose --env-file .env.production -f docker-compose.prod.yml down
}
