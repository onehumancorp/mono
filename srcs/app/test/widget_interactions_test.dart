import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:mocktail/mocktail.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:ohc_app/screens/agents_screen.dart';
import 'package:ohc_app/screens/ai_config_screen.dart';
import 'package:ohc_app/screens/channels_screen.dart';
import 'package:ohc_app/screens/chat_screen.dart';
import 'package:ohc_app/screens/login_screen.dart';
import 'package:ohc_app/screens/meetings_screen.dart';
import 'package:ohc_app/screens/skills_screen.dart';
import 'package:ohc_app/services/api_service.dart';
import 'package:ohc_app/services/auth_service.dart';
import 'package:ohc_app/services/centrifuge_service.dart';

class MockHttpClient extends Mock implements http.Client {}

class FakeUri extends Fake implements Uri {}

Widget _wrap(Widget child, {List<Override> overrides = const []}) {
  return ProviderScope(
    overrides: overrides,
    child: MaterialApp(home: child),
  );
}

ApiService _mockApi(MockHttpClient client) =>
    ApiService(baseUrl: 'http://localhost', token: 'tok', client: client);

void main() {
  setUpAll(() {
    registerFallbackValue(FakeUri());
    SharedPreferences.setMockInitialValues({});
  });

  // ── AuthNotifier ──────────────────────────────────────────────────────────

  group('AuthNotifier', () {
    test('build returns null when no token in prefs', () async {
      SharedPreferences.setMockInitialValues({});
      final container = ProviderContainer(overrides: [
        backendUrlProvider.overrideWithValue('http://localhost'),
      ]);
      addTearDown(container.dispose);

      final state = await container.read(authStateProvider.future);
      expect(state, isNull);
    });

    test('build returns null when token exists but server unreachable',
        () async {
      SharedPreferences.setMockInitialValues({'flutter.auth_token': 'bad-tok'});
      final container = ProviderContainer(overrides: [
        backendUrlProvider
            .overrideWithValue('http://127.0.0.1:1'), // unreachable
      ]);
      addTearDown(container.dispose);

      final state = await container.read(authStateProvider.future);
      expect(state, isNull);
    });

    test('login stores token and updates state', () async {
      SharedPreferences.setMockInitialValues({});
      final mockClient = MockHttpClient();
      when(() => mockClient.post(
            any(),
            headers: any(named: 'headers'),
            body: any(named: 'body'),
          )).thenAnswer((_) async => http.Response(
            jsonEncode({
              'token': 'new-token',
              'user': {
                'id': 'u1',
                'email': 'a@b.com',
                'name': 'Alice',
                'role': 'admin',
                'organization_id': 'org-1',
              },
            }),
            200,
          ));

      final container = ProviderContainer(overrides: [
        backendUrlProvider.overrideWithValue('http://localhost'),
        authServiceProvider.overrideWithValue(
            AuthService(baseUrl: 'http://localhost', client: mockClient)),
      ]);
      addTearDown(container.dispose);

      await container.read(authStateProvider.notifier).login('a@b.com', 'pw');
      final user = container.read(authStateProvider).valueOrNull;
      expect(user?.email, 'a@b.com');
      expect(user?.token, 'new-token');
    });

    test('logout clears state', () async {
      SharedPreferences.setMockInitialValues({});
      final mockClient = MockHttpClient();
      // Set up login first
      when(() => mockClient.post(
            any(),
            headers: any(named: 'headers'),
            body: any(named: 'body'),
          )).thenAnswer((_) async => http.Response(
            jsonEncode({
              'token': 'tok-123',
              'user': {
                'id': 'u1',
                'email': 'a@b.com',
                'name': 'A',
                'role': 'admin',
                'organization_id': 'o1',
              },
            }),
            200,
          ));

      final container = ProviderContainer(overrides: [
        backendUrlProvider.overrideWithValue('http://localhost'),
        authServiceProvider.overrideWithValue(
            AuthService(baseUrl: 'http://localhost', client: mockClient)),
      ]);
      addTearDown(container.dispose);

      await container.read(authStateProvider.notifier).login('a@b.com', 'pw');
      expect(container.read(authStateProvider).valueOrNull, isNotNull);

      await container.read(authStateProvider.notifier).logout();
      expect(container.read(authStateProvider).valueOrNull, isNull);
    });

    test('logout with no user still clears state', () async {
      SharedPreferences.setMockInitialValues({});
      final container = ProviderContainer(overrides: [
        backendUrlProvider.overrideWithValue('http://localhost'),
      ]);
      addTearDown(container.dispose);

      await container.read(authStateProvider.future);
      await container.read(authStateProvider.notifier).logout();
      expect(container.read(authStateProvider).valueOrNull, isNull);
    });
  });

  // ── LoginScreen interactions ──────────────────────────────────────────────

  group('LoginScreen interactions', () {
    testWidgets('shows password validation error when only password is empty',
        (tester) async {
      await tester.pumpWidget(_wrap(const LoginScreen()));
      await tester.pump();

      await tester.enterText(
          find.widgetWithText(TextFormField, 'Email'), 'test@example.com');
      await tester.tap(find.text('Sign In'));
      await tester.pumpAndSettle();

      expect(find.text('Enter your password'), findsOneWidget);
    });

    testWidgets('shows email validation error for invalid email',
        (tester) async {
      await tester.pumpWidget(_wrap(const LoginScreen()));
      await tester.pump();

      await tester.enterText(
          find.widgetWithText(TextFormField, 'Email'), 'notanemail');
      await tester.enterText(
          find.widgetWithText(TextFormField, 'Password'), 'password');
      await tester.tap(find.text('Sign In'));
      await tester.pumpAndSettle();

      expect(find.text('Enter a valid email'), findsOneWidget);
    });
  });

  // ── AgentsScreen interactions ─────────────────────────────────────────────

  group('AgentsScreen interactions', () {
    testWidgets('opens hire agent dialog when button is tapped',
        (tester) async {
      final mockClient = MockHttpClient();
      when(() => mockClient.get(any(), headers: any(named: 'headers')))
          .thenAnswer(
              (_) async => http.Response(jsonEncode(<dynamic>[]), 200));

      await tester.pumpWidget(_wrap(
        const AgentsScreen(),
        overrides: [apiServiceProvider.overrideWithValue(_mockApi(mockClient))],
      ));
      await tester.pumpAndSettle();

      await tester.tap(find.text('Hire Agent'));
      await tester.pumpAndSettle();

      expect(find.text('Hire'), findsWidgets);
      expect(find.byType(AlertDialog), findsOneWidget);
    });

    testWidgets('cancels hire dialog', (tester) async {
      final mockClient = MockHttpClient();
      when(() => mockClient.get(any(), headers: any(named: 'headers')))
          .thenAnswer(
              (_) async => http.Response(jsonEncode(<dynamic>[]), 200));

      await tester.pumpWidget(_wrap(
        const AgentsScreen(),
        overrides: [apiServiceProvider.overrideWithValue(_mockApi(mockClient))],
      ));
      await tester.pumpAndSettle();

      await tester.tap(find.text('Hire Agent'));
      await tester.pumpAndSettle();

      await tester.tap(find.text('Cancel'));
      await tester.pumpAndSettle();

      expect(find.byType(AlertDialog), findsNothing);
    });

    testWidgets('hires an agent via dialog', (tester) async {
      final mockClient = MockHttpClient();
      // First call for list, second for hire
      var callCount = 0;
      when(() => mockClient.get(any(), headers: any(named: 'headers')))
          .thenAnswer((_) async => http.Response(jsonEncode(<dynamic>[]), 200));
      when(() => mockClient.post(any(),
              headers: any(named: 'headers'), body: any(named: 'body')))
          .thenAnswer((_) async => http.Response(
              jsonEncode({
                'id': 'a-new',
                'name': 'NewBot',
                'role': 'engineer',
                'status': 'pending',
                'organization_id': 'org-1',
                'created_at': '2025-01-01T00:00:00Z',
              }),
              200));

      await tester.pumpWidget(_wrap(
        const AgentsScreen(),
        overrides: [apiServiceProvider.overrideWithValue(_mockApi(mockClient))],
      ));
      await tester.pumpAndSettle();

      await tester.tap(find.text('Hire Agent'));
      await tester.pumpAndSettle();

      await tester.enterText(
          find.widgetWithText(TextField, 'Agent Name'), 'NewBot');
      await tester.enterText(
          find.widgetWithText(TextField, 'Role'), 'engineer');

      await tester.tap(find.text('Hire').last);
      await tester.pump();
      await tester.pump(const Duration(seconds: 1));
      await tester.pumpAndSettle();
    });

    testWidgets('shows fire confirmation and cancels', (tester) async {
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

      await tester.pumpWidget(_wrap(
        const AgentsScreen(),
        overrides: [apiServiceProvider.overrideWithValue(_mockApi(mockClient))],
      ));
      await tester.pumpAndSettle();

      // Find the delete icon for the agent
      expect(find.text('Alice'), findsOneWidget);
    });
  });

  // ── MeetingsScreen interactions ───────────────────────────────────────────

  group('MeetingsScreen interactions', () {
    testWidgets('shows empty state and opens create dialog', (tester) async {
      final mockClient = MockHttpClient();
      when(() => mockClient.get(any(), headers: any(named: 'headers')))
          .thenAnswer(
              (_) async => http.Response(jsonEncode(<dynamic>[]), 200));

      await tester.pumpWidget(_wrap(
        const MeetingsScreen(),
        overrides: [apiServiceProvider.overrideWithValue(_mockApi(mockClient))],
      ));
      await tester.pumpAndSettle();

      expect(find.text('No active meeting rooms.'), findsOneWidget);
      await tester.tap(find.text('New Room'));
      await tester.pumpAndSettle();

      expect(find.byType(AlertDialog), findsOneWidget);
    });

    testWidgets('can cancel create room dialog', (tester) async {
      final mockClient = MockHttpClient();
      when(() => mockClient.get(any(), headers: any(named: 'headers')))
          .thenAnswer(
              (_) async => http.Response(jsonEncode(<dynamic>[]), 200));

      await tester.pumpWidget(_wrap(
        const MeetingsScreen(),
        overrides: [apiServiceProvider.overrideWithValue(_mockApi(mockClient))],
      ));
      await tester.pumpAndSettle();

      await tester.tap(find.text('New Room'));
      await tester.pumpAndSettle();

      await tester.tap(find.text('Cancel'));
      await tester.pumpAndSettle();

      expect(find.byType(AlertDialog), findsNothing);
    });

    testWidgets('creates a room via dialog', (tester) async {
      final mockClient = MockHttpClient();
      when(() => mockClient.get(any(), headers: any(named: 'headers')))
          .thenAnswer(
              (_) async => http.Response(jsonEncode(<dynamic>[]), 200));
      when(() => mockClient.post(any(),
              headers: any(named: 'headers'), body: any(named: 'body')))
          .thenAnswer((_) async =>
              http.Response(jsonEncode({'id': 'm1', 'name': 'Standup'}), 200));

      await tester.pumpWidget(_wrap(
        const MeetingsScreen(),
        overrides: [apiServiceProvider.overrideWithValue(_mockApi(mockClient))],
      ));
      await tester.pumpAndSettle();

      await tester.tap(find.text('New Room'));
      await tester.pumpAndSettle();

      await tester.enterText(
          find.widgetWithText(TextField, 'Room Name'), 'Standup');
      await tester.tap(find.text('Create'));
      await tester.pump(const Duration(seconds: 1));
      await tester.pumpAndSettle();
    });

    testWidgets('shows meeting list with join button', (tester) async {
      final mockClient = MockHttpClient();
      when(() => mockClient.get(any(), headers: any(named: 'headers')))
          .thenAnswer((_) async => http.Response(
              jsonEncode([
                {'id': 'm1', 'name': 'Standup', 'participants': []}
              ]),
              200));

      await tester.pumpWidget(_wrap(
        const MeetingsScreen(),
        overrides: [apiServiceProvider.overrideWithValue(_mockApi(mockClient))],
      ));
      await tester.pumpAndSettle();

      expect(find.text('Standup'), findsOneWidget);
      expect(find.text('Join'), findsOneWidget);
    });
  });

  // ── AiConfigScreen interactions ───────────────────────────────────────────

  group('AiConfigScreen interactions', () {
    testWidgets('shows empty state and opens add dialog', (tester) async {
      final mockClient = MockHttpClient();
      when(() => mockClient.get(any(), headers: any(named: 'headers')))
          .thenAnswer(
              (_) async => http.Response(jsonEncode(<dynamic>[]), 200));

      await tester.pumpWidget(_wrap(
        const AiConfigScreen(),
        overrides: [apiServiceProvider.overrideWithValue(_mockApi(mockClient))],
      ));
      await tester.pumpAndSettle();

      expect(find.text('No AI providers configured'), findsOneWidget);

      await tester.tap(find.text('Add Provider').first);
      await tester.pumpAndSettle();

      expect(find.byType(AlertDialog), findsOneWidget);
    });

    testWidgets('can cancel add provider dialog', (tester) async {
      final mockClient = MockHttpClient();
      when(() => mockClient.get(any(), headers: any(named: 'headers')))
          .thenAnswer(
              (_) async => http.Response(jsonEncode(<dynamic>[]), 200));

      await tester.pumpWidget(_wrap(
        const AiConfigScreen(),
        overrides: [apiServiceProvider.overrideWithValue(_mockApi(mockClient))],
      ));
      await tester.pumpAndSettle();

      await tester.tap(find.text('Add Provider').first);
      await tester.pumpAndSettle();

      await tester.tap(find.text('Cancel'));
      await tester.pumpAndSettle();

      expect(find.byType(AlertDialog), findsNothing);
    });

    testWidgets('shows providers list with edit key button', (tester) async {
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

      await tester.pumpWidget(_wrap(
        const AiConfigScreen(),
        overrides: [apiServiceProvider.overrideWithValue(_mockApi(mockClient))],
      ));
      await tester.pumpAndSettle();

      expect(find.text('OpenAI'), findsOneWidget);
    });
  });

  // ── ChannelsScreen interactions ───────────────────────────────────────────

  group('ChannelsScreen interactions', () {
    testWidgets('shows empty state and opens add channel dialog',
        (tester) async {
      final mockClient = MockHttpClient();
      when(() => mockClient.get(any(), headers: any(named: 'headers')))
          .thenAnswer(
              (_) async => http.Response(jsonEncode(<dynamic>[]), 200));

      await tester.pumpWidget(_wrap(
        const ChannelsScreen(),
        overrides: [apiServiceProvider.overrideWithValue(_mockApi(mockClient))],
      ));
      await tester.pumpAndSettle();

      expect(find.text('No channels yet'), findsOneWidget);

      await tester.tap(find.text('Add Channel').first);
      await tester.pumpAndSettle();

      expect(find.byType(AlertDialog), findsOneWidget);
    });

    testWidgets('can cancel add channel dialog', (tester) async {
      final mockClient = MockHttpClient();
      when(() => mockClient.get(any(), headers: any(named: 'headers')))
          .thenAnswer(
              (_) async => http.Response(jsonEncode(<dynamic>[]), 200));

      await tester.pumpWidget(_wrap(
        const ChannelsScreen(),
        overrides: [apiServiceProvider.overrideWithValue(_mockApi(mockClient))],
      ));
      await tester.pumpAndSettle();

      await tester.tap(find.text('Add Channel').first);
      await tester.pumpAndSettle();

      await tester.tap(find.text('Cancel'));
      await tester.pumpAndSettle();

      expect(find.byType(AlertDialog), findsNothing);
    });
  });

  // ── SkillsScreen interactions ─────────────────────────────────────────────

  group('SkillsScreen interactions', () {
    testWidgets('shows empty state when no skills', (tester) async {
      final mockClient = MockHttpClient();
      when(() => mockClient.get(any(), headers: any(named: 'headers')))
          .thenAnswer(
              (_) async => http.Response(jsonEncode(<dynamic>[]), 200));

      await tester.pumpWidget(_wrap(
        const SkillsScreen(),
        overrides: [apiServiceProvider.overrideWithValue(_mockApi(mockClient))],
      ));
      await tester.pumpAndSettle();

      expect(find.text('No skills in this category.'), findsOneWidget);
    });

    testWidgets('shows installed skill with toggle', (tester) async {
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

      await tester.pumpWidget(_wrap(
        const SkillsScreen(),
        overrides: [apiServiceProvider.overrideWithValue(_mockApi(mockClient))],
      ));
      await tester.pumpAndSettle();

      expect(find.text('web_search'), findsOneWidget);
    });
  });

  // ── ChatScreen interactions ───────────────────────────────────────────────

  group('ChatScreen interactions', () {
    testWidgets('renders chat screen with null centrifuge service',
        (tester) async {
      await tester.pumpWidget(_wrap(
        const ChatScreen(),
        overrides: [
          apiServiceProvider.overrideWithValue(null),
          centrifugeServiceProvider.overrideWithValue(null),
          authStateProvider.overrideWith(() => _FakeAuthNotifier(null)),
        ],
      ));
      await tester.pump();
      expect(find.byType(Scaffold), findsOneWidget);
    });

    testWidgets('shows empty messages area', (tester) async {
      await tester.pumpWidget(_wrap(
        const ChatScreen(),
        overrides: [
          centrifugeServiceProvider.overrideWithValue(null),
          authStateProvider.overrideWith(() => _FakeAuthNotifier(
                const AuthUser(
                  id: 'u1',
                  email: 'a@b.com',
                  name: 'Alice',
                  role: 'admin',
                  organizationId: 'org-1',
                  token: 'tok',
                ),
              )),
        ],
      ));
      await tester.pump();
      expect(find.byType(Scaffold), findsOneWidget);
    });
  });
}

// ── Fake AuthNotifier ─────────────────────────────────────────────────────

class _FakeAuthNotifier extends AuthNotifier {
  final AuthUser? _user;
  _FakeAuthNotifier(this._user);

  @override
  Future<AuthUser?> build() async => _user;
}
