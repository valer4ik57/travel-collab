$ErrorActionPreference = "Stop"

$Frontend = Split-Path -Parent $PSScriptRoot
Set-Location $Frontend

if (!(Test-Path "android")) {
    Write-Host "Android platform is missing. Running mobile init first..." -ForegroundColor Yellow
    .\scripts\mobile-init.ps1
}

npx cap open android
