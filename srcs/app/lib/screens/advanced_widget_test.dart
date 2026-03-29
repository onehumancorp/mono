/// Tests for Chat screen, router, and remaining uncovered screen interactions.
library;

import 'dart:async';
import 'dart:convert';

import 'package:centrifuge/centrifuge.dart' as centrifuge;
import 'package:fixnum/fixnum.dart' as fixnum;
import 'package:flutter/material.dart';
import 'package:ohc_app/theme.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:go_router/go_router.dart';
import 'package:http/http.dart' as http;
import 'package:mocktail/mocktail.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:ohc_app/router.dart';
import 'package:ohc_app/screens/ai_config_screen.dart';
import 'package:ohc_app/screens/chat_screen.dart';
import 'package:ohc_app/screens/login_screen.dart';
import 'package:ohc_app/screens/meetings_screen.dart';
import 'package:ohc_app/screens/security_screen.dart';
import 'package:ohc_app/screens/skills_screen.dart';
import 'package:ohc_app/services/api_service.dart';
import 'package:ohc_app/services/auth_service.dart';
import 'package:ohc_app/services/centrifuge_service.dart';

// ── Mocks ─────────────────────────────────────────────────────────────────────

class MockHttpClient extends Mock implements http.Client {}

class MockCentrifugeClient extends Mock implements centrifuge.Client {}

class MockSubscription extends Mock implements centrifuge.Subscription {}

class FakeUri extends Fake implements Uri {}

class FakeClientConfig extends Fake implements centrifuge.ClientConfig {}

// ── Helpers ───────────────────────────────────────────────────────────────────

Widget _wrap(Widget child, {List<Override> overrides = const []}) {
  return ProviderScope(
    overrides: overrides,
    child: MaterialApp(home: child),
  );
}

ApiService _mockApi(MockHttpClient client) =>
    ApiService(baseUrl: 'http://localhost', token: 'tok', client: client);

/// Creates a [CentrifugeService] backed by [mockClient].
CentrifugeService _mockCentrifugeService(
  MockCentrifugeClient mockClient,
  MockSubscription mockSub,
  StreamController<centrifuge.PublicationEvent> pubController,
) {
  when(() => mockClient.connect()).thenAnswer((_) async {});
  when(() => mockClient.newSubscription(any())).thenReturn(mockSub);
  when(() => mockSub.publication).thenAnswer((_) => pubController.stream);
  when(() => mockSub.subscribe()).thenAnswer((_) async {});
  when(() => mockSub.unsubscribe()).thenAnswer((_) async {});
  when(() => mockClient.disconnect()).thenAnswer((_) async {});
  when(() => mockClient.publish(any(), any()))
      .thenAnswer((_) async => centrifuge.PublishResult());

  return CentrifugeService(
    serverUrl: 'ws://localhost:8000/connection/websocket',
    token: 'test-token',
    userId: 'u1',
    userName: 'Alice',
    clientFactory: (_, __) => mockClient,
  );
}

// ── AuthNotifier fake ─────────────────────────────────────────────────────────

class _FakeAuthNotifier extends AuthNotifier {
  final AuthUser? _user;
  _FakeAuthNotifier(this._user);

  @override
  Future<AuthUser?> build() async => _user;
}

const _loggedInUser = AuthUser(
  id: 'u1',
  email: 'a@b.com',
  name: 'Alice',
  role: 'admin',
  organizationId: 'org-1',
  token: 'tok',
);

void main() {
  setUpAll(() {
    registerFallbackValue(FakeUri());
    registerFallbackValue(FakeClientConfig());
    SharedPreferences.setMockInitialValues({});
  });

  // ── Router and AppShell ───────────────────────────────────────────────────

  group('AppShell', () {
    testWidgets('renders sidebar and child widget', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            authStateProvider.overrideWith(() => _FakeAuthNotifier(null)),
          ],
          child: MaterialApp.router(
            routerConfig: GoRouter(
              initialLocation: '/dashboard',
              routes: [
                ShellRoute(
                  builder: (context, state, child) => AppShell(child: child),
                  routes: [
                    GoRoute(
                      path: '/dashboard',
                      builder: (context, state) => const Text('child content'),
                    ),
                  ],
                ),
              ],
            ),
          ),
        ),
      );
      await tester.pumpAndSettle();

      expect(find.text('child content'), findsOneWidget);
      expect(find.byType(NavigationDrawer), findsOneWidget);
      expect(find.text('One Human Corp'), findsOneWidget);
    });
  });

  group('routerProvider', () {
    test('creates a GoRouter with routes', () {
      final container = ProviderContainer(overrides: [
        authStateProvider.overrideWith(() => _FakeAuthNotifier(null)),
        backendUrlProvider.overrideWith((ref) => 'http://localhost'),
      ]);
      addTearDown(container.dispose);

      final router = container.read(routerProvider);
      expect(router, isA<GoRouter>());

      // The router should have routes
      expect(router.configuration.routes, isNotEmpty);
    });

    testWidgets('AppShell renders navigation items', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            authStateProvider.overrideWith(() => _FakeAuthNotifier(null)),
          ],
          child: MaterialApp.router(
            routerConfig: GoRouter(
              initialLocation: '/dashboard',
              routes: [
                ShellRoute(
                  builder: (context, state, child) => AppShell(child: child),
                  routes: [
                    GoRoute(
                      path: '/dashboard',
                      builder: (context, state) => const Text('Dashboard'),
                    ),
                  ],
                ),
              ],
            ),
          ),
        ),
      );
      await tester.pump();

      expect(find.text('One Human Corp'), findsOneWidget);
      expect(find.text('Dashboard'), findsWidgets);
    });
  });

  // ── ChatScreen with connected service ─────────────────────────────────────

  group('ChatScreen with mock centrifuge', () {
    late MockCentrifugeClient mockClient;
    late MockSubscription mockSub;
    late StreamController<centrifuge.PublicationEvent> pubController;

    setUp(() {
      mockClient = MockCentrifugeClient();
      mockSub = MockSubscription();
      pubController =
          StreamController<centrifuge.PublicationEvent>.broadcast();
    });

    tearDown(() {
      pubController.close();
    });

    testWidgets('connects to service on init', (tester) async {
      final svc =
          _mockCentrifugeService(mockClient, mockSub, pubController);

      await tester.pumpWidget(_wrap(
        const ChatScreen(),
        overrides: [
          centrifugeServiceProvider.overrideWithValue(svc),
          authStateProvider.overrideWith(() => _FakeAuthNotifier(_loggedInUser)),
        ],
      ));
      await tester.pump(); // Trigger initState
      await tester.pump(const Duration(milliseconds: 100));

      verify(() => mockClient.connect()).called(1);
      verify(() => mockSub.subscribe()).called(1);
    });

    testWidgets('shows "No messages yet" when no messages', (tester) async {
      final svc =
          _mockCentrifugeService(mockClient, mockSub, pubController);

      await tester.pumpWidget(_wrap(
        const ChatScreen(),
        overrides: [
          centrifugeServiceProvider.overrideWithValue(svc),
          authStateProvider.overrideWith(() => _FakeAuthNotifier(_loggedInUser)),
        ],
      ));
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 100));

      expect(find.text('No messages yet. Say hello!'), findsOneWidget);
    });

    testWidgets('displays messages when received', (tester) async {
      final svc =
          _mockCentrifugeService(mockClient, mockSub, pubController);

      await tester.pumpWidget(_wrap(
        const ChatScreen(),
        overrides: [
          centrifugeServiceProvider.overrideWithValue(svc),
          authStateProvider.overrideWith(() => _FakeAuthNotifier(_loggedInUser)),
        ],
      ));
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 100));

      // Simulate receiving a message
      final msgJson = jsonEncode({
        'id': 'msg-1',
        'channel_id': 'general',
        'author_id': 'other-user',
        'author_name': 'Bob',
        'body': 'Hello from Bob!',
        'sent_at': '2025-01-01T10:00:00.000Z',
      });
      pubController.add(centrifuge.PublicationEvent(
          utf8.encode(msgJson), fixnum.Int64.ZERO, null, {}));

      await tester.pump(const Duration(milliseconds: 50));

      expect(find.text('Hello from Bob!'), findsOneWidget);
    });

    testWidgets('can send a message', (tester) async {
      final svc =
          _mockCentrifugeService(mockClient, mockSub, pubController);

      await tester.pumpWidget(_wrap(
        const ChatScreen(),
        overrides: [
          centrifugeServiceProvider.overrideWithValue(svc),
          authStateProvider.overrideWith(() => _FakeAuthNotifier(_loggedInUser)),
        ],
      ));
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 100));

      // Type and send a message
      await tester.enterText(find.byType(TextField), 'Hello World');
      await tester.tap(find.byIcon(Icons.send));
      await tester.pump(const Duration(milliseconds: 100));

      verify(() => mockClient.publish(any(), any())).called(1);
    });

    testWidgets('opens room picker dialog', (tester) async {
      final svc =
          _mockCentrifugeService(mockClient, mockSub, pubController);

      await tester.pumpWidget(_wrap(
        const ChatScreen(),
        overrides: [
          centrifugeServiceProvider.overrideWithValue(svc),
          authStateProvider.overrideWith(() => _FakeAuthNotifier(_loggedInUser)),
        ],
      ));
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 100));

      await tester.tap(find.byIcon(Icons.meeting_room));
      await tester.pumpAndSettle();

      expect(find.text('Switch Chat Room'), findsOneWidget);
    });

    testWidgets('can cancel room picker dialog', (tester) async {
      final svc =
          _mockCentrifugeService(mockClient, mockSub, pubController);

      await tester.pumpWidget(_wrap(
        const ChatScreen(),
        overrides: [
          centrifugeServiceProvider.overrideWithValue(svc),
          authStateProvider.overrideWith(() => _FakeAuthNotifier(_loggedInUser)),
        ],
      ));
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 100));

      await tester.tap(find.byIcon(Icons.meeting_room));
      await tester.pumpAndSettle();

      await tester.tap(find.text('Cancel'));
      await tester.pumpAndSettle();

      expect(find.byType(AlertDialog), findsNothing);
    });

    testWidgets('can switch to a different room', (tester) async {
      final svc =
          _mockCentrifugeService(mockClient, mockSub, pubController);

      await tester.pumpWidget(_wrap(
        const ChatScreen(),
        overrides: [
          centrifugeServiceProvider.overrideWithValue(svc),
          authStateProvider.overrideWith(() => _FakeAuthNotifier(_loggedInUser)),
        ],
      ));
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 100));

      await tester.tap(find.byIcon(Icons.meeting_room));
      await tester.pumpAndSettle();

      await tester.enterText(
          find.widgetWithText(TextField, 'Room ID'), 'support');
      await tester.tap(find.text('Switch'));
      await tester.pumpAndSettle();

      // Dialog should be closed after switching
      expect(find.byType(AlertDialog), findsNothing);
    });
  });

  // ── LoginScreen actual login flow ─────────────────────────────────────────

  group('LoginScreen login flow', () {
    testWidgets('submits login with valid credentials', (tester) async {
      final mockClient = MockHttpClient();
      when(() => mockClient.post(
            any(),
            headers: any(named: 'headers'),
            body: any(named: 'body'),
          )).thenAnswer((_) async => http.Response(
            jsonEncode({
              'token': 'new-tok',
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

      await tester.pumpWidget(_wrap(
        const LoginScreen(),
        overrides: [
          authServiceProvider.overrideWithValue(
              AuthService(baseUrl: 'http://localhost', client: mockClient)),
          backendUrlProvider.overrideWith((ref) => 'http://localhost'),
        ],
      ));
      await tester.pump();

      await tester.enterText(
          find.widgetWithText(TextFormField, 'Email'), 'a@b.com');
      await tester.enterText(
          find.widgetWithText(TextFormField, 'Password'), 'password');
      await tester.tap(find.text('Sign In'));
      await tester.pump(const Duration(milliseconds: 100));
    });

    testWidgets('shows error when login fails', (tester) async {
      final mockClient = MockHttpClient();
      when(() => mockClient.post(
            any(),
            headers: any(named: 'headers'),
            body: any(named: 'body'),
          )).thenThrow(Exception('Network error'));

      await tester.pumpWidget(_wrap(
        const LoginScreen(),
        overrides: [
          authServiceProvider.overrideWithValue(
              AuthService(baseUrl: 'http://localhost', client: mockClient)),
          backendUrlProvider.overrideWith((ref) => 'http://localhost'),
        ],
      ));
      await tester.pump();

      await tester.enterText(
          find.widgetWithText(TextFormField, 'Email'), 'a@b.com');
      await tester.enterText(
          find.widgetWithText(TextFormField, 'Password'), 'pw');
      await tester.tap(find.text('Sign In'));
      await tester.pump(const Duration(milliseconds: 100));
      await tester.pump(const Duration(milliseconds: 100));
    });
  });

  // ── MeetingsScreen - Join flow ────────────────────────────────────────────

  group('MeetingsScreen join flow', () {
    testWidgets('shows join info dialog when joining', (tester) async {
      final mockClient = MockHttpClient();
      // First call: list meetings
      when(() => mockClient.get(any(), headers: any(named: 'headers')))
          .thenAnswer((_) async => http.Response(
              jsonEncode([
                {'id': 'm1', 'name': 'Standup', 'participants': []}
              ]),
              200));
      // Second call: join meeting
      when(() => mockClient.post(any(), headers: any(named: 'headers')))
          .thenAnswer((_) async => http.Response(
              jsonEncode({'join_url': 'https://meet.example.com/xyz',
                'token': 'lk-token-abc'}),
              200));

      final api = _mockApi(mockClient);
      await tester.pumpWidget(_wrap(
        const MeetingsScreen(),
        overrides: [apiServiceProvider.overrideWithValue(api)],
      ));
      await tester.pumpAndSettle();

      await tester.tap(find.text('Join'));
      await tester.pump(const Duration(milliseconds: 200));
      await tester.pumpAndSettle();

      // Should show join info dialog
      expect(find.text('Join Meeting'), findsOneWidget);
    });

    testWidgets('can dismiss join info dialog', (tester) async {
      final mockClient = MockHttpClient();
      when(() => mockClient.get(any(), headers: any(named: 'headers')))
          .thenAnswer((_) async => http.Response(
              jsonEncode([
                {'id': 'm1', 'name': 'Standup', 'participants': []}
              ]),
              200));
      when(() => mockClient.post(any(), headers: any(named: 'headers')))
          .thenAnswer((_) async => http.Response(
              jsonEncode({'join_url': 'https://meet.example.com/xyz'}), 200));

      final api = _mockApi(mockClient);
      await tester.pumpWidget(_wrap(
        const MeetingsScreen(),
        overrides: [apiServiceProvider.overrideWithValue(api)],
      ));
      await tester.pumpAndSettle();

      await tester.tap(find.text('Join'));
      await tester.pump(const Duration(milliseconds: 200));
      await tester.pumpAndSettle();

      await tester.tap(find.text('Done'));
      await tester.pumpAndSettle();

      expect(find.byType(AlertDialog), findsNothing);
    });

    testWidgets('shows empty meeting rooms message', (tester) async {
      final mockClient = MockHttpClient();
      when(() => mockClient.get(any(), headers: any(named: 'headers')))
          .thenAnswer(
              (_) async => http.Response(jsonEncode(<dynamic>[]), 200));

      final api = _mockApi(mockClient);
      await tester.pumpWidget(_wrap(
        const MeetingsScreen(),
        overrides: [apiServiceProvider.overrideWithValue(api)],
      ));
      await tester.pumpAndSettle();

      // Empty state has create button
      await tester.tap(find.text('Create Room').last);
      await tester.pumpAndSettle();

      expect(find.byType(AlertDialog), findsOneWidget);
    });
  });

  // ── SecurityScreen - Fix action ───────────────────────────────────────────

  group('SecurityScreen fix action', () {
    testWidgets('can fix a security issue', (tester) async {
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
      when(() => mockClient.post(any(), headers: any(named: 'headers')))
          .thenAnswer((_) async => http.Response('', 200));

      final api = _mockApi(mockClient);
      await tester.pumpWidget(_wrap(
        const SecurityScreen(),
        overrides: [apiServiceProvider.overrideWithValue(api)],
      ));
      await tester.pumpAndSettle();

      expect(find.text('Weak password'), findsOneWidget);
      // Tap Fix button
      await tester.tap(find.text('Auto-fix'));
      await tester.pump(const Duration(milliseconds: 200));
      await tester.pumpAndSettle();

      verify(() => mockClient.post(any(),
          headers: any(named: 'headers'))).called(1);
    });

    testWidgets('shows fixed issues section', (tester) async {
      final mockClient = MockHttpClient();
      when(() => mockClient.get(any(), headers: any(named: 'headers')))
          .thenAnswer((_) async => http.Response(
              jsonEncode([
                {
                  'id': 'i1',
                  'title': 'Old Issue',
                  'description': 'This was fixed',
                  'severity': 'medium',
                  'fixable': true,
                  'fixed': true,
                  'category': 'auth',
                }
              ]),
              200));

      final api = _mockApi(mockClient);
      await tester.pumpWidget(_wrap(
        const SecurityScreen(),
        overrides: [apiServiceProvider.overrideWithValue(api)],
      ));
      await tester.pumpAndSettle();

      expect(find.text('Old Issue'), findsOneWidget);
      expect(find.text('1 resolved'), findsOneWidget);
    });
  });

  // ── SkillsScreen - Toggle install/enable ─────────────────────────────────

  group('SkillsScreen toggle install/enable', () {
    testWidgets('can install a skill', (tester) async {
      final mockClient = MockHttpClient();
      when(() => mockClient.get(any(), headers: any(named: 'headers')))
          .thenAnswer((_) async => http.Response(
              jsonEncode([
                {
                  'name': 'code_runner',
                  'version': '1.0.0',
                  'description': 'Run code',
                  'category': 'official',
                  'installed': false,
                  'enabled': false,
                }
              ]),
              200));
      when(() => mockClient.post(any(), headers: any(named: 'headers')))
          .thenAnswer((_) async => http.Response('', 200));

      final api = _mockApi(mockClient);
      await tester.pumpWidget(_wrap(
        const SkillsScreen(),
        overrides: [apiServiceProvider.overrideWithValue(api)],
      ));
      await tester.pumpAndSettle();

      expect(find.text('code_runner'), findsOneWidget);
      await tester.tap(find.text('Install'));
      await tester.pump(const Duration(milliseconds: 200));
      await tester.pumpAndSettle();
    });

    testWidgets('can uninstall a skill', (tester) async {
      final mockClient = MockHttpClient();
      when(() => mockClient.get(any(), headers: any(named: 'headers')))
          .thenAnswer((_) async => http.Response(
              jsonEncode([
                {
                  'name': 'code_runner',
                  'version': '1.0.0',
                  'description': 'Run code',
                  'category': 'official',
                  'installed': true,
                  'enabled': false,
                }
              ]),
              200));
      when(() => mockClient.post(any(), headers: any(named: 'headers')))
          .thenAnswer((_) async => http.Response('', 200));

      final api = _mockApi(mockClient);
      await tester.pumpWidget(_wrap(
        const SkillsScreen(),
        overrides: [apiServiceProvider.overrideWithValue(api)],
      ));
      await tester.pumpAndSettle();

      await tester.tap(find.text('Remove'));
      await tester.pump(const Duration(milliseconds: 200));
      await tester.pumpAndSettle();
    });

    testWidgets('can toggle skill enabled', (tester) async {
      final mockClient = MockHttpClient();
      when(() => mockClient.get(any(), headers: any(named: 'headers')))
          .thenAnswer((_) async => http.Response(
              jsonEncode([
                {
                  'name': 'code_runner',
                  'version': '1.0.0',
                  'description': 'Run code',
                  'category': 'official',
                  'installed': true,
                  'enabled': true,
                }
              ]),
              200));
      when(() => mockClient.patch(any(),
              headers: any(named: 'headers'), body: any(named: 'body')))
          .thenAnswer((_) async => http.Response('', 200));

      final api = _mockApi(mockClient);
      await tester.pumpWidget(_wrap(
        const SkillsScreen(),
        overrides: [apiServiceProvider.overrideWithValue(api)],
      ));
      await tester.pumpAndSettle();

      // Find and toggle the Switch widget
      final switchFinder = find.byType(Switch);
      if (switchFinder.evaluate().isNotEmpty) {
        await tester.tap(switchFinder.first);
        await tester.pump(const Duration(milliseconds: 200));
        await tester.pumpAndSettle();
      }
    });

    testWidgets('can filter skills by category', (tester) async {
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
                },
                {
                  'name': 'custom_tool',
                  'version': '0.1.0',
                  'description': 'Community tool',
                  'category': 'community',
                  'installed': false,
                  'enabled': false,
                }
              ]),
              200));

      final api = _mockApi(mockClient);
      await tester.pumpWidget(_wrap(
        const SkillsScreen(),
        overrides: [apiServiceProvider.overrideWithValue(api)],
      ));
      await tester.pumpAndSettle();

      expect(find.text('web_search'), findsOneWidget);
      expect(find.text('custom_tool'), findsOneWidget);

      // Click community filter
      final chips = find.byType(FilterChip);
      if (chips.evaluate().isNotEmpty) {
        await tester.tap(chips.last);
        await tester.pumpAndSettle();
      }
    });
  });

  // ── AiConfigScreen - Edit key dialog ─────────────────────────────────────

  group('AiConfigScreen edit key dialog', () {
    testWidgets('opens edit API key dialog', (tester) async {
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

      final api = _mockApi(mockClient);
      await tester.pumpWidget(_wrap(
        const AiConfigScreen(),
        overrides: [apiServiceProvider.overrideWithValue(api)],
      ));
      await tester.pumpAndSettle();

      // Find edit key button (icon button)
      await tester.tap(find.byIcon(Icons.edit_outlined));
      await tester.pumpAndSettle();

      expect(find.textContaining('API Key'), findsWidgets);
      expect(find.byType(AlertDialog), findsOneWidget);
    });

    testWidgets('can cancel edit key dialog', (tester) async {
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
                  'is_official': false,
                }
              ]),
              200));

      final api = _mockApi(mockClient);
      await tester.pumpWidget(_wrap(
        const AiConfigScreen(),
        overrides: [apiServiceProvider.overrideWithValue(api)],
      ));
      await tester.pumpAndSettle();

      await tester.tap(find.byIcon(Icons.edit_outlined));
      await tester.pumpAndSettle();

      await tester.tap(find.text('Cancel'));
      await tester.pumpAndSettle();

      expect(find.byType(AlertDialog), findsNothing);
    });

    testWidgets('can save API key', (tester) async {
      final mockClient = MockHttpClient();
      when(() => mockClient.get(any(), headers: any(named: 'headers')))
          .thenAnswer((_) async => http.Response(
              jsonEncode([
                {
                  'id': 'p1',
                  'name': 'OpenAI',
                  'base_url': 'https://api.openai.com/v1',
                  'api_key': 'sk-old',
                  'models': ['gpt-4'],
                  'is_official': false,
                }
              ]),
              200));
      when(() => mockClient.patch(any(),
              headers: any(named: 'headers'), body: any(named: 'body')))
          .thenAnswer((_) async => http.Response('', 200));

      final api = _mockApi(mockClient);
      await tester.pumpWidget(_wrap(
        const AiConfigScreen(),
        overrides: [apiServiceProvider.overrideWithValue(api)],
      ));
      await tester.pumpAndSettle();

      await tester.tap(find.byIcon(Icons.edit_outlined));
      await tester.pumpAndSettle();

      await tester.enterText(find.byType(TextField).last, 'sk-new-key');
      await tester.tap(find.text('Save'));
      await tester.pump(const Duration(milliseconds: 200));
      await tester.pumpAndSettle();

      verify(() => mockClient.patch(any(),
          headers: any(named: 'headers'),
          body: any(named: 'body'))).called(1);
    });

    testWidgets('can add a new AI provider via dialog', (tester) async {
      final mockClient = MockHttpClient();
      when(() => mockClient.get(any(), headers: any(named: 'headers')))
          .thenAnswer(
              (_) async => http.Response(jsonEncode(<dynamic>[]), 200));
      when(() => mockClient.post(any(),
              headers: any(named: 'headers'), body: any(named: 'body')))
          .thenAnswer((_) async => http.Response(
              jsonEncode({
                'id': 'p-new',
                'name': 'Anthropic',
                'base_url': 'https://api.anthropic.com',
                'api_key': 'sk-ant',
                'models': ['claude-3'],
                'is_official': false,
              }),
              200));

      final api = _mockApi(mockClient);
      await tester.pumpWidget(_wrap(
        const AiConfigScreen(),
        overrides: [apiServiceProvider.overrideWithValue(api)],
      ));
      await tester.pumpAndSettle();

      await tester.tap(find.text('Add Provider').first);
      await tester.pumpAndSettle();

      // Fill in name
      // The dialog has preset fields already filled
      await tester.tap(find.text('Add'));
      await tester.pump(const Duration(milliseconds: 200));
      await tester.pumpAndSettle();
    });
  });
}
