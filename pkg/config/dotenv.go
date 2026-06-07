package config

import (
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

// LoadDotEnv loads variables from the project .env file.
// Search order: ./.env, ./app_back/.env, then walk up to the directory that contains go.mod.
// Already-set environment variables are not overwritten (godotenv default).
func LoadDotEnv() {
	for _, path := range dotEnvCandidates() {
		if _, err := os.Stat(path); err != nil {
			continue
		}
		if err := godotenv.Load(path); err == nil {
			return
		}
	}
}

func dotEnvCandidates() []string {
	wd, err := os.Getwd()
	if err != nil {
		return []string{".env"}
	}

	seen := make(map[string]struct{})
	add := func(paths *[]string, p string) {
		if p == "" {
			return
		}
		clean := filepath.Clean(p)
		if _, ok := seen[clean]; ok {
			return
		}
		seen[clean] = struct{}{}
		*paths = append(*paths, clean)
	}

	var out []string
	add(&out, filepath.Join(wd, ".env"))
	add(&out, filepath.Join(wd, "app_back", ".env"))

	if root, ok := findGoModuleRoot(wd); ok {
		add(&out, filepath.Join(root, ".env"))
	}

	add(&out, ".env")
	return out
}

func findGoModuleRoot(start string) (string, bool) {
	dir := start
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false
		}
		dir = parent
	}
}
