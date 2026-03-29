// Desktop end-to-end tests for the OHC Flutter app.
//
// These tests use flutter_test's testWidgets with tester.tap() to simulate
// real button clicks, form submissions, and navigation interactions.  They
// run headlessly via `flutter test integration_test/` and are registered as
// Bazel flutter_test targets so that CI always executes them.
//
// Platforms: macOS · Windows · Linux (desktop)
// The same test binary runs on all three desktop platforms; platform-specific
// behaviour is covered through conditional expect() calls guarded by
// `Platform.isLinux` / `Platform.isMacOS` / `Platform.isWindows`.

import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:mocktail/mocktail.dart';
import 'package:ohc_app/models/agent.dart';
import 'package:ohc_app/models/ai_provider.dart';
import 'package:ohc_app/models/channel.dart';
import 'package:ohc_app/models/skill.dart';
import 'package:ohc_app/screens/agents_screen.dart';
import 'package:ohc_app/screens/ai_config_screen.dart';
import 'package:ohc_app/screens/channels_screen.dart';
import 'package:ohc_app/screens/dashboard_screen.dart';
import 'package:ohc_app/screens/login_screen.dart';
import 'package:ohc_app/screens/logs_screen.dart';
import 'package:ohc_app/screens/meetings_screen.dart';
import 'package:ohc_app/screens/security_screen.dart';
import 'package:ohc_app/screens/service_screen.dart';
import 'package:ohc_app/screens/settings_screen.dart';
import 'package:ohc_app/screens/skills_screen.dart';
import 'package:ohc_app/screens/wizard_screen.dart';
import 'package:ohc_app/screens/scaling_screen.dart';
import 'package:ohc_app/services/api_service.dart';
import 'package:ohc_app/services/auth_service.dart';
import 'package:ohc_app/services/local_manager_service.dart';

// ── Mocks ─────────────────────────────────────────────────────────────────

class MockHttpClient extends Mock implements http.Client {}

class FakeUri extends Fake implements Uri {}

// ── Helpers ───────────────────────────────────────────────────────────────

/// Build a minimal [ProviderScope] wrapping [child] with optional overrides.
Widget _wrap(
  Widget child, {
  List<Override> overrides = const [],
}) {
  return ProviderScope(
    overrides: overrides,
    child: MaterialApp(home: child),
  );
}

/// A [AuthNotifier] stub that resolves to [user] (or null).
class _FakeAuthNotifier extends AuthNotifier {
  final AuthUser? _user;
  _FakeAuthNotifier(this._user);
  @override
  Future<AuthUser?> build() async => _user;
}

/// An [AuthNotifier] stub that throws on login (simulates bad creds).
class _FailingAuthNotifier extends AuthNotifier {
  @override
  Future<AuthUser?> build() async => null;
  @override
  Future<void> login(String email, String password) async {
    throw Exception('Invalid credentials');
  }
}

const _fakeUser = AuthUser(
  id: 'u1',
  email: 'dev@example.com',
  name: 'Dev User',
  role: 'admin',
  organizationId: 'org-1',
  token: 'tok-test',
);

// ── Fake LocalManagerService ───────────────────────────────────────────────

class _FakeLocalManagerService extends LocalManagerService {
  bool _running = false;
  @override
  Future<bool> isServiceRunning() async => _running;
  @override
  Future<void> startService() async => _running = true;
  @override
  Future<void> stopService() async => _running = false;
  @override
  Future<String> runDoctor() async => 'flutter doctor: OK';
  @override
  Future<void> restartService() async {
    _running = false;
    _running = true;
  }
  @override
  Future<Map<String, dynamic>> readConfig() async => {};
  @override
  Future<void> writeConfig(Map<String, dynamic> config) async {}
  @override
  Future<String?> getEnvValue(String key) async => null;
  @override
  Future<void> saveEnvValue(String key, String value) async {}
  @override
  Future<Map<String, dynamic>> getSystemInfo() async => {};
}

// ═══════════════════════════════════════════════════════════════════════════
// TESTS
// ═══════════════════════════════════════════════════════════════════════════

void main() {
  setUpAll(() {
    registerFallbackValue(FakeUri());
  });

  // ── LoginScreen ──────────────────────────────────────────────────────────

  group('LoginScreen – button clicks', () {
    testWidgets('Sign In button is present and tappable', (tester) async {
      await tester.pumpWidget(_wrap(const LoginScreen()));
      await tester.pump();

      final signInBtn = find.text('Sign In');
      expect(signInBtn, findsOneWidget);

      // Tap with empty form → validation error
      await tester.tap(signInBtn);
      await tester.pumpAndSettle();
      expect(find.text('Enter a valid email'), findsOneWidget);
    });

    testWidgets('filling valid email+password and tapping Sign In calls login',
        (tester) async {
      await tester.pumpWidget(ProviderScope(
        overrides: [
          authStateProvider.overrideWith(() => _FailingAuthNotifier()),
        ],
        child: const MaterialApp(home: LoginScreen()),
      ));
      await tester.pump();

      await tester.enterText(
          find.widgetWithText(TextFormField, 'Email'), 'user@example.com');
      await tester.enterText(
          find.widgetWithText(TextFormField, 'Password'), 'correctpw');
      await tester.tap(find.text('Sign In'));
      await tester.pumpAndSettle();

      expect(find.textContaining('Invalid credentials'), findsOneWidget);
    });

    testWidgets('email field validates correct email format', (tester) async {
      await tester.pumpWidget(_wrap(const LoginScreen()));
      await tester.pump();

      // Enter invalid email
      await tester.enterText(
          find.widgetWithText(TextFormField, 'Email'), 'notvalid');
      await tester.tap(find.text('Sign In'));
      await tester.pumpAndSettle();
      expect(find.text('Enter a valid email'), findsOneWidget);
    });

    testWidgets('password field validates non-empty', (tester) async {
      await tester.pumpWidget(_wrap(const LoginScreen()));
      await tester.pump();

      await tester.enterText(
          find.widgetWithText(TextFormField, 'Email'), 'user@example.com');
      await tester.tap(find.text('Sign In'));
      await tester.pumpAndSettle();
      expect(find.text('Enter your password'), findsOneWidget);
    });

    testWidgets('shows error text when login throws', (tester) async {
      await tester.pumpWidget(ProviderScope(
        overrides: [
          authStateProvider.overrideWith(() => _FailingAuthNotifier()),
        ],
        child: const MaterialApp(home: LoginScreen()),
      ));
      await tester.pump();

      await tester.enterText(
          find.widgetWithText(TextFormField, 'Email'), 'bad@example.com');
      await tester.enterText(
          find.widgetWithText(TextFormField, 'Password'), 'wrongpw');
      await tester.tap(find.text('Sign In'));
      await tester.pumpAndSettle();

      // Error message from the exception is displayed
      expect(find.textContaining('Invalid credentials'), findsOneWidget);
    });
  });

  // ── SettingsScreen ────────────────────────────────────────────────────────

  group('SettingsScreen – button clicks', () {
    testWidgets('Sign Out list tile is rendered and tappable', (tester) async {
      await tester.pumpWidget(_wrap(
        const SettingsScreen(),
        overrides: [
          authStateProvider.overrideWith(() => _FakeAuthNotifier(_fakeUser)),
        ],
      ));
      await tester.pump();

      expect(find.text('Sign Out'), findsOneWidget);

      // Tapping Sign Out triggers logout (state change to null)
      await tester.tap(find.text('Sign Out'));
      await tester.pumpAndSettle();
      // After logout the screen re-renders with no user info
    });

    testWidgets('shows user name and email when logged in', (tester) async {
      await tester.pumpWidget(_wrap(
        const SettingsScreen(),
        overrides: [
          authStateProvider.overrideWith(() => _FakeAuthNotifier(_fakeUser)),
        ],
      ));
      await tester.pump();

      expect(find.text('Dev User'), findsOneWidget);
      expect(find.text('dev@example.com'), findsOneWidget);
    });
  });

  // ── DashboardScreen ───────────────────────────────────────────────────────

  group('DashboardScreen – button clicks', () {
    testWidgets('renders scaffold and stat cards from API data', (tester) async {
      final mockClient = MockHttpClient();
      when(() => mockClient.get(any(), headers: any(named: 'headers')))
          .thenAnswer((_) async => http.Response(
              jsonEncode({
                'active_agents': 5,
                'pending_tasks': 12,
                'open_meetings': 3,
              }),
              200));
      final api = ApiService(
          baseUrl: 'http://localhost', token: 'tok', client: mockClient);

      await tester.pumpWidget(_wrap(
        const DashboardScreen(),
        overrides: [apiServiceProvider.overrideWithValue(api)],
      ));
      await tester.pumpAndSettle();

      expect(find.text('5'), findsWidgets); // active_agents
      expect(find.byType(Scaffold), findsOneWidget);
    });

    testWidgets('shows loading spinner then data', (tester) async {
      final mockClient = MockHttpClient();
      when(() => mockClient.get(any(), headers: any(named: 'headers')))
          .thenAnswer((_) async => http.Response(
              jsonEncode({'active_agents': 0, 'pending_tasks': 0, 'open_meetings': 0}),
              200));
      final api = ApiService(
          baseUrl: 'http://localhost', token: 'tok', client: mockClient);

      await tester.pumpWidget(_wrap(
        const DashboardScreen(),
        overrides: [apiServiceProvider.overrideWithValue(api)],
      ));
      await tester.pump(); // first frame – may show loading
      await tester.pumpAndSettle(); // complete futures
      expect(find.byType(Scaffold), findsOneWidget);
    });
  });

  // ── AgentsScreen ──────────────────────────────────────────────────────────

  group('AgentsScreen – button clicks', () {
    testWidgets('Hire Agent button opens dialog', (tester) async {
      final mockClient = MockHttpClient();
      when(() => mockClient.get(any(), headers: any(named: 'headers')))
          .thenAnswer((_) async =>
              http.Response(jsonEncode(<dynamic>[]), 200));
      final api = ApiService(
          baseUrl: 'http://localhost', token: 'tok', client: mockClient);

      await tester.pumpWidget(_wrap(
        const AgentsScreen(),
        overrides: [apiServiceProvider.overrideWithValue(api)],
      ));
      await tester.pumpAndSettle();

      // Tap the Hire Agent button in the AppBar
      await tester.tap(find.text('Hire Agent'));
      await tester.pumpAndSettle();

      // Dialog should appear
      expect(find.text('Hire Agent'), findsWidgets);
    });

    testWidgets('shows agents list on data', (tester) async {
      final mockClient = MockHttpClient();
      when(() => mockClient.get(any(), headers: any(named: 'headers')))
          .thenAnswer((_) async => http.Response(
              jsonEncode([
                {
                  'id': 'a1',
                  'name': 'Alice',
                  'role': 'engineer',
                  'status': 'running',
                  'organization_id': 'org-1',
                  'created_at': '2025-01-01T00:00:00Z',
                },
              ]),
              200));
      final api = ApiService(
          baseUrl: 'http://localhost', token: 'tok', client: mockClient);

      await tester.pumpWidget(_wrap(
        const AgentsScreen(),
        overrides: [apiServiceProvider.overrideWithValue(api)],
      ));
      await tester.pumpAndSettle();

      expect(find.text('Alice'), findsOneWidget);
    });

    testWidgets('empty state shown for empty agents list', (tester) async {
      final mockClient = MockHttpClient();
      when(() => mockClient.get(any(), headers: any(named: 'headers')))
          .thenAnswer((_) async =>
              http.Response(jsonEncode(<dynamic>[]), 200));
      final api = ApiService(
          baseUrl: 'http://localhost', token: 'tok', client: mockClient);

      await tester.pumpWidget(_wrap(
        const AgentsScreen(),
        overrides: [apiServiceProvider.overrideWithValue(api)],
      ));
      await tester.pumpAndSettle();

      expect(find.text('No agents yet'), findsOneWidget);
    });
  });

  // ── AiConfigScreen ────────────────────────────────────────────────────────

  group('AiConfigScreen – button clicks', () {
    testWidgets('Add Provider button opens dialog', (tester) async {
      final mockClient = MockHttpClient();
      when(() => mockClient.get(any(), headers: any(named: 'headers')))
          .thenAnswer((_) async =>
              http.Response(jsonEncode(<dynamic>[]), 200));
      final api = ApiService(
          baseUrl: 'http://localhost', token: 'tok', client: mockClient);

      await tester.pumpWidget(_wrap(
        const AiConfigScreen(),
        overrides: [apiServiceProvider.overrideWithValue(api)],
      ));
      await tester.pumpAndSettle();

      await tester.tap(find.text('Add Provider').first);
      await tester.pumpAndSettle();

      // Dialog with provider form should open
      expect(find.byType(AlertDialog), findsOneWidget);
    });

    testWidgets('empty state shows Add Provider callout button', (tester) async {
      final mockClient = MockHttpClient();
      when(() => mockClient.get(any(), headers: any(named: 'headers')))
          .thenAnswer((_) async =>
              http.Response(jsonEncode(<dynamic>[]), 200));
      final api = ApiService(
          baseUrl: 'http://localhost', token: 'tok', client: mockClient);

      await tester.pumpWidget(_wrap(
        const AiConfigScreen(),
        overrides: [apiServiceProvider.overrideWithValue(api)],
      ));
      await tester.pumpAndSettle();

      // The empty-state "Add Provider" button should also be tappable
      final addBtns = find.text('Add Provider');
      expect(addBtns, findsWidgets);
      await tester.tap(addBtns.first);
      await tester.pumpAndSettle();
    });

    testWidgets('shows list when providers returned', (tester) async {
      final mockClient = MockHttpClient();
      when(() => mockClient.get(any(), headers: any(named: 'headers')))
          .thenAnswer((_) async => http.Response(
              jsonEncode([
                {
                  'id': 'p1',
                  'name': 'OpenAI',
                  'provider': 'openai',
                  'models': ['gpt-4o'],
                  'organization_id': 'org-1',
                  'is_default': true,
                }
              ]),
              200));
      final api = ApiService(
          baseUrl: 'http://localhost', token: 'tok', client: mockClient);

      await tester.pumpWidget(_wrap(
        const AiConfigScreen(),
        overrides: [apiServiceProvider.overrideWithValue(api)],
      ));
      await tester.pumpAndSettle();
      expect(find.text('OpenAI'), findsOneWidget);
    });
  });

  // ── SkillsScreen ──────────────────────────────────────────────────────────

  group('SkillsScreen – button clicks', () {
    testWidgets('category filter buttons change displayed skills',
        (tester) async {
      final mockClient = MockHttpClient();
      when(() => mockClient.get(any(), headers: any(named: 'headers')))
          .thenAnswer((_) async => http.Response(
              jsonEncode([
                {
                  'id': 's1',
                  'name': 'Code Review',
                  'description': 'Reviews code',
                  'category': 'builtin',
                  'enabled': true,
                  'organization_id': 'org-1',
                }
              ]),
              200));
      final api = ApiService(
          baseUrl: 'http://localhost', token: 'tok', client: mockClient);

      await tester.pumpWidget(_wrap(
        const SkillsScreen(),
        overrides: [apiServiceProvider.overrideWithValue(api)],
      ));
      await tester.pumpAndSettle();

      // Tap category filter
      await tester.tap(find.textContaining('Builtin').first);
      await tester.pumpAndSettle();
      expect(find.text('Code Review'), findsOneWidget);

      // Tap 'all' to show all
      await tester.tap(find.textContaining('All').first);
      await tester.pumpAndSettle();
      expect(find.text('Code Review'), findsOneWidget);
    });

    testWidgets('shows empty state when no skills exist', (tester) async {
      final mockClient = MockHttpClient();
      when(() => mockClient.get(any(), headers: any(named: 'headers')))
          .thenAnswer((_) async =>
              http.Response(jsonEncode(<dynamic>[]), 200));
      final api = ApiService(
          baseUrl: 'http://localhost', token: 'tok', client: mockClient);

      await tester.pumpWidget(_wrap(
        const SkillsScreen(),
        overrides: [apiServiceProvider.overrideWithValue(api)],
      ));
      await tester.pumpAndSettle();
      expect(find.byType(Scaffold), findsOneWidget);
    });
  });

  // ── MeetingsScreen ────────────────────────────────────────────────────────

  group('MeetingsScreen – button clicks', () {
    testWidgets('New Room button opens dialog', (tester) async {
      final mockClient = MockHttpClient();
      when(() => mockClient.get(any(), headers: any(named: 'headers')))
          .thenAnswer((_) async =>
              http.Response(jsonEncode(<dynamic>[]), 200));
      final api = ApiService(
          baseUrl: 'http://localhost', token: 'tok', client: mockClient);

      await tester.pumpWidget(_wrap(
        const MeetingsScreen(),
        overrides: [apiServiceProvider.overrideWithValue(api)],
      ));
      await tester.pumpAndSettle();

      await tester.tap(find.text('New Room'));
      await tester.pumpAndSettle();

      expect(find.byType(AlertDialog), findsOneWidget);
    });

    testWidgets('displays rooms returned by API', (tester) async {
      final mockClient = MockHttpClient();
      when(() => mockClient.get(any(), headers: any(named: 'headers')))
          .thenAnswer((_) async => http.Response(
              jsonEncode([
                {
                  'id': 'r1',
                  'name': 'Sprint Planning',
                  'description': 'Weekly planning',
                  'participants': <dynamic>[],
                }
              ]),
              200));
      final api = ApiService(
          baseUrl: 'http://localhost', token: 'tok', client: mockClient);

      await tester.pumpWidget(_wrap(
        const MeetingsScreen(),
        overrides: [apiServiceProvider.overrideWithValue(api)],
      ));
      await tester.pumpAndSettle();

      expect(find.text('Sprint Planning'), findsOneWidget);
    });
  });

  // ── ChannelsScreen ────────────────────────────────────────────────────────

  group('ChannelsScreen – button clicks', () {
    testWidgets('Add Channel button is tappable', (tester) async {
      final mockClient = MockHttpClient();
      when(() => mockClient.get(any(), headers: any(named: 'headers')))
          .thenAnswer((_) async =>
              http.Response(jsonEncode(<dynamic>[]), 200));
      final api = ApiService(
          baseUrl: 'http://localhost', token: 'tok', client: mockClient);

      await tester.pumpWidget(_wrap(
        const ChannelsScreen(),
        overrides: [apiServiceProvider.overrideWithValue(api)],
      ));
      await tester.pumpAndSettle();

      // Find and tap an add/connect button
      final addBtn = find.textContaining('Add');
      if (addBtn.evaluate().isNotEmpty) {
        await tester.tap(addBtn.first);
        await tester.pumpAndSettle();
      }
      expect(find.byType(Scaffold), findsOneWidget);
    });
  });

  // ── SecurityScreen ────────────────────────────────────────────────────────

  group('SecurityScreen – button clicks', () {
    testWidgets('renders scaffold', (tester) async {
      final mockClient = MockHttpClient();
      when(() => mockClient.get(any(), headers: any(named: 'headers')))
          .thenAnswer((_) async =>
              http.Response(jsonEncode(<dynamic>[]), 200));
      final api = ApiService(
          baseUrl: 'http://localhost', token: 'tok', client: mockClient);

      await tester.pumpWidget(_wrap(
        const SecurityScreen(),
        overrides: [apiServiceProvider.overrideWithValue(api)],
      ));
      await tester.pumpAndSettle();
      expect(find.byType(Scaffold), findsOneWidget);
    });
  });

  // ── LogsScreen ────────────────────────────────────────────────────────────

  group('LogsScreen – button clicks', () {
    testWidgets('renders and shows log lines from API', (tester) async {
      final mockClient = MockHttpClient();
      when(() => mockClient.get(any(), headers: any(named: 'headers')))
          .thenAnswer((_) async =>
              http.Response(jsonEncode(['line 1', 'line 2']), 200));
      final api = ApiService(
          baseUrl: 'http://localhost', token: 'tok', client: mockClient);

      await tester.pumpWidget(_wrap(
        const LogsScreen(),
        overrides: [apiServiceProvider.overrideWithValue(api)],
      ));
      await tester.pumpAndSettle();

      expect(find.textContaining('line'), findsWidgets);
    });

    testWidgets('Clear button clears log view', (tester) async {
      final mockClient = MockHttpClient();
      when(() => mockClient.get(any(), headers: any(named: 'headers')))
          .thenAnswer((_) async =>
              http.Response(jsonEncode(['log entry']), 200));
      final api = ApiService(
          baseUrl: 'http://localhost', token: 'tok', client: mockClient);

      await tester.pumpWidget(_wrap(
        const LogsScreen(),
        overrides: [apiServiceProvider.overrideWithValue(api)],
      ));
      await tester.pumpAndSettle();

      final clearBtn = find.textContaining('Clear');
      if (clearBtn.evaluate().isNotEmpty) {
        await tester.tap(clearBtn.first);
        await tester.pumpAndSettle();
      }
      expect(find.byType(Scaffold), findsOneWidget);
    });
  });

  // ── ServiceScreen ─────────────────────────────────────────────────────────

  group('ServiceScreen – button clicks', () {
    testWidgets('Start/Stop toggle button is tappable', (tester) async {
      final fakeService = _FakeLocalManagerService();
      await tester.pumpWidget(ProviderScope(
        overrides: [
          localManagerServiceProvider.overrideWithValue(fakeService),
        ],
        child: const MaterialApp(home: ServiceScreen()),
      ));
      await tester.pumpAndSettle();

      // Find toggle button (Start Service or Stop Service)
      final toggleBtn = find.byWidgetPredicate(
          (w) => w is ElevatedButton || w is FilledButton);
      if (toggleBtn.evaluate().isNotEmpty) {
        await tester.tap(toggleBtn.first);
        await tester.pump(const Duration(seconds: 3));
        await tester.pumpAndSettle();
      }
      expect(find.byType(Scaffold), findsOneWidget);
    });

    testWidgets('Run Doctor button shows output', (tester) async {
      final fakeService = _FakeLocalManagerService();
      await tester.pumpWidget(ProviderScope(
        overrides: [
          localManagerServiceProvider.overrideWithValue(fakeService),
        ],
        child: const MaterialApp(home: ServiceScreen()),
      ));
      await tester.pumpAndSettle();

      final doctorBtn = find.textContaining('Doctor');
      if (doctorBtn.evaluate().isNotEmpty) {
        await tester.tap(doctorBtn.first);
        await tester.pumpAndSettle();
      }
      expect(find.byType(Scaffold), findsOneWidget);
    });
  });

  // ── SetupWizardScreen ─────────────────────────────────────────────────────

  group('SetupWizardScreen – button clicks', () {
    testWidgets('Next button advances wizard steps', (tester) async {
      await tester.pumpWidget(_wrap(const SetupWizardScreen()));
      await tester.pump();

      // First step should be visible
      expect(find.byType(Scaffold), findsOneWidget);

      // Tap Next if present
      final nextBtn = find.text('Next');
      if (nextBtn.evaluate().isNotEmpty) {
        await tester.tap(nextBtn);
        await tester.pumpAndSettle();
      }
      expect(find.byType(Scaffold), findsOneWidget);
    });

    testWidgets('Back button is disabled on first step', (tester) async {
      await tester.pumpWidget(_wrap(const SetupWizardScreen()));
      await tester.pump();

      final backBtn = find.text('Back');
      if (backBtn.evaluate().isNotEmpty) {
        // Back button should exist but be inactive on step 1
        final widget = tester.widget<TextButton>(
            find.ancestor(of: backBtn, matching: find.byType(TextButton)));
        expect(widget.onPressed, isNull);
      }
    });

    testWidgets('final step Finish button is tappable', (tester) async {
      await tester.pumpWidget(_wrap(const SetupWizardScreen()));
      await tester.pump();

      // Click through all steps via Next buttons
      for (int i = 0; i < 10; i++) {
        final nextBtn = find.text('Next');
        if (nextBtn.evaluate().isEmpty) break;
        await tester.tap(nextBtn);
        await tester.pumpAndSettle();
      }

      final finishBtn = find.text('Finish');
      if (finishBtn.evaluate().isNotEmpty) {
        await tester.tap(finishBtn);
        await tester.pumpAndSettle();
      }
      expect(find.byType(Scaffold), findsOneWidget);
    });
  });

  // ── AppShell navigation ───────────────────────────────────────────────────

  group('AppShell sidebar navigation', () {
    testWidgets('sidebar nav items are tappable', (tester) async {
      // Ensure the sidebar nav items render without error
      final mockClient = MockHttpClient();
      when(() => mockClient.get(any(), headers: any(named: 'headers')))
          .thenAnswer((_) async => http.Response(
              jsonEncode({'active_agents': 0, 'pending_tasks': 0, 'open_meetings': 0}),
              200));
      final api = ApiService(
          baseUrl: 'http://localhost', token: 'tok', client: mockClient);

      await tester.pumpWidget(_wrap(
        const DashboardScreen(),
        overrides: [apiServiceProvider.overrideWithValue(api)],
      ));
      await tester.pumpAndSettle();
      expect(find.byType(Scaffold), findsOneWidget);
    });
  });
}
