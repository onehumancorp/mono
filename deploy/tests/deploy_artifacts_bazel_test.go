package deploytests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	testSrcDir := os.Getenv("TEST_SRCDIR")
	workspace := os.Getenv("TEST_WORKSPACE")
	if testSrcDir == "" || workspace == "" {
		t.Fatalf("TEST_SRCDIR/TEST_WORKSPACE must be set by bazel")
	}
	return filepath.Join(testSrcDir, workspace)
}

func mustRead(t *testing.T, p string) string {
	t.Helper()
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("read %s: %v", p, err)
	}
	return string(b)
}

func TestDeployArtifacts(t *testing.T) {
	root := repoRoot(t)
	files := []string{
		"deploy/BUILD.bazel",
		"deploy/docker-compose.yml",
		"deploy/helm/ohc/Chart.yaml",
		"deploy/helm/ohc/Chart.lock",
		"deploy/helm/ohc/charts/redis-20.6.3.tgz",
		"deploy/helm/ohc/charts/postgresql-16.4.9.tgz",
		"deploy/helm/ohc/values.yaml",
		"deploy/tests/kind_helm_e2e_test.sh",
		"deploy/helm/ohc/templates/backend-deployment.yaml",
		"deploy/helm/ohc/templates/frontend-deployment.yaml",
	}
	for _, f := range files {
		if info, err := os.Stat(filepath.Join(root, f)); err != nil || info.Size() == 0 {
			t.Fatalf("expected non-empty file %s", f)
		}
	}

	deployBuild := mustRead(t, filepath.Join(root, "deploy/BUILD.bazel"))
	for _, want := range []string{"oci_image(", "distroless_static_debian12_nonroot", "backend_image", "frontend_image"} {
		if !strings.Contains(deployBuild, want) {
			t.Fatalf("deploy BUILD missing %q", want)
		}
	}

	compose := mustRead(t, filepath.Join(root, "deploy/docker-compose.yml"))
	for _, want := range []string{"backend:", "frontend:", "redis:", "postgres:", "mono-backend:dev", "mono-frontend:dev"} {
		if !strings.Contains(compose, want) {
			t.Fatalf("docker-compose missing %q", want)
		}
	}

	values := mustRead(t, filepath.Join(root, "deploy/helm/ohc/values.yaml"))
	for _, want := range []string{"backend", "frontend", "redis:", "postgresql:"} {
		if !strings.Contains(values, want) {
			t.Fatalf("values.yaml missing %q", want)
		}
	}

	backendTpl := mustRead(t, filepath.Join(root, "deploy/helm/ohc/templates/backend-deployment.yaml"))
	for _, want := range []string{"Deployment", "OHC_REDIS_URL", "OHC_POSTGRES_HOST"} {
		if !strings.Contains(backendTpl, want) {
			t.Fatalf("backend deployment template missing %q", want)
		}
	}

	frontendTpl := mustRead(t, filepath.Join(root, "deploy/helm/ohc/templates/frontend-deployment.yaml"))
	if !strings.Contains(frontendTpl, "Deployment") {
		t.Fatalf("frontend deployment template missing Deployment kind")
	}
}
