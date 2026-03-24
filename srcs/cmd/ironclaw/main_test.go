package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ── helpers ───────────────────────────────────────────────────────────────────

// captureRun executes run() and returns stdout content.
func captureRun(t *testing.T, args []string) (string, error) {
	t.Helper()

	// Create a temporary file for stdout capture.
	tmp, err := os.CreateTemp(t.TempDir(), "stdout-*")
	if err != nil {
		t.Fatalf("create temp stdout: %v", err)
	}
	defer tmp.Close()

	stderr, err := os.CreateTemp(t.TempDir(), "stderr-*")
	if err != nil {
		t.Fatalf("create temp stderr: %v", err)
	}
	defer stderr.Close()

	runErr := run(args, tmp, stderr)
	if _, err := tmp.Seek(0, 0); err != nil {
		t.Fatalf("seek stdout: %v", err)
	}
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(tmp); err != nil {
		t.Fatalf("read stdout: %v", err)
	}
	return buf.String(), runErr
}

// ── version ───────────────────────────────────────────────────────────────────

func TestRunVersion(t *testing.T) {
	out, err := captureRun(t, []string{"version"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "ironclaw") {
		t.Errorf("expected 'ironclaw' in version output, got: %q", out)
	}
}

// ── help ──────────────────────────────────────────────────────────────────────

func TestRunHelp(t *testing.T) {
	for _, arg := range []string{"help", "--help", "-h"} {
		t.Run(arg, func(t *testing.T) {
			out, err := captureRun(t, []string{arg})
			if err != nil {
				t.Fatalf("unexpected error for %q: %v", arg, err)
			}
			if !strings.Contains(out, "Usage:") {
				t.Errorf("expected usage info in output, got: %q", out)
			}
		})
	}
}

// ── unknown command ───────────────────────────────────────────────────────────

func TestRunUnknownCommand(t *testing.T) {
	_, err := captureRun(t, []string{"foobar"})
	if err == nil {
		t.Fatal("expected error for unknown command")
	}
	if !strings.Contains(err.Error(), "foobar") {
		t.Errorf("expected command name in error, got: %v", err)
	}
}

func TestRunNoArgs(t *testing.T) {
	_, err := captureRun(t, []string{})
	if err == nil {
		t.Fatal("expected error when no arguments provided")
	}
}

// ── auth ──────────────────────────────────────────────────────────────────────

func TestRunAuth_MissingKey(t *testing.T) {
	t.Setenv("IRONCLAW_API_KEY", "")
	_, err := captureRun(t, []string{"auth"})
	if err == nil {
		t.Fatal("expected error when no API key supplied")
	}
}

func TestRunAuth_WithFlagValue(t *testing.T) {
	out, err := captureRun(t, []string{"auth", "--api-key", "test-secret"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "authenticated") {
		t.Errorf("expected 'authenticated' in output, got: %q", out)
	}
}

func TestRunAuth_WithFlagEquals(t *testing.T) {
	out, err := captureRun(t, []string{"auth", "--api-key=test-secret"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "authenticated") {
		t.Errorf("expected 'authenticated' in output, got: %q", out)
	}
}

func TestRunAuth_WithEnvVar(t *testing.T) {
	t.Setenv("IRONCLAW_API_KEY", "env-key-123")
	out, err := captureRun(t, []string{"auth"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "authenticated") {
		t.Errorf("expected 'authenticated' in output, got: %q", out)
	}
}

// ── scan ──────────────────────────────────────────────────────────────────────

func TestRunScan_CleanDirectory(t *testing.T) {
	dir := t.TempDir()
	// Write a clean file with no pattern matches.
	if err := os.WriteFile(filepath.Join(dir, "clean.go"), []byte("package main\n"), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}

	out, err := captureRun(t, []string{"scan", dir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "No findings") {
		t.Errorf("expected 'No findings' for clean dir, got: %q", out)
	}
}

func TestRunScan_FindingDetected(t *testing.T) {
	dir := t.TempDir()
	content := `package main
// TODO: fix security issue here
`
	if err := os.WriteFile(filepath.Join(dir, "vuln.go"), []byte(content), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}

	out, err := captureRun(t, []string{"scan", dir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "finding") {
		t.Errorf("expected findings in output, got: %q", out)
	}
}

func TestRunScan_JSONOutput(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "ok.go"), []byte("package ok\n"), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}

	out, err := captureRun(t, []string{"scan", "--json", dir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, `"target"`) {
		t.Errorf("expected JSON output with 'target' field, got: %q", out)
	}
}

func TestRunScan_JSONFlag_Equals(t *testing.T) {
	dir := t.TempDir()
	out, err := captureRun(t, []string{"scan", "--json", "--target=" + dir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, `"findings"`) {
		t.Errorf("expected JSON with 'findings' key, got: %q", out)
	}
}

func TestRunScan_InvalidTarget(t *testing.T) {
	_, err := captureRun(t, []string{"scan", "/nonexistent/path/xyz"})
	if err == nil {
		t.Fatal("expected error for non-existent target")
	}
}

func TestRunScan_DefaultsToCurrentDir(t *testing.T) {
	// Scanning '.' should not panic regardless of what is in the repo root.
	// We intentionally ignore both the output and any error because the
	// current directory content is non-deterministic in a test environment.
	_, _ = captureRun(t, []string{"scan"})
}

func TestRunScan_SingleFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "file.go")
	if err := os.WriteFile(path, []byte("package main\n"), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}

	out, err := captureRun(t, []string{"scan", path})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "No findings") {
		t.Errorf("expected 'No findings' for single clean file, got: %q", out)
	}
}

// ── providers ─────────────────────────────────────────────────────────────────

func TestRunProviders(t *testing.T) {
	out, err := captureRun(t, []string{"providers"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, name := range []string{"claude", "gemini", "opencode", "openclaw", "ironclaw", "builtin"} {
		if !strings.Contains(out, name) {
			t.Errorf("expected provider %q in output, got: %q", name, out)
		}
	}
}

// ── analyseFile ───────────────────────────────────────────────────────────────

func TestAnalyseFile_HardcodedPassword(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.go")
	content := `var cfg = config{password = "hunter2"}`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}

	findings := analyseFile(path)
	found := false
	for _, f := range findings {
		if f.Severity == "HIGH" && strings.Contains(f.Message, "password") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected HIGH finding for hardcoded password, got: %v", findings)
	}
}

func TestAnalyseFile_NoFindings(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "clean.go")
	if err := os.WriteFile(path, []byte("package main\nfunc main(){}\n"), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}

	if findings := analyseFile(path); len(findings) != 0 {
		t.Errorf("expected no findings for clean file, got: %v", findings)
	}
}

func TestAnalyseFile_Unreadable(t *testing.T) {
	// analyseFile should return nil (no panic) for non-existent path.
	findings := analyseFile("/tmp/does-not-exist-ironclaw-test-12345")
	if findings != nil {
		t.Errorf("expected nil findings for unreadable file, got: %v", findings)
	}
}

// ── performScan ───────────────────────────────────────────────────────────────

func TestPerformScan_MultipleFindingsSummary(t *testing.T) {
	dir := t.TempDir()
	content := "TODO: fix security\npassword = \"secret\"\n"
	if err := os.WriteFile(filepath.Join(dir, "bad.go"), []byte(content), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}

	result, err := performScan(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Findings) < 2 {
		t.Errorf("expected at least 2 findings, got: %d", len(result.Findings))
	}
	if !strings.Contains(result.Summary, "findings") {
		t.Errorf("expected plural 'findings' in summary for 2+, got: %q", result.Summary)
	}
}

func TestPerformScan_ExactlyOneFinding(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "one.go"), []byte("TODO: fix security"), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	result, err := performScan(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Summary != "1 finding detected." {
		t.Errorf("expected singular summary, got: %q", result.Summary)
	}
}

func TestPerformScan_BadPath(t *testing.T) {
	_, err := performScan("/definitely/does/not/exist")
	if err == nil {
		t.Fatal("expected error for invalid target path")
	}
}
