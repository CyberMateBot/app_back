# Repair PostgreSQL after a failed password-reset attempt.
# Run PowerShell AS ADMINISTRATOR:
#   cd c:\cybermate\back\app_back
#   .\scripts\repair-postgres-service.ps1

param(
    [string]$ServiceName = "postgresql-x64-18"
)

$ErrorActionPreference = "Stop"

function Test-IsAdmin {
    $current = [Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()
    return $current.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
}

if (-not (Test-IsAdmin)) {
    Write-Host "Run this script as administrator." -ForegroundColor Red
    exit 1
}

$dataDir = "C:\Program Files\PostgreSQL\18\data"
$pgHbaPath = Join-Path $dataDir "pg_hba.conf"
$backupPath = "$pgHbaPath.bak-cybermate"

if (-not (Test-Path $pgHbaPath)) {
    throw "pg_hba.conf not found: $pgHbaPath"
}

function Start-PostgresService {
    $svc = Get-Service $ServiceName
    if ($svc.Status -eq "Running") {
        return
    }
    Write-Host "Starting PostgreSQL service ..."
    Start-Service $ServiceName -ErrorAction Stop
    Start-Sleep -Seconds 2
}

if (Test-Path $backupPath) {
    Write-Host "Restoring pg_hba.conf from backup ..."
    Copy-Item $backupPath $pgHbaPath -Force
}

try {
    Start-PostgresService
    Write-Host "PostgreSQL is running." -ForegroundColor Green
    Write-Host "Next step: .\scripts\reset-postgres-password.ps1"
    exit 0
} catch {
    Write-Host "Start with backup failed: $($_.Exception.Message)"
    Write-Host "Trying trust config without UTF-8 BOM ..."
}

$trustTemplate = Join-Path $PSScriptRoot "pg_hba.trust.conf"
if (-not (Test-Path $trustTemplate)) {
    throw "Trust template not found: $trustTemplate"
}

$utf8NoBom = New-Object System.Text.UTF8Encoding($false)
$trustRules = Get-Content $trustTemplate -Raw
[System.IO.File]::WriteAllText($pgHbaPath, $trustRules, $utf8NoBom)

Start-PostgresService
Write-Host "PostgreSQL is running with temporary trust auth." -ForegroundColor Green
Write-Host "Next step: .\scripts\reset-postgres-password.ps1"
