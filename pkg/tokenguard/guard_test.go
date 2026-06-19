package tokenguard

import (
	"context"
	"errors"
	"testing"
)

func TestCheckIdentity_EmptyTelegramID(t *testing.T) {
	g := &Guard{}
	if err := g.CheckIdentity(context.Background(), "", ""); !errors.Is(err, ErrTelegramIDRequired) {
		t.Fatalf("err = %v, want ErrTelegramIDRequired", err)
	}
}

func TestCheckAccess_NilGuard(t *testing.T) {
	var g *Guard
	if err := g.CheckAccess(context.Background(), "123", ""); err != nil {
		t.Fatalf("nil guard should skip check, got %v", err)
	}
}
