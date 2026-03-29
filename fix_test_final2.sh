cat << 'PYEOF' > fix.py
import re

with open('srcs/app/lib/screens/widget_test.dart', 'r') as f:
    content = f.read()

# Force replace both tests completely in widget_test.dart

old_block = """    testWidgets('renders without error when no user logged in', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            clientSettingsProvider.overrideWith((ref) => _FakeSettingsNotifierWidgetTest(ref)..state = const AsyncData(ClientSettings(backendUrl: 'test', standaloneMode: false))),
          ],
          child: const MaterialApp(home: SettingsScreen()),
        ),
      );
      await tester.pump();
      expect(find.text('Sign Out'), findsOneWidget);
    });

    testWidgets('renders user info when user is logged in', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            clientSettingsProvider.overrideWith((ref) => _FakeSettingsNotifierWidgetTest(ref)..state = const AsyncData(ClientSettings(backendUrl: 'test', standaloneMode: false))),
            authStateProvider.overrideWith(() => _FakeAuthNotifier(const AuthUser(
                  id: '1',
                  email: 'test@test.com',
                  name: 'Test User',
                  role: 'user',
                  organizationId: 'org1',
                  token: 'tok',
                ))),
          ],
          child: const MaterialApp(home: SettingsScreen()),
        ),
      );
      await tester.pump();
      expect(find.text('Test User'), findsOneWidget);
      expect(find.text('test@test.com'), findsOneWidget);
    });"""

new_block = """    testWidgets('renders without error when no user logged in', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            clientSettingsProvider.overrideWith((ref) => _FakeSettingsNotifierWidgetTest(ref)..state = const AsyncData(ClientSettings(backendUrl: 'test', standaloneMode: false))),
            localManagerServiceProvider.overrideWithValue(FakeLocalManagerService()),
          ],
          child: const MaterialApp(home: SettingsScreen()),
        ),
      );
      await tester.pumpAndSettle();
      expect(find.text('Sign Out'), findsOneWidget);
    });

    testWidgets('renders user info when user is logged in', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            clientSettingsProvider.overrideWith((ref) => _FakeSettingsNotifierWidgetTest(ref)..state = const AsyncData(ClientSettings(backendUrl: 'test', standaloneMode: false))),
            localManagerServiceProvider.overrideWithValue(FakeLocalManagerService()),
            authStateProvider.overrideWith(() => _FakeAuthNotifier(const AuthUser(
                  id: '1',
                  email: 'test@test.com',
                  name: 'Test User',
                  role: 'user',
                  organizationId: 'org1',
                  token: 'tok',
                ))),
          ],
          child: const MaterialApp(home: SettingsScreen()),
        ),
      );
      await tester.pumpAndSettle();
      expect(find.text('Test User'), findsOneWidget);
      expect(find.text('test@test.com'), findsOneWidget);
    });"""

content = content.replace(old_block, new_block)

fake_local_manager = """
class FakeLocalManagerService extends LocalManagerService {
  @override
  Future<void> startService() async {}
  @override
  Future<void> stopService() async {}
  @override
  Future<bool> isServiceRunning() async => true;
  @override
  Future<String> runDoctor() async => 'OK';
}
"""

if "class FakeLocalManagerService" not in content:
    content = content.replace("class _FakeAuthNotifier extends AuthNotifier {", fake_local_manager + "\nclass _FakeAuthNotifier extends AuthNotifier {")
    content = content.replace("import 'package:flutter_test/flutter_test.dart';", "import 'package:flutter_test/flutter_test.dart';\nimport 'package:ohc_app/services/local_manager_service.dart';")


with open('srcs/app/lib/screens/widget_test.dart', 'w') as f:
    f.write(content)

with open('srcs/app/test/desktop_e2e_test.dart', 'r') as f:
    content2 = f.read()

old_block2 = """    testWidgets('Sign Out list tile is rendered and tappable', (tester) async {
      await tester.pumpWidget(_wrap(
        const SettingsScreen(),
        overrides: [
          authStateProvider.overrideWith(() => _FakeAuthNotifier(_fakeUser)),
          clientSettingsProvider.overrideWith((ref) => _FakeSettingsNotifier(ref)..state = const AsyncData(ClientSettings(backendUrl: 'test', standaloneMode: false))),
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
          clientSettingsProvider.overrideWith((ref) => _FakeSettingsNotifier(ref)..state = const AsyncData(ClientSettings(backendUrl: 'test', standaloneMode: false))),
        ],
      ));
      await tester.pump();

      expect(find.text('Dev User'), findsOneWidget);
      expect(find.text('dev@example.com'), findsOneWidget);
    });"""

new_block2 = """    testWidgets('Sign Out list tile is rendered and tappable', (tester) async {
      await tester.pumpWidget(_wrap(
        const SettingsScreen(),
        overrides: [
          authStateProvider.overrideWith(() => _FakeAuthNotifier(_fakeUser)),
          clientSettingsProvider.overrideWith((ref) => _FakeSettingsNotifier(ref)..state = const AsyncData(ClientSettings(backendUrl: 'test', standaloneMode: false))),
          localManagerServiceProvider.overrideWithValue(FakeLocalManagerService()),
        ],
      ));
      await tester.pumpAndSettle();

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
          clientSettingsProvider.overrideWith((ref) => _FakeSettingsNotifier(ref)..state = const AsyncData(ClientSettings(backendUrl: 'test', standaloneMode: false))),
          localManagerServiceProvider.overrideWithValue(FakeLocalManagerService()),
        ],
      ));
      await tester.pumpAndSettle();

      expect(find.text('Dev User'), findsOneWidget);
      expect(find.text('dev@example.com'), findsOneWidget);
    });"""

content2 = content2.replace(old_block2, new_block2)

if "class FakeLocalManagerService" not in content2:
    content2 = content2.replace("class _FakeAuthNotifier extends AuthNotifier {", fake_local_manager + "\nclass _FakeAuthNotifier extends AuthNotifier {")
    content2 = content2.replace("import 'package:flutter_test/flutter_test.dart';", "import 'package:flutter_test/flutter_test.dart';\nimport 'package:ohc_app/services/local_manager_service.dart';")

with open('srcs/app/test/desktop_e2e_test.dart', 'w') as f:
    f.write(content2)


PYEOF
python3 fix.py
