# Start the backend locally without Docker: go run ./cmd/service
# Usage: .\scripts\run.ps1

$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $PSScriptRoot
Set-Location $root

Write-Host "Working directory: $root"

$busy = netstat -ano | Select-String -Pattern ':8090\s.*LISTENING|:8091\s.*LISTENING'
if ($busy) {
    Write-Host "WARN: ports 8090/8091 already in use. Stop the old backend (Ctrl+C) or:" -ForegroundColor Yellow
    Write-Host '  netstat -ano | findstr ":8090 :8091"' -ForegroundColor Yellow
    Write-Host "  taskkill /PID <pid> /F" -ForegroundColor Yellow
}

Write-Host "Ensure Postgres is up: docker compose up -d"
Write-Host "Starting backend (HTTP :8090, gRPC :8091)..."
go run ./cmd/service
