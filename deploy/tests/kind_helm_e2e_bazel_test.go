package deploytests

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestKindHelmE2E(t *testing.T) {
	testSrcDir := os.Getenv("TEST_SRCDIR")
	workspace := os.Getenv("TEST_WORKSPACE")
	if testSrcDir == "" || workspace == "" {
		t.Fatalf("TEST_SRCDIR/TEST_WORKSPACE must be set by bazel")
	}
	repo := filepath.Join(testSrcDir, workspace)
	script := filepath.Join(repo, "deploy", "tests", "kind_helm_e2e_test.sh")

	cmd := exec.Command("bash", script)
	cmd.Dir = repo
	cmd.Env = os.Environ()
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("kind helm e2e script failed: %v\n%s", err, string(out))
	}
}
