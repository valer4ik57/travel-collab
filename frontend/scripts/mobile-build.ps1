param(
    [ValidateSet("Production", "Local", "Custom")]
    [string]$Mode = "Production",

    [string]$BackendHost = "",
    [string]$BackendPort = "8080",

    [string]$ApiBaseUrl = "https://travel-collab.ru/api/v1",
    [string]$WsBaseUrl = "wss://travel-collab.ru/api/v1/ws",

    [switch]$Install
)

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

function Ensure-AndroidSdk {
    if (-not $env:ANDROID_HOME) {
        $defaultSdk = Join-Path $env:LOCALAPPDATA "Android\Sdk"
        if (Test-Path $defaultSdk) {
            $env:ANDROID_HOME = $defaultSdk
        }
    }

    if (-not $env:ANDROID_HOME) {
        throw "ANDROID_HOME is not set. Expected Android SDK, for example: C:\Users\valer\AppData\Local\Android\Sdk"
    }

    Add-ToPathIfExists (Join-Path $env:ANDROID_HOME "platform-tools")
}

function Ensure-Java {
    if (Has-JavaHome $env:JAVA_HOME) {
        Add-ToPathIfExists (Join-Path $env:JAVA_HOME "bin")
        return
    }

    $candidates = @()
    if ($env:ProgramFiles) {
        $candidates += (Join-Path $env:ProgramFiles "Android\Android Studio\jbr")
        $candidates += (Join-Path $env:ProgramFiles "Android\Android Studio\jre")
    }
    if (${env:ProgramFiles(x86)}) {
        $candidates += (Join-Path ${env:ProgramFiles(x86)} "Android\Android Studio\jbr")
        $candidates += (Join-Path ${env:ProgramFiles(x86)} "Android\Android Studio\jre")
    }
    if ($env:LOCALAPPDATA) {
        $candidates += (Join-Path $env:LOCALAPPDATA "Programs\Android Studio\jbr")
        $candidates += (Join-Path $env:LOCALAPPDATA "Programs\Android Studio\jre")
    }

    foreach ($candidate in $candidates) {
        if (Has-JavaHome $candidate) {
            $env:JAVA_HOME = $candidate
            Add-ToPathIfExists (Join-Path $env:JAVA_HOME "bin")
            return
        }
    }

    $java = Get-Command java -ErrorAction SilentlyContinue
    if ($java) {
        return
    }

    throw "Java was not found. Install Android Studio or set JAVA_HOME to Android Studio JBR, usually: C:\Program Files\Android\Android Studio\jbr"
}

function Get-LocalBackendHost {
    try {
        $candidate = Get-NetIPConfiguration |
            Where-Object { $_.IPv4DefaultGateway -and $_.NetAdapter.Status -eq "Up" } |
            Select-Object -First 1
        if ($candidate -and $candidate.IPv4Address.IPAddress) {
            return $candidate.IPv4Address.IPAddress
        }
    } catch {
        return ""
    }
    return ""
}

if ($Mode -eq "Local") {
    if (-not $BackendHost) {
        $BackendHost = Get-LocalBackendHost
    }

    if (-not $BackendHost) {
        Write-Host "Could not detect local IP automatically." -ForegroundColor Red
        Write-Host "Run for example: powershell -ExecutionPolicy Bypass -File .\scripts\mobile-build.ps1 -Mode Local -BackendHost 192.168.1.166" -ForegroundColor Yellow
        exit 1
    }

    $ApiBaseUrl = "http://${BackendHost}:${BackendPort}/api/v1"
    $WsBaseUrl = "ws://${BackendHost}:${BackendPort}/api/v1/ws"
}

if ($Mode -eq "Custom") {
    if (-not $ApiBaseUrl -or -not $WsBaseUrl) {
        throw "For -Mode Custom set both -ApiBaseUrl and -WsBaseUrl."
    }
}

Ensure-AndroidSdk
Ensure-Java

$env:VITE_API_BASE_URL = $ApiBaseUrl.TrimEnd('/')
$env:VITE_WS_BASE_URL = $WsBaseUrl.TrimEnd('/')

Write-Host "Travel-Collab Android build" -ForegroundColor Cyan
Write-Host "Mode:        $Mode" -ForegroundColor Yellow
Write-Host "Backend API: $env:VITE_API_BASE_URL" -ForegroundColor Yellow
Write-Host "Backend WS:  $env:VITE_WS_BASE_URL" -ForegroundColor Yellow
Write-Host "Android SDK: $env:ANDROID_HOME" -ForegroundColor Gray
Write-Host "JAVA_HOME:   $env:JAVA_HOME" -ForegroundColor Gray
Write-Host ""

Write-Host "Installing frontend and Capacitor dependencies..." -ForegroundColor Cyan
if (Test-Path "package-lock.json") {
    Invoke-Checked npm ci
} else {
    Invoke-Checked npm install
}

Write-Host "Building web assets for mobile..." -ForegroundColor Cyan
Invoke-Checked npm run build

if (!(Test-Path "android")) {
    Write-Host "Adding Android platform..." -ForegroundColor Cyan
    Invoke-Checked npx cap add android
}

Write-Host "Syncing web assets to Android..." -ForegroundColor Cyan
Invoke-Checked npx cap sync android

Write-Host "Building debug APK..." -ForegroundColor Cyan
Push-Location "android"
try {
    Invoke-Checked ".\gradlew.bat" assembleDebug
} finally {
    Pop-Location
}

$apk = Join-Path $Frontend "android\app\build\outputs\apk\debug\app-debug.apk"
if (Test-Path $apk) {
    Write-Host ""
    Write-Host "APK built successfully:" -ForegroundColor Green
    Write-Host $apk -ForegroundColor Cyan
    Write-Host ""

    if ($Install) {
        Write-Host "Installing APK to connected Android device..." -ForegroundColor Cyan
        Invoke-Checked adb install -r $apk
        Write-Host "APK installed." -ForegroundColor Green
    } else {
        Write-Host "Install it manually or run the script with -Install when the phone is connected through USB debugging." -ForegroundColor Gray
    }
} else {
    Write-Host "APK was not found after build. Check Gradle output above." -ForegroundColor Red
    exit 1
}
