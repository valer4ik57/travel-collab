param(
    [string]$Domain = "localhost",
    [switch]$Local
)

$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $PSScriptRoot
$target = Join-Path $root ".env.production"

if (Test-Path $target) {
    Write-Host ".env.production already exists: $target" -ForegroundColor Yellow
    Write-Host "Remove it manually if you want to regenerate it."
    exit 0
}

function New-SecretBase64([int]$Bytes = 48) {
    $rng = [System.Security.Cryptography.RandomNumberGenerator]::Create()
    $buffer = New-Object byte[] $Bytes
    $rng.GetBytes($buffer)
    return [Convert]::ToBase64String($buffer)
}

function New-UrlSafeSecret([int]$Length = 40) {
    # DATABASE_URL contains the database password directly, so the password must not
    # contain URL-reserved characters such as @, :, /, ?, #, &, + or =.
    # Keep this secret alphanumeric to avoid URL parsing errors in pgx/libpq.
    $alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
    $rng = [System.Security.Cryptography.RandomNumberGenerator]::Create()
    $buffer = New-Object byte[] $Length
    $rng.GetBytes($buffer)
    $chars = New-Object char[] $Length
    for ($i = 0; $i -lt $Length; $i++) {
        $chars[$i] = $alphabet[[int]($buffer[$i] % $alphabet.Length)]
    }
    return -join $chars
}

$dbPassword = New-UrlSafeSecret 40
$jwtSecret = New-SecretBase64 48

if ($Local -or $Domain -eq "localhost") {
    $appEnv = "development"
    $siteAddress = ":80"
    $frontendUrl = "http://localhost"
    $frontendUrls = "http://localhost,http://127.0.0.1,http://localhost:80,http://127.0.0.1:80,capacitor://localhost"
} else {
    $appEnv = "production"
    $siteAddress = $Domain
    $frontendUrl = "https://$Domain"
    $frontendUrls = "https://$Domain"
}

$content = @"
APP_ENV=$appEnv
TRAVEL_COLLAB_SITE_ADDRESS=$siteAddress
CADDY_EMAIL=
HTTP_PORT=80
HTTPS_PORT=443

POSTGRES_USER=travel_collab
POSTGRES_PASSWORD=$dbPassword
POSTGRES_DB=travel_collab
DATABASE_URL=postgres://travel_collab:$dbPassword@postgres:5432/travel_collab?sslmode=disable

JWT_SECRET=$jwtSecret
PORT=8080
FRONTEND_URL=$frontendUrl
FRONTEND_URLS=$frontendUrls

VITE_API_BASE_URL=/api/v1
VITE_WS_BASE_URL=

AUTH_RATE_LIMIT_REQUESTS=20
AUTH_RATE_LIMIT_WINDOW=1m
MAX_REQUEST_BODY_BYTES=1048576

GITHUB_CLIENT_ID=
GITHUB_CLIENT_SECRET=
GITHUB_CALLBACK_URL=
"@

Set-Content -Path $target -Value $content -Encoding UTF8
Write-Host "Created $target" -ForegroundColor Green
Write-Host "APP_ENV=$appEnv"
Write-Host "FRONTEND_URL=$frontendUrl"
Write-Host "TRAVEL_COLLAB_SITE_ADDRESS=$siteAddress"
Write-Host "Database password contains only URL-safe alphanumeric characters."
