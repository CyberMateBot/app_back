package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
	repoModels "github.com/twelvepills-936/tgapp-/internal/repository/models"
	ucModels "github.com/twelvepills-936/tgapp-/internal/usecase/models"
	"github.com/twelvepills-936/tgapp-/pkg/admincatalog"
)

var defaultAdminSettings = ucModels.AdminSettingsOutput{
	RegistrationBonus:    20,
	ReferralBonus:        300,
	ReferralRefereeBonus: 30,
	TokenExpiryDays:      60,
	MaintenanceMode:      false,
	YookassaEnabled:      true,
	TelegramStarsEnabled: true,
	CoinRateRub:          1.0,
}

func (uc *useCase) ListAdminEvents(ctx context.Context, limit int32) (ucModels.AdminListEventsOutput, error) {
	items, err := uc.repo.ListAdminEvents(ctx, nil, limit)
	if err != nil {
		return ucModels.AdminListEventsOutput{}, err
	}

	out := ucModels.AdminListEventsOutput{
		Data: make([]ucModels.AdminEventItem, 0, len(items)),
	}
	for _, item := range items {
		out.Data = append(out.Data, ucModels.AdminEventItem{
			ID:      item.ID,
			Time:    item.Time.UTC().Format(time.RFC3339),
			User:    item.User,
			Action:  item.Action,
			Details: item.Details,
		})
	}
	return out, nil
}

func (uc *useCase) ListAdminTransactions(ctx context.Context, input ucModels.AdminListTransactionsInput) (ucModels.AdminListTransactionsOutput, error) {
	input.Normalize()
	offset := (input.Page - 1) * input.PerPage

	stats, err := uc.repo.GetAdminTransactionStats(ctx, nil)
	if err != nil {
		return ucModels.AdminListTransactionsOutput{}, err
	}

	items, total, err := uc.repo.ListAdminTokenTransactions(ctx, nil, input.Operation, input.PerPage, offset)
	if err != nil {
		return ucModels.AdminListTransactionsOutput{}, err
	}

	out := ucModels.AdminListTransactionsOutput{
		Stats: ucModels.AdminTransactionStatsOutput{
			CreditsMonth:    stats.CreditsMonth,
			DebitsMonth:     stats.DebitsMonth,
			OperationsMonth: stats.OperationsMonth,
			AvgAmount:       stats.AvgAmount,
		},
		Data:  make([]ucModels.AdminTransactionItem, 0, len(items)),
		Total: total,
	}

	for _, item := range items {
		out.Data = append(out.Data, mapAdminTransaction(item))
	}
	return out, nil
}

func mapAdminTransaction(item repoModels.AdminTokenTransaction) ucModels.AdminTransactionItem {
	if strings.EqualFold(strings.TrimSpace(item.Source), "payment") {
		return mapAdminPaymentTransaction(item)
	}

	op := strings.ToLower(strings.TrimSpace(item.Operation))
	typeLabel := "Списание админом"
	amountLabel := fmt.Sprintf("-%d монет", item.Amount)
	if op == "credit" {
		typeLabel = "Начисление админом"
		amountLabel = fmt.Sprintf("+%d монет", item.Amount)
	}

	return ucModels.AdminTransactionItem{
		ID:          item.ID,
		User:        item.UserName,
		Type:        op,
		TypeLabel:   typeLabel,
		Amount:      item.Amount,
		AmountLabel: amountLabel,
		Method:      "admin",
		MethodLabel: "Админка",
		CreatedAt:   item.CreatedAt.UTC().Format(time.RFC3339),
		Status:      "completed",
		StatusLabel: "Успешно",
	}
}

func mapAdminPaymentTransaction(item repoModels.AdminTokenTransaction) ucModels.AdminTransactionItem {
	kind := strings.ToLower(strings.TrimSpace(item.PaymentKind))
	typeLabel := "Покупка монет"
	if kind == "subscription" {
		typeLabel = "Подписка"
	}

	amountLabel := fmt.Sprintf("+%d монет", item.Amount)
	if item.AmountRub > 0 {
		amountLabel = fmt.Sprintf("+%d монет · %.0f ₽", item.Amount, item.AmountRub)
	}

	status, statusLabel := mapPaymentStatus(item.Status)

	return ucModels.AdminTransactionItem{
		ID:          item.ID,
		User:        item.UserName,
		Type:        "credit",
		TypeLabel:   typeLabel,
		Amount:      item.Amount,
		AmountLabel: amountLabel,
		Method:      "yookassa",
		MethodLabel: "ЮKassa",
		CreatedAt:   item.CreatedAt.UTC().Format(time.RFC3339),
		Status:      status,
		StatusLabel: statusLabel,
	}
}

func mapPaymentStatus(raw string) (string, string) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "pending":
		return "pending", "Ожидание"
	case "canceled":
		return "canceled", "Отменён"
	case "refunded":
		return "refunded", "Возврат"
	default:
		return "succeeded", "Успешно"
	}
}

func (uc *useCase) ListAdminBroadcasts(ctx context.Context, input ucModels.AdminListBroadcastsInput) (ucModels.AdminListBroadcastsOutput, error) {
	input.Normalize()
	offset := (input.Page - 1) * input.PerPage

	items, total, err := uc.repo.ListAdminBroadcasts(ctx, nil, input.PerPage, offset)
	if err != nil {
		return ucModels.AdminListBroadcastsOutput{}, err
	}

	out := ucModels.AdminListBroadcastsOutput{
		Data:  make([]ucModels.AdminBroadcastItem, 0, len(items)),
		Total: total,
	}
	for _, item := range items {
		status := "completed"
		statusLabel := "Отправлено"
		if item.FailedCount > 0 && item.SentCount == 0 {
			status = "failed"
			statusLabel = "Ошибка"
		} else if item.FailedCount > 0 {
			status = "partial"
			statusLabel = "Частично"
		}

		out.Data = append(out.Data, ucModels.AdminBroadcastItem{
			ID:          item.ID,
			Message:     item.Message,
			Target:      item.Target,
			TargetLabel: broadcastTargetLabel(item.Target),
			Sent:        item.SentCount,
			Failed:      item.FailedCount,
			CreatedAt:   item.CreatedAt.UTC().Format(time.RFC3339),
			Status:      status,
			StatusLabel: statusLabel,
		})
	}
	return out, nil
}

func broadcastTargetLabel(target string) string {
	if strings.EqualFold(strings.TrimSpace(target), "active") {
		return "Активные (7 дней)"
	}
	return "Все пользователи"
}

func (uc *useCase) GetAdminSettings(ctx context.Context) (ucModels.AdminSettingsOutput, error) {
	raw, err := uc.repo.GetAdminSettings(ctx, nil)
	if err != nil {
		return ucModels.AdminSettingsOutput{}, err
	}
	return mergeAdminSettings(raw), nil
}

func (uc *useCase) UpdateAdminSettings(ctx context.Context, input ucModels.AdminUpdateSettingsInput) (ucModels.AdminSettingsOutput, error) {
	current, err := uc.GetAdminSettings(ctx)
	if err != nil {
		return ucModels.AdminSettingsOutput{}, err
	}

	if input.RegistrationBonus != nil {
		current.RegistrationBonus = *input.RegistrationBonus
	}
	if input.ReferralBonus != nil {
		current.ReferralBonus = *input.ReferralBonus
	}
	if input.ReferralRefereeBonus != nil {
		current.ReferralRefereeBonus = *input.ReferralRefereeBonus
	}
	if input.TokenExpiryDays != nil {
		current.TokenExpiryDays = *input.TokenExpiryDays
	}
	if input.MaintenanceMode != nil {
		current.MaintenanceMode = *input.MaintenanceMode
	}
	if input.YookassaEnabled != nil {
		current.YookassaEnabled = *input.YookassaEnabled
	}
	if input.TelegramStarsEnabled != nil {
		current.TelegramStarsEnabled = *input.TelegramStarsEnabled
	}
	if input.CoinRateRub != nil {
		current.CoinRateRub = *input.CoinRateRub
	}

	keys := map[string]any{
		"registration_bonus":      current.RegistrationBonus,
		"referral_bonus":          current.ReferralBonus,
		"referral_referee_bonus":  current.ReferralRefereeBonus,
		"token_expiry_days":       current.TokenExpiryDays,
		"maintenance_mode":        current.MaintenanceMode,
		"yookassa_enabled":        current.YookassaEnabled,
		"telegram_stars_enabled":  current.TelegramStarsEnabled,
		"coin_rate_rub":           current.CoinRateRub,
	}

	for key, value := range keys {
		if upsertErr := uc.repo.UpsertAdminSetting(ctx, nil, key, value); upsertErr != nil {
			return ucModels.AdminSettingsOutput{}, upsertErr
		}
	}

	return current, nil
}

func mergeAdminSettings(raw map[string]json.RawMessage) ucModels.AdminSettingsOutput {
	out := defaultAdminSettings
	if raw == nil {
		return out
	}

	decodeInt := func(key string, target *int64) {
		value, ok := raw[key]
		if !ok || len(value) == 0 {
			return
		}
		var n int64
		if err := json.Unmarshal(value, &n); err == nil {
			*target = n
		}
	}
	decodeBool := func(key string, target *bool) {
		value, ok := raw[key]
		if !ok || len(value) == 0 {
			return
		}
		var b bool
		if err := json.Unmarshal(value, &b); err == nil {
			*target = b
		}
	}
	decodeFloat := func(key string, target *float64) {
		value, ok := raw[key]
		if !ok || len(value) == 0 {
			return
		}
		var f float64
		if err := json.Unmarshal(value, &f); err == nil {
			*target = f
		}
	}

	decodeInt("registration_bonus", &out.RegistrationBonus)
	decodeInt("referral_bonus", &out.ReferralBonus)
	decodeInt("referral_referee_bonus", &out.ReferralRefereeBonus)
	decodeInt("token_expiry_days", &out.TokenExpiryDays)
	decodeBool("maintenance_mode", &out.MaintenanceMode)
	decodeBool("yookassa_enabled", &out.YookassaEnabled)
	decodeBool("telegram_stars_enabled", &out.TelegramStarsEnabled)
	decodeFloat("coin_rate_rub", &out.CoinRateRub)

	return out
}

func (uc *useCase) ListAdminModels(ctx context.Context) (ucModels.AdminListModelsOutput, error) {
	overrides, err := uc.repo.ListModelConfigs(ctx, nil)
	if err != nil {
		return ucModels.AdminListModelsOutput{}, err
	}

	out := ucModels.AdminListModelsOutput{
		Data: make([]ucModels.AdminModelItem, 0, 64),
	}

	for _, base := range admincatalog.BaseEntries() {
		item := ucModels.AdminModelItem{
			ID:       base.ID,
			Name:     base.Name,
			Provider: base.Provider,
			Category: base.Category,
			Price:    int64(base.Price),
			Enabled:  base.Enabled,
		}
		if override, ok := overrides[base.ID]; ok {
			item.Name = override.Name
			item.Provider = override.Provider
			item.Category = override.Category
			item.Price = override.PriceCoins
			item.Enabled = override.Enabled
		}
		out.Data = append(out.Data, item)
	}

	return out, nil
}

func (uc *useCase) UpdateAdminModel(ctx context.Context, input ucModels.AdminUpdateModelInput) (ucModels.AdminModelItem, error) {
	modelID := strings.TrimSpace(input.ModelID)
	if modelID == "" {
		return ucModels.AdminModelItem{}, fmt.Errorf("%w: model_id is required", ucModels.ErrInvalidInput)
	}

	var base admincatalog.Entry
	found := false
	for _, entry := range admincatalog.BaseEntries() {
		if entry.ID == modelID {
			base = entry
			found = true
			break
		}
	}
	if !found {
		return ucModels.AdminModelItem{}, fmt.Errorf("%w: model not found", ucModels.ErrInvalidInput)
	}

	item := ucModels.AdminModelItem{
		ID:       base.ID,
		Name:     base.Name,
		Provider: base.Provider,
		Category: base.Category,
		Price:    int64(base.Price),
		Enabled:  base.Enabled,
	}

	overrides, err := uc.repo.ListModelConfigs(ctx, nil)
	if err != nil {
		return ucModels.AdminModelItem{}, err
	}
	if override, ok := overrides[modelID]; ok {
		item.Name = override.Name
		item.Provider = override.Provider
		item.Category = override.Category
		item.Price = override.PriceCoins
		item.Enabled = override.Enabled
	}

	if input.Price != nil {
		if *input.Price < 0 {
			return ucModels.AdminModelItem{}, fmt.Errorf("%w: price must be >= 0", ucModels.ErrInvalidInput)
		}
		item.Price = *input.Price
	}
	if input.Enabled != nil {
		item.Enabled = *input.Enabled
	}

	cfg := repoModels.ModelConfig{
		ModelID:    item.ID,
		Category:   item.Category,
		Name:       item.Name,
		Provider:   item.Provider,
		PriceCoins: item.Price,
		Enabled:    item.Enabled,
	}
	if err := uc.repo.UpsertModelConfig(ctx, nil, cfg); err != nil {
		return ucModels.AdminModelItem{}, err
	}

	return item, nil
}
