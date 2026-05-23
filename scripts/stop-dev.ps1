$ErrorActionPreference = "Continue"

$Root = Split-Path -Parent $PSScriptRoot
Set-Location $Root

Write-Host "Stopping backend/frontend processes..." -ForegroundColor Yellow

Get-Process go -ErrorAction SilentlyContinue | Stop-Process -Force
Get-Process node -ErrorAction SilentlyContinue | Stop-Process -Force

Write-Host "Stopping Docker Compose services..." -ForegroundColor Yellow
docker compose stop

Write-Host "Travel-Collab dev environment stopped." -ForegroundColor Green