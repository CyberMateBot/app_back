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
  V20260528000100__web_auth_and_prompts.sql
  V20260607000100__profile_is_active.sql
  V20260608000100__token_transactions.sql
  V20260608000200__wallets_profile_unique.sql
  V20260620000100__admin_panel_extras.sql
  V20260616000100__home_widgets.sql
  V20260621000100__nullable_token_tx_admin.sql
  V20260622000100__user_subscriptions.sql
  V20260724000100__rebalance_billing_catalog.sql
)

for f in "${MIGRATIONS[@]}"; do
  echo "Applying $f ..."
  psql -h "$PG_HOST" -p "$PG_PORT" -U "$PG_USER" -d "$PG_DBNAME" -v ON_ERROR_STOP=1 -f "internal/migrations/$f"
done

echo "Done. Tables:"
psql -h "$PG_HOST" -p "$PG_PORT" -U "$PG_USER" -d "$PG_DBNAME" -c '\dt'
