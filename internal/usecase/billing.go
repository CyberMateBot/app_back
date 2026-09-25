package usecase

import (
	"context"
	"encoding/json"
	"sort"

	ucModels "github.com/twelvepills-936/tgapp-/internal/usecase/models"
	"github.com/twelvepills-936/tgapp-/pkg/billing"
)

const (
	adminSettingSubscriptionPlans = "subscription_plans"
	adminSettingCoinPacks         = "coin_packs"
)

func (uc *useCase) ListAdminSubscriptionPlans(ctx context.Context) (ucModels.AdminListSubscriptionPlansOutput, error) {
	items, err := uc.loadSubscriptionPlans(ctx)
	if err != nil {
		return ucModels.AdminListSubscriptionPlansOutput{}, err
	}
	return ucModels.AdminListSubscriptionPlansOutput{Data: items}, nil
}

func (uc *useCase) UpdateAdminSubscriptionPlans(ctx context.Context, input ucModels.AdminUpdateSubscriptionPlansInput) (ucModels.AdminListSubscriptionPlansOutput, error) {
	items, err := ucModels.NormalizeSubscriptionPlans(input.Data)
	if err != nil {
		return ucModels.AdminListSubscriptionPlansOutput{}, err
	}
	if err := uc.saveSubscriptionPlans(ctx, items); err != nil {
		return ucModels.AdminListSubscriptionPlansOutput{}, err
	}
	return ucModels.AdminListSubscriptionPlansOutput{Data: items}, nil
}

func (uc *useCase) ListAdminCoinPacks(ctx context.Context) (ucModels.AdminListCoinPacksOutput, error) {
	items, err := uc.loadCoinPacks(ctx)
	if err != nil {
		return ucModels.AdminListCoinPacksOutput{}, err
	}
	return ucModels.AdminListCoinPacksOutput{Data: items}, nil
}

func (uc *useCase) UpdateAdminCoinPacks(ctx context.Context, input ucModels.AdminUpdateCoinPacksInput) (ucModels.AdminListCoinPacksOutput, error) {
	items, err := ucModels.NormalizeCoinPacks(input.Data)
	if err != nil {
		return ucModels.AdminListCoinPacksOutput{}, err
	}
	if err := uc.saveCoinPacks(ctx, items); err != nil {
		return ucModels.AdminListCoinPacksOutput{}, err
	}
	return ucModels.AdminListCoinPacksOutput{Data: items}, nil
}

func (uc *useCase) ResetAdminSubscriptionPlans(ctx context.Context) (ucModels.AdminListSubscriptionPlansOutput, error) {
	items := billing.DefaultSubscriptionPlans()
	if err := uc.saveSubscriptionPlans(ctx, items); err != nil {
		return ucModels.AdminListSubscriptionPlansOutput{}, err
	}
	return ucModels.AdminListSubscriptionPlansOutput{Data: items}, nil
}

func (uc *useCase) ResetAdminCoinPacks(ctx context.Context) (ucModels.AdminListCoinPacksOutput, error) {
	items := billing.DefaultCoinPacks()
	if err := uc.saveCoinPacks(ctx, items); err != nil {
		return ucModels.AdminListCoinPacksOutput{}, err
	}
	return ucModels.AdminListCoinPacksOutput{Data: items}, nil
}

func (uc *useCase) GetPublicBillingCatalog(ctx context.Context) (ucModels.PublicBillingCatalogOutput, error) {
	settings, err := uc.GetAdminSettings(ctx)
	if err != nil {
		return ucModels.PublicBillingCatalogOutput{}, err
	}
	plans, err := uc.loadSubscriptionPlans(ctx)
	if err != nil {
		return ucModels.PublicBillingCatalogOutput{}, err
	}
	packs, err := uc.loadCoinPacks(ctx)
	if err != nil {
		return ucModels.PublicBillingCatalogOutput{}, err
	}
	return ucModels.PublicBillingCatalogOutput{
		CoinRateRub: settings.CoinRateRub,
		Plans:       ucModels.FilterEnabledPlans(plans),
		CoinPacks:   ucModels.FilterEnabledCoinPacks(packs),
	}, nil
}

func (uc *useCase) loadSubscriptionPlans(ctx context.Context) ([]ucModels.SubscriptionPlanItem, error) {
	raw, err := uc.repo.GetAdminSettings(ctx, nil)
	if err != nil {
		return nil, err
	}
	items, ok := decodeSettingSlice[ucModels.SubscriptionPlanItem](raw, adminSettingSubscriptionPlans)
	if !ok || len(items) == 0 {
		return billing.DefaultSubscriptionPlans(), nil
	}
	items = normalizeLoadedPlans(items)
	sortPlans(items)
	return items, nil
}

func (uc *useCase) saveSubscriptionPlans(ctx context.Context, items []ucModels.SubscriptionPlanItem) error {
	sortPlans(items)
	return uc.repo.UpsertAdminSetting(ctx, nil, adminSettingSubscriptionPlans, items)
}

func (uc *useCase) loadCoinPacks(ctx context.Context) ([]ucModels.CoinPackItem, error) {
	raw, err := uc.repo.GetAdminSettings(ctx, nil)
	if err != nil {
		return nil, err
	}
	items, ok := decodeSettingSlice[ucModels.CoinPackItem](raw, adminSettingCoinPacks)
	if !ok || len(items) == 0 {
		return billing.DefaultCoinPacks(), nil
	}
	items = normalizeLoadedCoinPacks(items)
	sortCoinPacks(items)
	return items, nil
}

func (uc *useCase) saveCoinPacks(ctx context.Context, items []ucModels.CoinPackItem) error {
	sortCoinPacks(items)
	return uc.repo.UpsertAdminSetting(ctx, nil, adminSettingCoinPacks, items)
}

func decodeSettingSlice[T any](raw map[string]json.RawMessage, key string) ([]T, bool) {
	value, ok := raw[key]
	if !ok || len(value) == 0 {
		return nil, false
	}
	var items []T
	if err := json.Unmarshal(value, &items); err != nil {
		return nil, false
	}
	return items, true
}

func sortPlans(items []ucModels.SubscriptionPlanItem) {
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].SortOrder == items[j].SortOrder {
			return items[i].ID < items[j].ID
		}
		return items[i].SortOrder < items[j].SortOrder
	})
}

func sortCoinPacks(items []ucModels.CoinPackItem) {
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].SortOrder == items[j].SortOrder {
			return items[i].ID < items[j].ID
		}
		return items[i].SortOrder < items[j].SortOrder
	})
}

// normalizeLoadedPlans fixes legacy admin_settings rows saved before the enabled
// flag existed, and fills in missing defaults only when fields were left unpopulated,
// preserving all admin customizations (prices, coins, features, and locked lists).
func normalizeLoadedPlans(items []ucModels.SubscriptionPlanItem) []ucModels.SubscriptionPlanItem {
	if len(items) == 0 {
		return billing.DefaultSubscriptionPlans()
	}
	defaults := billing.DefaultSubscriptionPlans()
	defaultsByID := make(map[string]ucModels.SubscriptionPlanItem, len(defaults))
	for _, d := range defaults {
		defaultsByID[d.ID] = d
	}

	enabled := 0
	out := make([]ucModels.SubscriptionPlanItem, len(items))
	copy(out, items)
	for i := range out {
		if out[i].Enabled {
			enabled++
		}
		if d, ok := defaultsByID[out[i].ID]; ok {
			if out[i].Coins <= 0 {
				out[i].Coins = d.Coins
			}
			if len(out[i].Features) == 0 {
				out[i].Features = d.Features
			}
			if out[i].Locked == nil {
				out[i].Locked = d.Locked
			}
			if out[i].Name == "" {
				out[i].Name = d.Name
			}
			if out[i].BadgeClass == "" {
				out[i].BadgeClass = d.BadgeClass
			}
			if out[i].PriceSub == "" {
				out[i].PriceSub = d.PriceSub
			}
		}
	}
	if enabled == 0 {
		for i := range out {
			out[i].Enabled = true
		}
	}
	return out
}

// normalizeLoadedCoinPacks preserves admin customizations (prices, coins, names,
// badges) while providing safe fallbacks for missing/zero fields.
func normalizeLoadedCoinPacks(items []ucModels.CoinPackItem) []ucModels.CoinPackItem {
	defaults := billing.DefaultCoinPacks()
	if len(items) == 0 {
		return defaults
	}
	defaultsByID := make(map[string]ucModels.CoinPackItem, len(defaults))
	for _, d := range defaults {
		defaultsByID[d.ID] = d
	}

	out := make([]ucModels.CoinPackItem, 0, len(items))
	for _, item := range items {
		if d, ok := defaultsByID[item.ID]; ok {
			if item.Coins <= 0 {
				item.Coins = d.Coins
			}
			if item.PriceRub <= 0 {
				item.PriceRub = d.PriceRub
			}
			if item.Name == "" {
				item.Name = d.Name
			}
		}
		out = append(out, item)
	}
	return out
}
