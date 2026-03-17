# Developer Guide: Bazel Command Reference

## Core Test Commands

Run all default tests:

```bash
bazelisk test //...
```

Run key backend tests:

```bash
bazelisk test //srcs/orchestration:orchestration_test
bazelisk test //srcs/dashboard:dashboard_test
bazelisk test //srcs/cmd/ohc:ohc_test
```

Run frontend tests through Bazel wrappers:

```bash
bazelisk test //srcs/frontend:frontend_layout_test
bazelisk test //srcs/frontend:frontend_unit_bazel_test
bazelisk test //srcs/frontend:frontend_e2e_bazel_test --test_output=errors
```

Run deploy artifact checks:

```bash
bazelisk test //deploy:deploy_artifacts_test
```

Run kind + helm E2E deploy test (manual target):

```bash
bazelisk test //deploy:kind_helm_e2e_test --test_output=errors
```

## OCI Image Commands

Build image loader targets:

```bash
bazelisk build //deploy:backend_image_tarball //deploy:frontend_image_tarball
```

Load images into local Docker daemon:

```bash
bazelisk run //deploy:backend_image_tarball
bazelisk run //deploy:frontend_image_tarball
```

## Notes

- Frontend npm tests are executed by Bazel `sh_test` wrappers.
- `kind_helm_e2e_test` expects local infra dependencies (`kind`, `kubectl`, `helm`, `docker`).
