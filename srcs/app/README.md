# OHC App — Flutter Cross-Platform Client

## Identity
The `app` module is a cross-platform front-end for **One Human Corp** that runs on Android, iOS, macOS, Windows, Linux, and Web, providing a unified mobile and desktop experience for the CEO.

## Architecture

```
srcs/app/
├── lib/
│   ├── main.dart           # App entry point
│   ├── router.dart         # GoRouter navigation + AppShell
│   ├── models/             # Domain models (Agent, etc.)
│   ├── screens/            # Full-page UI screens
│   │   ├── login_screen.dart
│   │   ├── dashboard_screen.dart
│   │   ├── agents_screen.dart
│   │   ├── meetings_screen.dart
│   │   ├── chat_screen.dart
│   │   └── settings_screen.dart
│   ├── services/           # API & auth services (Riverpod providers)
│   │   ├── auth_service.dart
│   │   └── api_service.dart
│   └── widgets/            # Reusable UI components
├── android/                # Android-specific files
├── ios/                    # iOS-specific files
├── macos/                  # macOS-specific files
├── windows/                # Windows-specific files
├── linux/                  # Linux-specific files
└── web/                    # Web-specific files
```

## Quick Start

### Prerequisites
- Flutter SDK ≥ 3.3.0 ([install](https://docs.flutter.dev/get-started/install))
- Dart SDK ≥ 3.3.0 (bundled with Flutter)
- For Android: Android Studio + NDK
- For iOS/macOS: Xcode ≥ 15
- For Windows: Visual Studio 2022 (Desktop development with C++)

```bash
cd srcs/app

# Install dependencies
flutter pub get

# Run on any connected device / simulator
flutter run

# Target a specific platform
flutter run -d macos
flutter run -d windows
flutter run -d android
flutter run -d ios
flutter run -d chrome    # Web
```

## Developer Workflow

- **Run all tests:** `flutter test`
- **Format code:** `flutter format .`
- **Build all platforms:** `bazelisk build //...` (Integration in progress)

## Configuration

The app connects to the OHC backend via the `BACKEND_URL` environment variable.
For local development, the default is `http://localhost:8080`.

```bash
flutter run --dart-define=BACKEND_URL=http://localhost:8080
```

For production builds:

```bash
flutter build apk --dart-define=BACKEND_URL=https://api.yourcompany.com
flutter build ipa --dart-define=BACKEND_URL=https://api.yourcompany.com
flutter build macos --dart-define=BACKEND_URL=https://api.yourcompany.com
flutter build windows --dart-define=BACKEND_URL=https://api.yourcompany.com
```

## Shared Backend Logic

The app communicates with the OHC backend (`srcs/core` Rust library exposed
through the Go dashboard server).  All business logic — agent scheduling,
meeting rooms, chat integration — lives in `srcs/core` and is shared between:

- This Flutter app (via the REST/gRPC API)
- The Tauri desktop app (`srcs/desktop`)
- The web frontend (`srcs/frontend`)
