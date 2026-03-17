package frontendtests

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestFrontendUnitViaNPM(t *testing.T) {
	workdir := copyFrontendToTemp(t)

	npmCI := exec.Command("npm", "ci", "--no-audit", "--no-fund")
	npmCI.Dir = workdir
	npmCI.Env = os.Environ()
	if out, err := npmCI.CombinedOutput(); err != nil {
		t.Fatalf("npm ci failed: %v\n%s", err, string(out))
	}

	npmTest := exec.Command("npm", "run", "test")
	npmTest.Dir = workdir
	npmTest.Env = os.Environ()
	if out, err := npmTest.CombinedOutput(); err != nil {
		t.Fatalf("npm run test failed: %v\n%s", err, string(out))
	}

	if _, err := os.Stat(filepath.Join(workdir, "package.json")); err != nil {
		t.Fatalf("workspace unexpectedly invalid after test run: %v", err)
	}
}
