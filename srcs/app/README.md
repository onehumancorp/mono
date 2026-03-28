# OHC App — Flutter Cross-Platform Client

## Identity

The `app` module is a cross-platform front-end for **One Human Corp** that runs on Android, iOS, macOS, Windows, Linux, and Web, providing a unified mobile and desktop experience for the CEO.

## Architecture

```text
srcs/app/
├── lib/
│   ├── main.dart           # App entry point
│   ├── router.dart         # GoRouter navigation + AppShell
│   ├── models/             # Domain models + package-local tests
│   ├── screens/            # Full-page UI screens + widget tests
│   └── services/           # API, auth, local manager, and service tests
├── e2e/                    # Playwright specs + screenshot capture runner
├── test/                   # Bazel wrapper scripts + desktop e2e test
├── android/                # Android-specific files
├── ios/                    # iOS-specific files
├── macos/                  # macOS-specific files
├── windows/                # Windows-specific files
├── linux/                  # Linux-specific files
└── web/                    # Web-specific files
```

## Quick Start

The app's day-to-day build, run, test, and screenshot flows are Bazel-native.
Those flows use the repository's Flutter toolchain and do not require a manual
Flutter SDK install.

### Prerequisites

- `bazelisk`
- `python3`
- `node` (for Playwright web e2e and screenshot capture)

```bash
# Serve the Bazel-built Flutter web app on http://127.0.0.1:8081
bazelisk run //srcs/app:start
```

## Developer Workflow

The wrapper scripts under `srcs/app/test/` are now thin Bazel front ends:

```bash
# All app unit + widget tests
srcs/app/test/flutter_unit_test.sh

# One specific test file
srcs/app/test/flutter_single_test.sh lib/services/local_manager_service_test.dart

# Per-package coverage checks (>= 90% each)
srcs/app/test/flutter_coverage_check.sh
```

Direct Bazel targets are available for non-wrapper workflows:

```bash
# Layout validation
bazelisk test //srcs/app:app_layout_test

# Desktop button-click regression coverage
bazelisk test //srcs/app:app_desktop_e2e_test

# Web Playwright end-to-end validation
bazelisk test //srcs/app:app_web_e2e_test
```

To regenerate documentation screenshots:

```bash
bazelisk run //srcs/app:capture_screenshots
```

The corresponding VS Code task is `App: Capture Flutter screenshots`.
The generated images are referenced from [docs/app/user-guide.md](../../docs/app/user-guide.md).

## Native Platform Note

The repository still contains Android, iOS, macOS, Windows, and Linux project
folders. The hermetic Bazel workflow in this repo currently standardizes the
test, local web serving, Playwright e2e, and documentation screenshot flows.
Native packaging for those OS-specific targets still depends on the appropriate
host platform SDKs when you need to produce native binaries.

## Configuration

The app connects to the OHC backend via the `BACKEND_URL` environment variable.
For local development, the default is `http://localhost:8080`.

## Shared Backend Logic

The app communicates with the OHC backend (`srcs/core` Rust library exposed
through the Go dashboard server).  All business logic — agent scheduling,
meeting rooms, chat integration — lives in `srcs/core` and is shared between:

- This Flutter app (via the REST/gRPC API)
- The Tauri desktop app (`srcs/desktop`)
- The web frontend (`srcs/frontend`)
