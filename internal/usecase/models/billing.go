package models

import (
	"fmt"
	"strings"

	"github.com/twelvepills-936/tgapp-/pkg/billing"
)

type SubscriptionPlanItem = billing.SubscriptionPlan
type CoinPackItem = billing.CoinPack

type AdminListSubscriptionPlansOutput struct {
	Data []SubscriptionPlanItem `json:"data"`
}

type AdminUpdateSubscriptionPlansInput struct {
	Data []SubscriptionPlanItem `json:"data"`
}

type AdminListCoinPacksOutput struct {
	Data []CoinPackItem `json:"data"`
}

type AdminUpdateCoinPacksInput struct {
	Data []CoinPackItem `json:"data"`
}

type PublicBillingCatalogOutput struct {
	CoinRateRub float64                `json:"coin_rate_rub"`
	Plans       []SubscriptionPlanItem `json:"plans"`
	CoinPacks   []CoinPackItem         `json:"coin_packs"`
}

func NormalizeSubscriptionPlans(items []SubscriptionPlanItem) ([]SubscriptionPlanItem, error) {
	if len(items) == 0 {
		return nil, fmt.Errorf("%w: at least one plan is required", ErrInvalidInput)
	}
	seen := make(map[string]struct{}, len(items))
	out := make([]SubscriptionPlanItem, 0, len(items))
	for i, item := range items {
		id := strings.ToLower(strings.TrimSpace(item.ID))
		if id == "" {
			return nil, fmt.Errorf("%w: plan id is required", ErrInvalidInput)
		}
		if _, ok := seen[id]; ok {
			return nil, fmt.Errorf("%w: duplicate plan id %q", ErrInvalidInput, id)
		}
		seen[id] = struct{}{}
		item.ID = id
		item.Name = strings.TrimSpace(item.Name)
		if item.Name == "" {
			return nil, fmt.Errorf("%w: plan name is required", ErrInvalidInput)
		}
		if item.PriceRub < 0 {
			return nil, fmt.Errorf("%w: plan price must be >= 0", ErrInvalidInput)
		}
		if item.Coins < 0 {
			return nil, fmt.Errorf("%w: plan coins must be >= 0", ErrInvalidInput)
		}
		if item.SortOrder == 0 {
			item.SortOrder = int32(i + 1)
		}
		if item.BadgeClass == "" {
			item.BadgeClass = "free"
		}
		if item.PriceSub == "" {
			if item.PriceRub == 0 {
				item.PriceSub = "навсегда"
			} else {
				item.PriceSub = "/ месяц"
			}
		}
		out = append(out, item)
	}
	return out, nil
}

func NormalizeCoinPacks(items []CoinPackItem) ([]CoinPackItem, error) {
	if len(items) == 0 {
		return nil, fmt.Errorf("%w: at least one coin pack is required", ErrInvalidInput)
	}
	seen := make(map[string]struct{}, len(items))
	out := make([]CoinPackItem, 0, len(items))
	for i, item := range items {
		id := strings.ToLower(strings.TrimSpace(item.ID))
		if id == "" {
			return nil, fmt.Errorf("%w: coin pack id is required", ErrInvalidInput)
		}
		if _, ok := seen[id]; ok {
			return nil, fmt.Errorf("%w: duplicate coin pack id %q", ErrInvalidInput, id)
		}
		seen[id] = struct{}{}
		item.ID = id
		item.Name = strings.TrimSpace(item.Name)
		if item.Name == "" {
			return nil, fmt.Errorf("%w: coin pack name is required", ErrInvalidInput)
		}
		if item.Coins <= 0 {
			return nil, fmt.Errorf("%w: coin pack coins must be > 0", ErrInvalidInput)
		}
		if item.PriceRub <= 0 {
			return nil, fmt.Errorf("%w: coin pack price must be > 0", ErrInvalidInput)
		}
		if item.SortOrder == 0 {
			item.SortOrder = int32(i + 1)
		}
		out = append(out, item)
	}
	return out, nil
}

func FilterEnabledPlans(items []SubscriptionPlanItem) []SubscriptionPlanItem {
	out := make([]SubscriptionPlanItem, 0, len(items))
	for _, item := range items {
		if item.Enabled {
			out = append(out, item)
		}
	}
	return out
}

func FilterEnabledCoinPacks(items []CoinPackItem) []CoinPackItem {
	out := make([]CoinPackItem, 0, len(items))
	for _, item := range items {
		if item.Enabled {
			out = append(out, item)
		}
	}
	return out
}
