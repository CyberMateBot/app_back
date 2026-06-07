# Reset local PostgreSQL password for user "postgres".
# Run PowerShell AS ADMINISTRATOR.
#
# If PostgreSQL won't start after a previous attempt:
#   .\scripts\repair-postgres-service.ps1
# Then:
#   .\scripts\reset-postgres-password.ps1

param(
    [string]$NewPassword = "postgres",
    [string]$PgHost = "127.0.0.1",
    [string]$Port = "5432",
    [string]$User = "postgres",
    [string]$ServiceName = "postgresql-x64-18"
)

$ErrorActionPreference = "Stop"

function Test-IsAdmin {
    $current = [Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()
    return $current.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
}

function Write-NoBomText {
    param([string]$Path, [string]$Text)
    $utf8NoBom = New-Object System.Text.UTF8Encoding($false)
    [System.IO.File]::WriteAllText($Path, $Text, $utf8NoBom)
}

if (-not (Test-IsAdmin)) {
    Write-Host "Restart this script in an elevated PowerShell (Run as administrator)." -ForegroundColor Red
    exit 1
}

$dataDir = "C:\Program Files\PostgreSQL\18\data"
$pgHbaPath = Join-Path $dataDir "pg_hba.conf"
$pgHbaBackup = "$pgHbaPath.bak-cybermate"
$trustTemplate = Join-Path $PSScriptRoot "pg_hba.trust.conf"

if (-not (Test-Path $pgHbaPath)) {
    throw "pg_hba.conf not found: $pgHbaPath"
}
if (-not (Test-Path $trustTemplate)) {
    throw "Trust template not found: $trustTemplate"
}

$psqlCandidates = @(
    "C:\Program Files\PostgreSQL\18\bin\psql.exe",
    "C:\Program Files\PostgreSQL\17\bin\psql.exe"
)
$psqlExe = $psqlCandidates | Where-Object { Test-Path $_ } | Select-Object -First 1
if (-not $psqlExe) {
    $cmd = Get-Command psql -ErrorAction SilentlyContinue
    if ($cmd) { $psqlExe = $cmd.Source }
}
if (-not $psqlExe) {
    throw "psql.exe not found"
}

$service = Get-Service $ServiceName -ErrorAction SilentlyContinue
if (-not $service) {
    throw "Service not found: $ServiceName"
}
if ($service.Status -ne "Running") {
    Write-Host "PostgreSQL is stopped. Run .\scripts\repair-postgres-service.ps1 first." -ForegroundColor Yellow
    exit 1
}

if (-not (Test-Path $pgHbaBackup)) {
    Write-Host "Backing up pg_hba.conf ..."
    Copy-Item $pgHbaPath $pgHbaBackup -Force
}

Write-Host "Enabling temporary trust auth ..."
$trustRules = Get-Content $trustTemplate -Raw
Write-NoBomText -Path $pgHbaPath -Text $trustRules

Write-Host "Reloading PostgreSQL config ..."
& "C:\Program Files\PostgreSQL\18\bin\pg_ctl.exe" reload -D $dataDir
Start-Sleep -Seconds 2

Write-Host "Setting password for user '$User' ..."
$sql = "ALTER USER $User WITH PASSWORD '$NewPassword';"
& $psqlExe -h $PgHost -p $Port -U $User -d postgres -v ON_ERROR_STOP=1 -c $sql
if ($LASTEXITCODE -ne 0) {
    throw "ALTER USER failed"
}

Write-Host "Restoring pg_hba.conf ..."
if (Test-Path $pgHbaBackup) {
    Copy-Item $pgHbaBackup $pgHbaPath -Force
} else {
    throw "Backup missing: $pgHbaBackup"
}

& "C:\Program Files\PostgreSQL\18\bin\pg_ctl.exe" reload -D $dataDir

Write-Host ""
Write-Host "Done. Update app_back/.env:" -ForegroundColor Green
Write-Host "PG_HOST=127.0.0.1"
Write-Host "PG_PORT=$Port"
Write-Host "PG_USER=$User"
Write-Host "PG_PASSWORD=$NewPassword"
Write-Host "PG_DBNAME=myapp_db"
