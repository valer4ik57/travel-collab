$ErrorActionPreference = "Stop"

$Frontend = Split-Path -Parent $PSScriptRoot
Set-Location $Frontend

function Invoke-Checked {
    param(
        [Parameter(Mandatory = $true)]
        [string]$Command,
        [Parameter(ValueFromRemainingArguments = $true)]
        [string[]]$Arguments
    )

    & $Command @Arguments
    if ($LASTEXITCODE -ne 0) {
        throw "Command failed with exit code ${LASTEXITCODE}: $Command $($Arguments -join ' ')"
    }
}

function Add-ToPathIfExists {
    param([string]$PathToAdd)
    if ($PathToAdd -and (Test-Path $PathToAdd) -and ($env:Path -notlike "*$PathToAdd*")) {
        $env:Path = "$PathToAdd;$env:Path"
    }
}

function Has-JavaHome {
    param([string]$Candidate)
    return ($Candidate -and (Test-Path (Join-Path $Candidate "bin\java.exe")))
}

if (-not $env:ANDROID_HOME) {
    $defaultSdk = Join-Path $env:LOCALAPPDATA "Android\Sdk"
    if (Test-Path $defaultSdk) {
        $env:ANDROID_HOME = $defaultSdk
    }
}

if ($env:ANDROID_HOME) {
    Add-ToPathIfExists (Join-Path $env:ANDROID_HOME "platform-tools")
}

if (-not (Has-JavaHome $env:JAVA_HOME)) {
    $studioJbr = "C:\Program Files\Android\Android Studio\jbr"
    if (Has-JavaHome $studioJbr) {
        $env:JAVA_HOME = $studioJbr
    }
}

if ($env:JAVA_HOME) {
    Add-ToPathIfExists (Join-Path $env:JAVA_HOME "bin")
}

Write-Host "Installing frontend and Capacitor dependencies..." -ForegroundColor Cyan
Invoke-Checked npm install

if (!(Test-Path "android")) {
    Write-Host "Adding Android platform..." -ForegroundColor Cyan
    Invoke-Checked npx cap add android
} else {
    Write-Host "Android platform already exists." -ForegroundColor Yellow
}

Write-Host "Building web assets..." -ForegroundColor Cyan
Invoke-Checked npm run build

Write-Host "Syncing Capacitor Android project..." -ForegroundColor Cyan
Invoke-Checked npx cap sync android

Write-Host "Mobile project is ready." -ForegroundColor Green
