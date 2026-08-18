package migrations

import (
	"context"
	"embed"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed V20260620000100__admin_panel_extras.sql V20260616000100__home_widgets.sql V20260621000100__nullable_token_tx_admin.sql V20260622000100__user_subscriptions.sql V20260813000100__payments.sql
var adminPanelSQL embed.FS

var adminPanelFiles = []string{
	"V20260620000100__admin_panel_extras.sql",
	"V20260616000100__home_widgets.sql",
	"V20260621000100__nullable_token_tx_admin.sql",
	"V20260622000100__user_subscriptions.sql",
	"V20260813000100__payments.sql",
}

// ApplyAdminPanel ensures admin panel tables exist (idempotent CREATE IF NOT EXISTS).
func ApplyAdminPanel(ctx context.Context, pool *pgxpool.Pool) error {
	for _, name := range adminPanelFiles {
		raw, err := adminPanelSQL.ReadFile(name)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", name, err)
		}
		if _, err := pool.Exec(ctx, string(raw)); err != nil {
			return fmt.Errorf("apply migration %s: %w", name, err)
		}
		slog.InfoContext(ctx, "migration applied", slog.String("file", name))
	}
	return nil
}
