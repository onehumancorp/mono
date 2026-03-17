package frontendtests

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func frontendRunfilesRoot(t *testing.T) string {
	t.Helper()
	testSrcDir := os.Getenv("TEST_SRCDIR")
	workspace := os.Getenv("TEST_WORKSPACE")
	if testSrcDir == "" || workspace == "" {
		t.Fatalf("TEST_SRCDIR/TEST_WORKSPACE must be set by bazel")
	}
	return filepath.Join(testSrcDir, workspace, "srcs", "frontend")
}

func copyFrontendToTemp(t *testing.T) string {
	t.Helper()
	src := frontendRunfilesRoot(t)
	tmp := t.TempDir()
	cmd := exec.Command("cp", "-LR", src+"/.", tmp)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("copy frontend tree: %v\n%s", err, string(out))
	}
	for _, stale := range []string{"tsconfig.app.tsbuildinfo", "tsconfig.node.tsbuildinfo"} {
		_ = os.Remove(filepath.Join(tmp, stale))
	}
	return tmp
}

func findBackendBinary() (string, error) {
	testSrcDir := os.Getenv("TEST_SRCDIR")
	workspace := os.Getenv("TEST_WORKSPACE")
	if testSrcDir == "" || workspace == "" {
		return "", errors.New("TEST_SRCDIR/TEST_WORKSPACE must be set")
	}
	base := filepath.Join(testSrcDir, workspace, "srcs", "cmd", "ohc")
	candidates := []string{
		filepath.Join(base, "ohc"),
		filepath.Join(base, "ohc_", "ohc"),
	}
	for _, c := range candidates {
		if info, err := os.Stat(c); err == nil && !info.IsDir() {
			return c, nil
		}
	}

	var found string
	_ = filepath.Walk(base, func(path string, info os.FileInfo, err error) error {
		if err != nil || info == nil || info.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, "/ohc") {
			found = path
			return filepath.SkipDir
		}
		return nil
	})
	if found == "" {
		return "", errors.New("backend binary not found in runfiles")
	}
	return found, nil
}
