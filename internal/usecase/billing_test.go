package usecase

import (
	"context"
	"encoding/json"
	"sync"
	"testing"

	"github.com/jackc/pgx/v5"
	ucModels "github.com/twelvepills-936/tgapp-/internal/usecase/models"
	"github.com/twelvepills-936/tgapp-/pkg/billing"
	"github.com/twelvepills-936/tgapp-/pkg/config"
)

type billingFakeRepo struct {
	fakeRepo
	mu       sync.Mutex
	settings map[string]json.RawMessage
}

func newBillingFakeRepo() *billingFakeRepo {
	return &billingFakeRepo{
		settings: make(map[string]json.RawMessage),
	}
}

func (r *billingFakeRepo) GetAdminSettings(ctx context.Context, tx pgx.Tx) (map[string]json.RawMessage, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	copied := make(map[string]json.RawMessage, len(r.settings))
	for k, v := range r.settings {
		copied[k] = append(json.RawMessage(nil), v...)
	}
	return copied, nil
}

func (r *billingFakeRepo) UpsertAdminSetting(ctx context.Context, tx pgx.Tx, key string, value any) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}
	r.settings[key] = raw
	return nil
}

func TestBilling_PlansDefaultAndCustomUpdate(t *testing.T) {
	repo := newBillingFakeRepo()
	uc := NewUseCase(repo, config.ConfigJWT{})
	ctx := context.Background()

	// 1. Defaults returned initially
	initial, err := uc.ListAdminSubscriptionPlans(ctx)
	if err != nil {
		t.Fatalf("ListAdminSubscriptionPlans error: %v", err)
	}
	if len(initial.Data) != len(billing.DefaultSubscriptionPlans()) {
		t.Fatalf("expected %d default plans, got %d", len(billing.DefaultSubscriptionPlans()), len(initial.Data))
	}

	// 2. Admin edits plans:
	// - basic: price changed 149 -> 299, coins lowered 160 -> 75, custom features and locked
	// - pro: disabled
	updatedPlans := make([]ucModels.SubscriptionPlanItem, len(initial.Data))
	copy(updatedPlans, initial.Data)
	for i := range updatedPlans {
		if updatedPlans[i].ID == "basic" {
			updatedPlans[i].PriceRub = 299
			updatedPlans[i].Coins = 75
			updatedPlans[i].Features = []string{"Custom Basic Feature 1", "Custom Basic Feature 2"}
			updatedPlans[i].Locked = []string{"Video 4K", "Ultra Audio"}
		}
		if updatedPlans[i].ID == "pro" {
			updatedPlans[i].Enabled = false
		}
	}

	saved, err := uc.UpdateAdminSubscriptionPlans(ctx, ucModels.AdminUpdateSubscriptionPlansInput{Data: updatedPlans})
	if err != nil {
		t.Fatalf("UpdateAdminSubscriptionPlans error: %v", err)
	}
	if len(saved.Data) != len(updatedPlans) {
		t.Fatalf("expected %d plans in response, got %d", len(updatedPlans), len(saved.Data))
	}

	// 3. Load plans again - must preserve all custom edits
	loaded, err := uc.ListAdminSubscriptionPlans(ctx)
	if err != nil {
		t.Fatalf("ListAdminSubscriptionPlans reload error: %v", err)
	}

	var foundBasic, foundPro *ucModels.SubscriptionPlanItem
	for i := range loaded.Data {
		if loaded.Data[i].ID == "basic" {
			foundBasic = &loaded.Data[i]
		}
		if loaded.Data[i].ID == "pro" {
			foundPro = &loaded.Data[i]
		}
	}

	if foundBasic == nil {
		t.Fatal("basic plan not found in loaded plans")
	}
	if foundBasic.PriceRub != 299 {
		t.Errorf("expected basic price 299, got %d", foundBasic.PriceRub)
	}
	if foundBasic.Coins != 75 {
		t.Errorf("expected basic coins 75 (not clamped to 160), got %d", foundBasic.Coins)
	}
	if len(foundBasic.Features) != 2 || foundBasic.Features[0] != "Custom Basic Feature 1" {
		t.Errorf("expected custom features preserved, got %+v", foundBasic.Features)
	}
	if len(foundBasic.Locked) != 2 || foundBasic.Locked[0] != "Video 4K" {
		t.Errorf("expected custom locked preserved, got %+v", foundBasic.Locked)
	}

	if foundPro == nil {
		t.Fatal("pro plan not found in loaded plans")
	}
	if foundPro.Enabled {
		t.Errorf("expected pro to be disabled (Enabled=false), got true")
	}

	// 4. Public catalog should filter out disabled 'pro' plan and reflect 299 ₽
	catalog, err := uc.GetPublicBillingCatalog(ctx)
	if err != nil {
		t.Fatalf("GetPublicBillingCatalog error: %v", err)
	}
	for _, p := range catalog.Plans {
		if p.ID == "pro" {
			t.Errorf("disabled plan 'pro' should not appear in public catalog")
		}
		if p.ID == "basic" && p.PriceRub != 299 {
			t.Errorf("expected basic price 299 in public catalog, got %d", p.PriceRub)
		}
	}

	// 5. Reset to defaults
	resetOut, err := uc.ResetAdminSubscriptionPlans(ctx)
	if err != nil {
		t.Fatalf("ResetAdminSubscriptionPlans error: %v", err)
	}
	for _, p := range resetOut.Data {
		if p.ID == "basic" && p.PriceRub != 149 {
			t.Errorf("expected reset basic price 149, got %d", p.PriceRub)
		}
	}
}

func TestBilling_CoinPacksDefaultAndCustomUpdate(t *testing.T) {
	repo := newBillingFakeRepo()
	uc := NewUseCase(repo, config.ConfigJWT{})
	ctx := context.Background()

	initial, err := uc.ListAdminCoinPacks(ctx)
	if err != nil {
		t.Fatalf("ListAdminCoinPacks error: %v", err)
	}

	// Edit packs:
	// - pack-300: price changed 349 -> 499, coins lowered 300 -> 150, custom name ending in монет, badge cleared
	updatedPacks := make([]ucModels.CoinPackItem, len(initial.Data))
	copy(updatedPacks, initial.Data)
	for i := range updatedPacks {
		if updatedPacks[i].ID == "pack-300" {
			updatedPacks[i].PriceRub = 499
			updatedPacks[i].Coins = 150
			updatedPacks[i].Name = "Специальный пак 150 монет"
			updatedPacks[i].Badge = ""
		}
	}

	_, err = uc.UpdateAdminCoinPacks(ctx, ucModels.AdminUpdateCoinPacksInput{Data: updatedPacks})
	if err != nil {
		t.Fatalf("UpdateAdminCoinPacks error: %v", err)
	}

	loaded, err := uc.ListAdminCoinPacks(ctx)
	if err != nil {
		t.Fatalf("ListAdminCoinPacks reload error: %v", err)
	}

	var foundPack300 *ucModels.CoinPackItem
	for i := range loaded.Data {
		if loaded.Data[i].ID == "pack-300" {
			foundPack300 = &loaded.Data[i]
		}
	}

	if foundPack300 == nil {
		t.Fatal("pack-300 not found in loaded packs")
	}
	if foundPack300.PriceRub != 499 {
		t.Errorf("expected pack-300 price 499, got %d", foundPack300.PriceRub)
	}
	if foundPack300.Coins != 150 {
		t.Errorf("expected pack-300 coins 150 (not clamped to 300), got %d", foundPack300.Coins)
	}
	if foundPack300.Name != "Специальный пак 150 монет" {
		t.Errorf("expected custom pack name preserved, got %q", foundPack300.Name)
	}
	if foundPack300.Badge != "" {
		t.Errorf("expected cleared badge to stay empty, got %q", foundPack300.Badge)
	}

	// Reset packs
	resetOut, err := uc.ResetAdminCoinPacks(ctx)
	if err != nil {
		t.Fatalf("ResetAdminCoinPacks error: %v", err)
	}
	for _, p := range resetOut.Data {
		if p.ID == "pack-300" && p.PriceRub != 349 {
			t.Errorf("expected reset pack-300 price 349, got %d", p.PriceRub)
		}
	}
}
