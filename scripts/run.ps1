# Start the backend locally without Docker: go run ./cmd/service
# Usage: .\scripts\run.ps1

$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $PSScriptRoot
Set-Location $root

Write-Host "Working directory: $root"
Write-Host "Starting backend (HTTP :8090, gRPC :8091)..."
go run ./cmd/service
