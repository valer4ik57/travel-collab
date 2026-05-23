$ErrorActionPreference = "Stop"

$Root = Split-Path -Parent $PSScriptRoot
$Backend = Join-Path $Root "backend"
$Frontend = Join-Path $Root "frontend"

Write-Host ""
Write-Host "Starting Travel-Collab dev environment..." -ForegroundColor Cyan
Write-Host "Project root: $Root"
Write-Host ""

$NetworkIPv4 = $null
try {
    $NetworkIPv4 = Get-NetIPAddress -AddressFamily IPv4 |
        Where-Object { $_.IPAddress -like '192.168.*' -or $_.IPAddress -like '10.*' -or $_.IPAddress -like '172.16.*' -or $_.IPAddress -like '172.17.*' -or $_.IPAddress -like '172.18.*' -or $_.IPAddress -like '172.19.*' -or $_.IPAddress -like '172.20.*' -or $_.IPAddress -like '172.21.*' -or $_.IPAddress -like '172.22.*' -or $_.IPAddress -like '172.23.*' -or $_.IPAddress -like '172.24.*' -or $_.IPAddress -like '172.25.*' -or $_.IPAddress -like '172.26.*' -or $_.IPAddress -like '172.27.*' -or $_.IPAddress -like '172.28.*' -or $_.IPAddress -like '172.29.*' -or $_.IPAddress -like '172.30.*' -or $_.IPAddress -like '172.31.*' } |
        Where-Object { $_.InterfaceAlias -notmatch 'WSL|Docker|vEthernet|Loopback|Bluetooth|happ' } |
        Select-Object -First 1 -ExpandProperty IPAddress
} catch {
    $NetworkIPv4 = $null
}


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
if ($NetworkIPv4) {
    Write-Host "Frontend mobile web: http://$($NetworkIPv4):5173" -ForegroundColor Cyan
    Write-Host "Backend health LAN:  http://$($NetworkIPv4):8080/api/v1/health" -ForegroundColor Cyan
}
Write-Host "Backend:  http://localhost:8080" -ForegroundColor Cyan
Write-Host ""
Write-Host "If the browser does not open automatically, open the Frontend link manually." -ForegroundColor Gray
Write-Host ""

Start-Process "http://localhost:5173"