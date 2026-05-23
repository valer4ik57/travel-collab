$ErrorActionPreference = "Stop"

$Root = Split-Path -Parent $PSScriptRoot
$Backend = Join-Path $Root "backend"
$Frontend = Join-Path $Root "frontend"

Write-Host ""
Write-Host "Starting Travel-Collab dev environment..." -ForegroundColor Cyan
Write-Host "Project root: $Root"
Write-Host ""

Set-Location $Root

Write-Host "Starting PostgreSQL/PostGIS through Docker Compose..." -ForegroundColor Yellow
docker compose up -d

Write-Host ""
Write-Host "Waiting for PostgreSQL container health..." -ForegroundColor Yellow

$maxAttempts = 30
$attempt = 0

do {
    Start-Sleep -Seconds 1
    $attempt++

    $status = docker inspect --format='{{.State.Health.Status}}' travel_collab_postgres 2>$null

    if ($status -eq "healthy") {
        Write-Host "PostgreSQL is healthy." -ForegroundColor Green
        break
    }

    Write-Host "PostgreSQL status: $status ($attempt/$maxAttempts)"
} while ($attempt -lt $maxAttempts)

if ($status -ne "healthy") {
    Write-Host "PostgreSQL did not become healthy in time. Check Docker Desktop." -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "Starting backend in a separate PowerShell window..." -ForegroundColor Yellow

$backendCommand = @"
cd `"$Backend`"
`$env:DATABASE_URL = "postgres://postgres:postgres@127.0.0.1:5433/travel_collab?sslmode=disable"
go run ./cmd/main.go
"@

Start-Process powershell -ArgumentList "-NoExit", "-Command", $backendCommand

Write-Host "Starting frontend in a separate PowerShell window..." -ForegroundColor Yellow

$frontendCommand = @"
cd `"$Frontend`"
if (!(Test-Path "node_modules")) {
    npm install
}
npm run dev
"@

Start-Process powershell -ArgumentList "-NoExit", "-Command", $frontendCommand

Write-Host ""
Write-Host "Travel-Collab is starting." -ForegroundColor Green
Write-Host ""
Write-Host "Frontend: http://localhost:5173" -ForegroundColor Cyan
Write-Host "Backend:  http://localhost:8080" -ForegroundColor Cyan
Write-Host ""
Write-Host "If the browser does not open automatically, open the Frontend link manually." -ForegroundColor Gray
Write-Host ""

Start-Process "http://localhost:5173"