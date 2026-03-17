package frontendtests

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestFrontendLayout(t *testing.T) {
	root := frontendRunfilesRoot(t)
	required := []string{
		"package.json",
		"index.html",
		"vite.config.ts",
		"vitest.config.ts",
		filepath.Join("src", "main.tsx"),
		filepath.Join("src", "App.tsx"),
		filepath.Join("src", "App.test.tsx"),
	}
	for _, f := range required {
		if _, err := os.Stat(filepath.Join(root, f)); err != nil {
			t.Fatalf("required file missing %s: %v", f, err)
		}
	}

	pkgBytes, err := os.ReadFile(filepath.Join(root, "package.json"))
	if err != nil {
		t.Fatalf("read package.json: %v", err)
	}
	var pkg struct {
		Scripts map[string]string `json:"scripts"`
	}
	if err := json.Unmarshal(pkgBytes, &pkg); err != nil {
		t.Fatalf("parse package.json: %v", err)
	}
	for _, script := range []string{"dev", "build", "lint", "test", "test:e2e"} {
		if _, ok := pkg.Scripts[script]; !ok {
			t.Fatalf("missing npm script %q", script)
		}
	}
}
