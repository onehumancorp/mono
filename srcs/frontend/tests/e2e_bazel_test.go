package frontendtests

import (
	"os"
	"os/exec"
	"testing"
)

func TestFrontendE2EViaNPM(t *testing.T) {
	workdir := copyFrontendToTemp(t)
	backendBin, err := findBackendBinary()
	if err != nil {
		t.Fatalf("backend binary discovery failed: %v", err)
	}

	npmCI := exec.Command("npm", "ci", "--no-audit", "--no-fund")
	npmCI.Dir = workdir
	npmCI.Env = os.Environ()
	if out, err := npmCI.CombinedOutput(); err != nil {
		t.Fatalf("npm ci failed: %v\n%s", err, string(out))
	}

	npmE2E := exec.Command("npm", "run", "test:e2e")
	npmE2E.Dir = workdir
	npmE2E.Env = append(os.Environ(), "PLAYWRIGHT_BACKEND_COMMAND="+backendBin)
	if out, err := npmE2E.CombinedOutput(); err != nil {
		t.Fatalf("npm run test:e2e failed: %v\n%s", err, string(out))
	}
}
