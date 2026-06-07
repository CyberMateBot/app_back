package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindGoModuleRoot(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	root, ok := findGoModuleRoot(wd)
	if !ok {
		t.Fatal("expected go.mod in module tree")
	}

	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		t.Fatalf("go.mod missing in %s: %v", root, err)
	}
}

func TestDotEnvCandidatesIncludesModuleEnv(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	root, ok := findGoModuleRoot(wd)
	if !ok {
		t.Fatal("module root not found")
	}

	envPath := filepath.Join(root, ".env")
	candidates := dotEnvCandidates()

	found := false
	for _, c := range candidates {
		if filepath.Clean(c) == filepath.Clean(envPath) {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected %q in candidates: %v", envPath, candidates)
	}
}
