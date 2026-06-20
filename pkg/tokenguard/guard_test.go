package tokenguard

import (
	"testing"

	"github.com/twelvepills-936/tgapp-/pkg/telegramauth"
)

func TestIsDevelopmentEnv(t *testing.T) {
	t.Setenv("ENVIRONMENT", "development")
	if !isDevelopmentEnv() {
		t.Fatal("expected development")
	}

	t.Setenv("ENVIRONMENT", "production")
	if isDevelopmentEnv() {
		t.Fatal("expected not development")
	}
}

func TestDevelopmentMockInitData(t *testing.T) {
	t.Setenv("ENVIRONMENT", "development")

	raw := `user={"id":777000,"first_name":"Dev_User"}`
	if !telegramauth.InitDataMissingHash(raw) {
		t.Fatal("mock init data should not include hash")
	}

	id, err := telegramauth.ExtractUserID(raw)
	if err != nil {
		t.Fatalf("ExtractUserID() err = %v", err)
	}
	if id != "777000" {
		t.Fatalf("telegram id = %q, want 777000", id)
	}
}

func TestCheckAccessSkipsBalanceInDevelopment(t *testing.T) {
	t.Setenv("ENVIRONMENT", "development")
	g := &Guard{db: nil}
	if err := g.CheckAccess(t.Context(), "777000", ""); err != nil {
		t.Fatalf("CheckAccess() err = %v, want nil when db is nil", err)
	}
}
