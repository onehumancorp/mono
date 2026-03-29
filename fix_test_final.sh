cat << 'PYEOF' > fix.py
import re

with open('srcs/app/test/desktop_e2e_test.dart', 'r') as f:
    content = f.read()

# Fix desktop_e2e_test.dart
mock_settings = """
class _FakeSettingsNotifier extends ClientSettingsNotifier {
  _FakeSettingsNotifier(super.ref) : super();
  @override
  Future<void> _load() async {
    state = const AsyncData(ClientSettings(backendUrl: 'http://localhost', standaloneMode: false));
  }
}
"""

content = content.replace("class _FakeAuthNotifier extends AuthNotifier {", mock_settings + "\nclass _FakeAuthNotifier extends AuthNotifier {")

content = content.replace("authStateProvider.overrideWith(() => _FakeAuthNotifier(_fakeUser)),", "authStateProvider.overrideWith(() => _FakeAuthNotifier(_fakeUser)),\n          clientSettingsProvider.overrideWith((ref) => _FakeSettingsNotifier(ref)..state = const AsyncData(ClientSettings(backendUrl: 'test', standaloneMode: false))),")

content = content.replace("import 'package:flutter_test/flutter_test.dart';", "import 'package:flutter_test/flutter_test.dart';\nimport 'package:ohc_app/services/settings_service.dart';")

with open('srcs/app/test/desktop_e2e_test.dart', 'w') as f:
    f.write(content)

with open('srcs/app/lib/screens/widget_test.dart', 'r') as f:
    content2 = f.read()

mock_settings2 = """
class _FakeSettingsNotifierWidgetTest extends ClientSettingsNotifier {
  _FakeSettingsNotifierWidgetTest(super.ref) : super();
  @override
  Future<void> _load() async {
    state = const AsyncData(ClientSettings(backendUrl: 'http://localhost', standaloneMode: false));
  }
}
"""

content2 = content2.replace("class _FakeAuthNotifier extends AuthNotifier {", mock_settings2 + "\nclass _FakeAuthNotifier extends AuthNotifier {")

content2 = content2.replace("authStateProvider.overrideWith(() => _FakeAuthNotifier(const AuthUser(", "clientSettingsProvider.overrideWith((ref) => _FakeSettingsNotifierWidgetTest(ref)..state = const AsyncData(ClientSettings(backendUrl: 'test', standaloneMode: false))),\n            authStateProvider.overrideWith(() => _FakeAuthNotifier(const AuthUser(")

old_test = """    testWidgets('renders without error when no user logged in', (tester) async {
      await tester.pumpWidget(
        const ProviderScope(
          child: MaterialApp(home: SettingsScreen()),
        ),
      );"""

new_test = """    testWidgets('renders without error when no user logged in', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            clientSettingsProvider.overrideWith((ref) => _FakeSettingsNotifierWidgetTest(ref)..state = const AsyncData(ClientSettings(backendUrl: 'test', standaloneMode: false))),
          ],
          child: const MaterialApp(home: SettingsScreen()),
        ),
      );"""

content2 = content2.replace(old_test, new_test)

content2 = content2.replace("import 'package:flutter_test/flutter_test.dart';", "import 'package:flutter_test/flutter_test.dart';\nimport 'package:ohc_app/services/settings_service.dart';")

with open('srcs/app/lib/screens/widget_test.dart', 'w') as f:
    f.write(content2)

PYEOF
python3 fix.py
