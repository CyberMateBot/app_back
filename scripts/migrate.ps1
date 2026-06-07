# Apply SQL migrations to PostgreSQL (local install or Docker).
# Usage:
#   .\scripts\migrate.ps1                  # local psql (reads .env)
#   .\scripts\migrate.ps1 -UseDocker       # docker container tgapp_postgres

param(
    [switch]$UseDocker,
    [string]$Container = "tgapp_postgres",
    [string]$Database = "",
    [string]$User = "",
    [string]$PgHost = "",
    [string]$Port = ""
)

$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $PSScriptRoot
$envFile = Join-Path $root ".env"

if (Test-Path $envFile) {
    Get-Content $envFile | ForEach-Object {
        $line = $_.Trim()
        if ($line -eq "" -or $line.StartsWith("#")) { return }
        if ($line -match '^\s*([^#=]+)=(.*)$') {
            $name = $matches[1].Trim()
            $value = $matches[2].Trim()
            if ($name -ne "") {
                Set-Item -Path "env:$name" -Value $value
            }
        }
    }
}

if (-not $Database) { $Database = $env:PG_DBNAME; if (-not $Database) { $Database = "myapp_db" } }
if (-not $User) { $User = $env:PG_USER; if (-not $User) { $User = "postgres" } }
if (-not $PgHost) { $PgHost = $env:PG_HOST; if (-not $PgHost) { $PgHost = "localhost" } }
if (-not $Port) { $Port = $env:PG_PORT; if (-not $Port) { $Port = "5432" } }
if ($env:PG_PASSWORD) { $env:PGPASSWORD = $env:PG_PASSWORD }
elseif ($env:PG_PASS) { $env:PGPASSWORD = $env:PG_PASS }

$migrations = @(
    "V20250805000000__users.sql",
    "V20251103000100__cybermate_core.sql",
    "V20251103000200__admin_resources.sql",
    "V20250525000000__profile_ui_theme.sql",
    "V20260527000000__prompt_history.sql",
    "V20260528000100__web_auth_and_prompts.sql"
)

function Resolve-PsqlExecutable {
    $psql = Get-Command psql -ErrorAction SilentlyContinue
    if ($psql) {
        return $psql.Source
    }

    $candidates = @(
        "C:\Program Files\PostgreSQL\18\bin\psql.exe",
        "C:\Program Files\PostgreSQL\17\bin\psql.exe",
        "C:\Program Files\PostgreSQL\16\bin\psql.exe",
        "C:\Program Files\PostgreSQL\15\bin\psql.exe"
    )

    foreach ($candidate in $candidates) {
        if (Test-Path $candidate) {
            return $candidate
        }
    }

    return $null
}

function Invoke-MigrationSql {
    param([string]$SqlPath)

    if ($UseDocker) {
        Get-Content $SqlPath -Raw | docker exec -i $Container psql -U $User -d $Database -v ON_ERROR_STOP=1
        return
    }

    $psqlExe = Resolve-PsqlExecutable
    if (-not $psqlExe) {
        throw "psql not found. Install PostgreSQL or pass -UseDocker with a running container."
    }

    & $psqlExe -h $PgHost -p $Port -U $User -d $Database -v ON_ERROR_STOP=1 -f $SqlPath
    if ($LASTEXITCODE -ne 0) {
        throw "psql failed for $SqlPath (exit $LASTEXITCODE)"
    }
}

foreach ($f in $migrations) {
    $path = Join-Path $root "internal\migrations\$f"
    Write-Host "Applying $f ..."
    Invoke-MigrationSql -SqlPath $path
}

Write-Host "Done. Tables:"
if ($UseDocker) {
    docker exec $Container psql -U $User -d $Database -c "\dt"
} else {
    $psqlExe = Resolve-PsqlExecutable
    if (-not $psqlExe) {
        throw "psql not found."
    }
    & $psqlExe -h $PgHost -p $Port -U $User -d $Database -c "\dt"
}
