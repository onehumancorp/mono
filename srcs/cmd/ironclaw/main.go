// Package main is the entry-point for the ironclaw CLI.
//
// IronClaw is a security and audit-focused agent CLI that integrates with the
// OneHumanCorp platform.  It can authenticate against the IronClaw provider,
// run static-analysis security scans, and report findings.
//
// Usage:
//
//	ironclaw <command> [flags]
//
// Commands:
//
//	version   Print the version string and exit.
//	auth      Store or validate IronClaw API credentials.
//	scan      Run a security scan against the specified target path.
//	providers List all registered agent providers and their status.
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/onehumancorp/mono/srcs/agents"
)

// Version is injected at link time via -ldflags "-X main.Version=x.y.z".
var Version = "dev"

// exitFunc is a variable so tests can intercept calls to os.Exit.
var exitFunc = os.Exit

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "ironclaw: %v\n", err)
		exitFunc(1)
	}
}

// run is the testable entry-point.  args is os.Args[1:].
func run(args []string, stdout, stderr *os.File) error {
	if len(args) == 0 {
		return errors.New(usageText())
	}

	cmd, rest := args[0], args[1:]

	switch cmd {
	case "version":
		return runVersion(stdout)
	case "auth":
		return runAuth(rest, stdout)
	case "scan":
		return runScan(rest, stdout)
	case "providers":
		return runProviders(stdout)
	case "help", "--help", "-h":
		fmt.Fprintln(stdout, usageText())
		return nil
	default:
		return fmt.Errorf("unknown command %q — run 'ironclaw help'", cmd)
	}
}

func usageText() string {
	return `ironclaw – security and audit-focused agent CLI

Usage:
  ironclaw <command> [flags]

Commands:
  version     Print version and exit.
  auth        Authenticate with the IronClaw agent provider.
  scan        Run a security scan against a target path.
  providers   List all registered agent providers.
  help        Show this help message.`
}

// ── version ──────────────────────────────────────────────────────────────────

func runVersion(out *os.File) error {
	fmt.Fprintf(out, "ironclaw %s\n", Version)
	return nil
}

// ── auth ─────────────────────────────────────────────────────────────────────

func runAuth(args []string, out *os.File) error {
	apiKey := ""
	for i, a := range args {
		if a == "--api-key" && i+1 < len(args) {
			apiKey = args[i+1]
		}
		if strings.HasPrefix(a, "--api-key=") {
			apiKey = strings.TrimPrefix(a, "--api-key=")
		}
	}

	// Fall back to environment variable.
	if apiKey == "" {
		apiKey = os.Getenv("IRONCLAW_API_KEY")
	}

	if apiKey == "" {
		return errors.New("--api-key flag or IRONCLAW_API_KEY environment variable is required")
	}

	r := agents.DefaultRegistry()
	if err := r.Authenticate(agents.ProviderTypeIronClaw, agents.Credentials{APIKey: apiKey}); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	fmt.Fprintln(out, "ironclaw: authenticated successfully")
	return nil
}

// ── scan ─────────────────────────────────────────────────────────────────────

// ScanResult is returned by runScan and is JSON-serialisable.
// Accepts no parameters.
// Returns nothing.
// Produces no errors.
// Has no side effects.
type ScanResult struct {
	Target   string        `json:"target"`
	Findings []ScanFinding `json:"findings"`
	Summary  string        `json:"summary"`
}

// ScanFinding represents a single security finding.
// Accepts no parameters.
// Returns nothing.
// Produces no errors.
// Has no side effects.
type ScanFinding struct {
	Severity string `json:"severity"`
	File     string `json:"file"`
	Message  string `json:"message"`
}

func runScan(args []string, out *os.File) error {
	target := "."
	jsonOutput := false

	for i, a := range args {
		switch {
		case a == "--json":
			jsonOutput = true
		case a == "--target" && i+1 < len(args):
			target = args[i+1]
		case strings.HasPrefix(a, "--target="):
			target = strings.TrimPrefix(a, "--target=")
		case !strings.HasPrefix(a, "-"):
			// Positional argument treated as target.
			target = a
		}
	}

	result, err := performScan(target)
	if err != nil {
		return err
	}

	if jsonOutput {
		enc := json.NewEncoder(out)
		enc.SetIndent("", "  ")
		return enc.Encode(result)
	}

	fmt.Fprintf(out, "ironclaw scan: %s\n", result.Target)
	for _, f := range result.Findings {
		fmt.Fprintf(out, "  [%s] %s: %s\n", f.Severity, f.File, f.Message)
	}
	fmt.Fprintf(out, "%s\n", result.Summary)
	return nil
}

// performScan walks the target directory and collects findings.
func performScan(target string) (ScanResult, error) {
	abs, err := filepath.Abs(target)
	if err != nil {
		return ScanResult{}, fmt.Errorf("cannot resolve target path: %w", err)
	}

	info, err := os.Stat(abs)
	if err != nil {
		return ScanResult{}, fmt.Errorf("target path %q not found: %w", target, err)
	}

	result := ScanResult{Target: abs}

	if info.IsDir() {
		err = filepath.Walk(abs, func(path string, fi os.FileInfo, walkErr error) error {
			if walkErr != nil {
				// Record skipped entries so the caller can see what was inaccessible.
				result.Findings = append(result.Findings, ScanFinding{
					Severity: "INFO",
					File:     path,
					Message:  fmt.Sprintf("skipped (unreadable): %v", walkErr),
				})
				return nil
			}
			if fi.IsDir() {
				return nil
			}
			if findings := analyseFile(path); len(findings) > 0 {
				result.Findings = append(result.Findings, findings...)
			}
			return nil
		})
		if err != nil {
			return ScanResult{}, fmt.Errorf("scan failed: %w", err)
		}
	} else {
		result.Findings = analyseFile(abs)
	}

	count := len(result.Findings)
	switch count {
	case 0:
		result.Summary = "No findings – target appears clean."
	case 1:
		result.Summary = "1 finding detected."
	default:
		result.Summary = fmt.Sprintf("%d findings detected.", count)
	}

	return result, nil
}

// analyseFile applies simple heuristic checks to a single file.
func analyseFile(path string) []ScanFinding {
	var findings []ScanFinding

	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	content := string(data)

	checks := []struct {
		severity string
		pattern  string
		message  string
	}{
		{"HIGH", "TODO: fix security", "insecure TODO comment found"},
		{"HIGH", "os.Setenv(\"AWS_SECRET", "hardcoded AWS secret detected"},
		{"HIGH", "password = \"", "hardcoded password detected"},
		{"MEDIUM", "fmt.Sprintf(\"%s\", err)", "error message interpolation may leak sensitive info"},
		{"INFO", "// FIXME", "FIXME comment may indicate incomplete security fix"},
	}

	for _, c := range checks {
		if strings.Contains(content, c.pattern) {
			findings = append(findings, ScanFinding{
				Severity: c.severity,
				File:     path,
				Message:  c.message,
			})
		}
	}
	return findings
}

// ── providers ────────────────────────────────────────────────────────────────

func runProviders(out *os.File) error {
	r := agents.DefaultRegistry()
	infos := r.Infos()

	for _, info := range infos {
		status := "unauthenticated"
		if info.IsAuthenticated {
			status = "authenticated"
		}
		fmt.Fprintf(out, "%-12s  %-16s  %s\n", info.Type, status, info.Description)
	}
	return nil
}
