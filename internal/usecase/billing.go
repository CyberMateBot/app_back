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
	items := normalizeLoadedPlans(billing.DefaultSubscriptionPlans())
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
// flag existed (json omitempty → false → all plans filtered from public catalog).
func normalizeLoadedPlans(items []ucModels.SubscriptionPlanItem) []ucModels.SubscriptionPlanItem {
	if len(items) == 0 {
		return items
	}
	enabled := 0
	for _, item := range items {
		if item.Enabled {
			enabled++
		}
	}
	if enabled == 0 {
		out := make([]ucModels.SubscriptionPlanItem, len(items))
		copy(out, items)
		for i := range out {
			out[i].Enabled = true
		}
		return out
	}
	return items
}
