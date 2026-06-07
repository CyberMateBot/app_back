#!/usr/bin/env bash
# Apply SQL migrations to local PostgreSQL (reads .env from app_back).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

if [[ -f .env ]]; then
  set -a
  # shellcheck disable=SC1091
  source <(grep -v '^\s*#' .env | grep -v '^\s*$' | sed 's/\r$//')
  set +a
fi

: "${PG_HOST:=localhost}"
: "${PG_PORT:=5432}"
: "${PG_USER:=postgres}"
: "${PG_DBNAME:=myapp_db}"
export PGPASSWORD="${PG_PASSWORD:-${PG_PASS:-postgres}}"

MIGRATIONS=(
  V20250805000000__users.sql
  V20251103000100__cybermate_core.sql
  V20251103000200__admin_resources.sql
  V20250525000000__profile_ui_theme.sql
  V20260527000000__prompt_history.sql
)

for f in "${MIGRATIONS[@]}"; do
  echo "Applying $f ..."
  psql -h "$PG_HOST" -p "$PG_PORT" -U "$PG_USER" -d "$PG_DBNAME" -v ON_ERROR_STOP=1 -f "internal/migrations/$f"
done

echo "Done. Tables:"
psql -h "$PG_HOST" -p "$PG_PORT" -U "$PG_USER" -d "$PG_DBNAME" -c '\dt'
