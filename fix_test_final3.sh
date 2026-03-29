cat << 'PYEOF' > fix.py
import re

with open('srcs/app/lib/screens/widget_test.dart', 'r') as f:
    content = f.read()

# Add fake LocalManagerService back properly
fake_local_manager = """
class _FakeLocalManagerService extends LocalManagerService {
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

if "class _FakeLocalManagerService" not in content:
    content = content.replace("class _FakeSettingsNotifierWidgetTest extends ClientSettingsNotifier {", fake_local_manager + "\nclass _FakeSettingsNotifierWidgetTest extends ClientSettingsNotifier {")

content = content.replace("localManagerServiceProvider.overrideWithValue(FakeLocalManagerService()),", "localManagerServiceProvider.overrideWithValue(_FakeLocalManagerService()),")

with open('srcs/app/lib/screens/widget_test.dart', 'w') as f:
    f.write(content)

with open('srcs/app/test/desktop_e2e_test.dart', 'r') as f:
    content2 = f.read()

if "class _FakeLocalManagerService" not in content2:
    content2 = content2.replace("class _FakeSettingsNotifier extends ClientSettingsNotifier {", fake_local_manager + "\nclass _FakeSettingsNotifier extends ClientSettingsNotifier {")

content2 = content2.replace("localManagerServiceProvider.overrideWithValue(FakeLocalManagerService()),", "localManagerServiceProvider.overrideWithValue(_FakeLocalManagerService()),")

with open('srcs/app/test/desktop_e2e_test.dart', 'w') as f:
    f.write(content2)

PYEOF
python3 fix.py
