import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:mocktail/mocktail.dart';
import 'package:ohc_app/screens/agents_screen.dart';
import 'package:ohc_app/screens/ai_config_screen.dart';
import 'package:ohc_app/screens/channels_screen.dart';
import 'package:ohc_app/screens/chat_screen.dart';
import 'package:ohc_app/screens/dashboard_screen.dart';
import 'package:ohc_app/screens/login_screen.dart';
import 'package:ohc_app/screens/logs_screen.dart';
import 'package:ohc_app/screens/meetings_screen.dart';
import 'package:ohc_app/screens/security_screen.dart';
import 'package:ohc_app/screens/settings_screen.dart';
import 'package:ohc_app/screens/skills_screen.dart';
import 'package:ohc_app/services/api_service.dart';
import 'package:ohc_app/services/auth_service.dart';

class MockHttpClient extends Mock implements http.Client {}

class FakeUri extends Fake implements Uri {}

// Helper: wrap a widget with ProviderScope + MaterialApp.
Widget _wrap(Widget child,
    {List<Override> overrides = const []}) {
  return ProviderScope(
    overrides: overrides,
    child: MaterialApp(home: child),
  );
}

// A completed AsyncValue<AuthUser?> with a fake user.
const _fakeUser = AuthUser(
  id: 'u1',
  email: 'test@example.com',
  name: 'Test User',
  role: 'admin',
  organizationId: 'org-1',
  token: 'tok-test',
);

void main() {
  setUpAll(() {
    registerFallbackValue(FakeUri());
  });

  // ── LoginScreen ──────────────────────────────────────────────────────────

  group('LoginScreen', () {
    testWidgets('renders email/password fields and sign-in button',
        (tester) async {
      await tester.pumpWidget(_wrap(const LoginScreen()));
      await tester.pump();

      expect(find.byType(TextFormField), findsWidgets);
      expect(find.text('Sign In'), findsOneWidget);
    });

    testWidgets('shows validation error when form submitted empty',
        (tester) async {
      await tester.pumpWidget(_wrap(const LoginScreen()));
      await tester.pump();

      await tester.tap(find.text('Sign In'));
      await tester.pumpAndSettle();

      expect(find.text('Enter a valid email'), findsOneWidget);
    });
  });

  // ── SettingsScreen ────────────────────────────────────────────────────────

  group('SettingsScreen', () {
    testWidgets('renders without error when no user logged in', (tester) async {
      await tester.pumpWidget(_wrap(
        const SettingsScreen(),
        overrides: [
          authStateProvider.overrideWith(() => _FakeAuthNotifier(null)),
        ],
      ));
      await tester.pump();

      expect(find.text('Settings'), findsOneWidget);
      expect(find.text('Sign Out'), findsOneWidget);
    });

    testWidgets('renders user info when user is logged in', (tester) async {
      await tester.pumpWidget(_wrap(
        const SettingsScreen(),
        overrides: [
          authStateProvider.overrideWith(() => _FakeAuthNotifier(_fakeUser)),
        ],
      ));
      await tester.pump();

      expect(find.text('Test User'), findsOneWidget);
      expect(find.text('test@example.com'), findsOneWidget);
    });
  });

  // ── DashboardScreen ───────────────────────────────────────────────────────

  group('DashboardScreen', () {
    testWidgets('shows loading when API is null', (tester) async {
      await tester.pumpWidget(_wrap(
        const DashboardScreen(),
        overrides: [
          apiServiceProvider.overrideWithValue(null),
        ],
      ));
      await tester.pump();
      // With null API, the FutureProvider returns {} immediately (no loading).
      expect(find.byType(Scaffold), findsOneWidget);
    });

    testWidgets('shows dashboard data when API returns data', (tester) async {
      final mockClient = MockHttpClient();
      when(() => mockClient.get(any(), headers: any(named: 'headers')))
          .thenAnswer((_) async => http.Response(
              jsonEncode({
                'active_agents': 3,
                'pending_tasks': 7,
                'open_meetings': 1,
              }),
              200));

      final api = ApiService(
          baseUrl: 'http://localhost', token: 'tok', client: mockClient);

      await tester.pumpWidget(_wrap(
        const DashboardScreen(),
        overrides: [apiServiceProvider.overrideWithValue(api)],
      ));
      await tester.pumpAndSettle();

      expect(find.text('3'), findsOneWidget);
    });
  });

  // ── AgentsScreen ──────────────────────────────────────────────────────────

  group('AgentsScreen', () {
    testWidgets('shows empty state with null API', (tester) async {
      await tester.pumpWidget(_wrap(
        const AgentsScreen(),
        overrides: [apiServiceProvider.overrideWithValue(null)],
      ));
      await tester.pump();
      expect(find.byType(Scaffold), findsOneWidget);
    });

    testWidgets('shows agents list when API returns agents', (tester) async {
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
                }
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

    testWidgets('shows empty agents state when API returns empty list',
        (tester) async {
      final mockClient = MockHttpClient();
      when(() => mockClient.get(any(), headers: any(named: 'headers')))
          .thenAnswer(
              (_) async => http.Response(jsonEncode(<dynamic>[]), 200));

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

  // ── MeetingsScreen ────────────────────────────────────────────────────────

  group('MeetingsScreen', () {
    testWidgets('renders with null API', (tester) async {
      await tester.pumpWidget(_wrap(
        const MeetingsScreen(),
        overrides: [apiServiceProvider.overrideWithValue(null)],
      ));
      await tester.pump();
      expect(find.byType(Scaffold), findsOneWidget);
    });

    testWidgets('shows meetings when API returns data', (tester) async {
      final mockClient = MockHttpClient();
      when(() => mockClient.get(any(), headers: any(named: 'headers')))
          .thenAnswer((_) async => http.Response(
              jsonEncode([
                {'id': 'm1', 'name': 'Standup', 'participants': []}
              ]),
              200));

      final api = ApiService(
          baseUrl: 'http://localhost', token: 'tok', client: mockClient);

      await tester.pumpWidget(_wrap(
        const MeetingsScreen(),
        overrides: [apiServiceProvider.overrideWithValue(api)],
      ));
      await tester.pumpAndSettle();

      expect(find.text('Standup'), findsOneWidget);
    });
  });

  // ── ChatScreen ────────────────────────────────────────────────────────────

  group('ChatScreen', () {
    testWidgets('renders without error with null API', (tester) async {
      await tester.pumpWidget(_wrap(
        const ChatScreen(),
        overrides: [
          apiServiceProvider.overrideWithValue(null),
          authStateProvider.overrideWith(() => _FakeAuthNotifier(null)),
        ],
      ));
      await tester.pump();
      expect(find.byType(Scaffold), findsOneWidget);
    });
  });

  // ── ChannelsScreen ────────────────────────────────────────────────────────

  group('ChannelsScreen', () {
    testWidgets('renders with null API', (tester) async {
      await tester.pumpWidget(_wrap(
        const ChannelsScreen(),
        overrides: [apiServiceProvider.overrideWithValue(null)],
      ));
      await tester.pump();
      expect(find.byType(Scaffold), findsOneWidget);
    });

    testWidgets('shows channels when API returns data', (tester) async {
      final mockClient = MockHttpClient();
      when(() => mockClient.get(any(), headers: any(named: 'headers')))
          .thenAnswer((_) async => http.Response(
              jsonEncode([
                {
                  'id': 'ch1',
                  'organization_id': 'org-1',
                  'name': 'general',
                  'backend': 'slack',
                  'config': <String, String>{},
                  'enabled': true,
                  'created_at': '2025-01-01T00:00:00Z',
                }
              ]),
              200));

      final api = ApiService(
          baseUrl: 'http://localhost', token: 'tok', client: mockClient);

      await tester.pumpWidget(_wrap(
        const ChannelsScreen(),
        overrides: [apiServiceProvider.overrideWithValue(api)],
      ));
      await tester.pumpAndSettle();

      expect(find.text('general'), findsOneWidget);
    });
  });

  // ── AiConfigScreen ────────────────────────────────────────────────────────

  group('AiConfigScreen', () {
    testWidgets('renders with null API', (tester) async {
      await tester.pumpWidget(_wrap(
        const AiConfigScreen(),
        overrides: [apiServiceProvider.overrideWithValue(null)],
      ));
      await tester.pump();
      expect(find.byType(Scaffold), findsOneWidget);
    });

    testWidgets('shows AI providers when API returns data', (tester) async {
      final mockClient = MockHttpClient();
      when(() => mockClient.get(any(), headers: any(named: 'headers')))
          .thenAnswer((_) async => http.Response(
              jsonEncode([
                {
                  'id': 'p1',
                  'name': 'OpenAI',
                  'base_url': 'https://api.openai.com/v1',
                  'api_key': 'sk-...',
                  'models': ['gpt-4'],
                  'is_official': true,
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

  group('SkillsScreen', () {
    testWidgets('renders with null API', (tester) async {
      await tester.pumpWidget(_wrap(
        const SkillsScreen(),
        overrides: [apiServiceProvider.overrideWithValue(null)],
      ));
      await tester.pump();
      expect(find.byType(Scaffold), findsOneWidget);
    });

    testWidgets('shows skills when API returns data', (tester) async {
      final mockClient = MockHttpClient();
      when(() => mockClient.get(any(), headers: any(named: 'headers')))
          .thenAnswer((_) async => http.Response(
              jsonEncode([
                {
                  'name': 'web_search',
                  'version': '1.0.0',
                  'description': 'Search the web',
                  'category': 'official',
                  'installed': true,
                  'enabled': true,
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

      expect(find.text('web_search'), findsOneWidget);
    });
  });

  // ── LogsScreen ────────────────────────────────────────────────────────────

  group('LogsScreen', () {
    testWidgets('renders with null API', (tester) async {
      await tester.pumpWidget(_wrap(
        const LogsScreen(),
        overrides: [apiServiceProvider.overrideWithValue(null)],
      ));
      await tester.pump();
      expect(find.byType(Scaffold), findsOneWidget);
    });

    testWidgets('shows logs when API returns data', (tester) async {
      final mockClient = MockHttpClient();
      when(() => mockClient.get(any(), headers: any(named: 'headers')))
          .thenAnswer((_) async => http.Response(
              jsonEncode(['2025-01-01 INFO Starting up', '2025-01-01 INFO Ready']),
              200));

      final api = ApiService(
          baseUrl: 'http://localhost', token: 'tok', client: mockClient);

      await tester.pumpWidget(_wrap(
        const LogsScreen(),
        overrides: [apiServiceProvider.overrideWithValue(api)],
      ));
      await tester.pumpAndSettle();

      expect(find.textContaining('Starting up'), findsOneWidget);
    });
  });

  // ── SecurityScreen ────────────────────────────────────────────────────────

  group('SecurityScreen', () {
    testWidgets('renders with null API', (tester) async {
      await tester.pumpWidget(_wrap(
        const SecurityScreen(),
        overrides: [apiServiceProvider.overrideWithValue(null)],
      ));
      await tester.pump();
      expect(find.byType(Scaffold), findsOneWidget);
    });

    testWidgets('shows all-clear when no issues', (tester) async {
      final mockClient = MockHttpClient();
      when(() => mockClient.get(any(), headers: any(named: 'headers')))
          .thenAnswer(
              (_) async => http.Response(jsonEncode(<dynamic>[]), 200));

      final api = ApiService(
          baseUrl: 'http://localhost', token: 'tok', client: mockClient);

      await tester.pumpWidget(_wrap(
        const SecurityScreen(),
        overrides: [apiServiceProvider.overrideWithValue(api)],
      ));
      await tester.pumpAndSettle();

      expect(find.text('No security issues found'), findsOneWidget);
    });

    testWidgets('shows issues when API returns issues', (tester) async {
      final mockClient = MockHttpClient();
      when(() => mockClient.get(any(), headers: any(named: 'headers')))
          .thenAnswer((_) async => http.Response(
              jsonEncode([
                {
                  'id': 'i1',
                  'title': 'Weak password',
                  'description': 'Use a stronger password',
                  'severity': 'high',
                  'fixable': true,
                  'fixed': false,
                  'category': 'auth',
                }
              ]),
              200));

      final api = ApiService(
          baseUrl: 'http://localhost', token: 'tok', client: mockClient);

      await tester.pumpWidget(_wrap(
        const SecurityScreen(),
        overrides: [apiServiceProvider.overrideWithValue(api)],
      ));
      await tester.pumpAndSettle();

      expect(find.text('Weak password'), findsOneWidget);
    });
  });
}

// ── Fake AuthNotifier ────────────────────────────────────────────────────────

class _FakeAuthNotifier extends AuthNotifier {
  final AuthUser? _user;
  _FakeAuthNotifier(this._user);

  @override
  Future<AuthUser?> build() async => _user;
}
