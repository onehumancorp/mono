import 'dart:io';

import 'package:flutter_test/flutter_test.dart';
import 'package:ohc_app/services/local_manager_service.dart';

class _CommandTestService extends LocalManagerService {
  _CommandTestService({required this.running}) : super(homeOverride: '/tmp');

  bool running;
  int startCalls = 0;
  int runCalls = 0;
  ProcessResult nextRunResult = ProcessResult(0, 0, '', '');

  @override
  Future<bool> isServiceRunning() async => running;

  @override
  Future<Process> processStart(
    String executable,
    List<String> arguments, {
    bool runInShell = true,
  }) async {
    startCalls++;
    running = true;
    if (Platform.isWindows) {
      return Process.start('cmd', ['/c', 'exit', '0']);
    }
    return Process.start('sh', ['-c', 'true']);
  }

  @override
  Future<ProcessResult> processRun(
    String executable,
    List<String> arguments, {
    bool runInShell = true,
  }) async {
    runCalls++;
    if (arguments.contains('stop')) {
      running = false;
    }
    return nextRunResult;
  }
}

void main() {
  late Directory tempHome;
  late LocalManagerService service;

  setUp(() async {
    tempHome = await Directory.systemTemp.createTemp('ohc_local_manager_test_');
    service = LocalManagerService(homeOverride: tempHome.path);
  });

  tearDown(() async {
    if (await tempHome.exists()) {
      await tempHome.delete(recursive: true);
    }
  });

  test('readConfig returns empty map when config does not exist', () async {
    final cfg = await service.readConfig();
    expect(cfg, isEmpty);
  });

  test('writeConfig persists JSON and readConfig returns it', () async {
    final data = <String, dynamic>{
      'listen_addr': '0.0.0.0:18789',
      'org': 'org-1',
      'features': ['chat', 'skills'],
    };

    await service.writeConfig(data);
    final cfg = await service.readConfig();

    expect(cfg['listen_addr'], '0.0.0.0:18789');
    expect(cfg['org'], 'org-1');
    expect((cfg['features'] as List).length, 2);
  });

  test('saveEnvValue and getEnvValue round trip and update', () async {
    await service.saveEnvValue('API_KEY', 'first');
    expect(await service.getEnvValue('API_KEY'), 'first');

    await service.saveEnvValue('API_KEY', 'updated');
    expect(await service.getEnvValue('API_KEY'), 'updated');

    await service.saveEnvValue('BASE_URL', 'http://localhost:18789');
    expect(await service.getEnvValue('BASE_URL'), 'http://localhost:18789');
  });

  test('getEnvValue returns null when env file missing or key absent', () async {
    expect(await service.getEnvValue('MISSING_KEY'), isNull);

    await service.saveEnvValue('ONLY_KEY', 'value');
    expect(await service.getEnvValue('ANOTHER_KEY'), isNull);
  });

  test('getSystemInfo returns expected fields', () async {
    final info = await service.getSystemInfo();

    expect(info['os'], isA<String>());
    expect(info['os_version'], isA<String>());
    expect(info['dart_version'], isA<String>());
    expect(info['hostname'], isA<String>());
    expect(info['cpus'], isA<int>());
  });

  test('isServiceRunning detects open port 18789', () async {
    ServerSocket? server;
    try {
      server = await ServerSocket.bind(
        InternetAddress.loopbackIPv4,
        18789,
        shared: true,
      );
      expect(await service.isServiceRunning(), isTrue);
    } on SocketException {
      // If something else in the environment already occupies the port,
      // the service should still report "running".
      expect(await service.isServiceRunning(), isTrue);
    } finally {
      await server?.close();
    }
  });

  test('startService does nothing when already running', () async {
    final cmdService = _CommandTestService(running: true);
    await cmdService.startService();
    expect(cmdService.startCalls, 0);
  });

  test('startService starts process when not running', () async {
    final cmdService = _CommandTestService(running: false);
    await cmdService.startService();
    expect(cmdService.startCalls, 1);
    expect(cmdService.running, isTrue);
  });

  test('stopService, restartService and runDoctor use command paths', () async {
    final cmdService = _CommandTestService(running: true)
      ..nextRunResult = ProcessResult(0, 0, 'doctor ok', 'warn');

    await cmdService.stopService();
    expect(cmdService.runCalls, 1);
    expect(cmdService.running, isFalse);

    await cmdService.restartService();
    expect(cmdService.runCalls, 2);
    expect(cmdService.startCalls, 1);
    expect(cmdService.running, isTrue);

    final doctorOutput = await cmdService.runDoctor();
    expect(cmdService.runCalls, 3);
    expect(doctorOutput, contains('doctor ok'));
    expect(doctorOutput, contains('warn'));
  });

  test('default processRun executes a shell command', () async {
    final defaultService = LocalManagerService(homeOverride: tempHome.path);
    ProcessResult result;

    if (Platform.isWindows) {
      result = await defaultService.processRun('cmd', ['/c', 'echo', 'ok']);
    } else {
      result = await defaultService.processRun('sh', ['-c', 'echo ok']);
    }

    expect(result.stdout.toString().toLowerCase(), contains('ok'));
  });

  test('default processStart starts and exits a shell command', () async {
    final defaultService = LocalManagerService(homeOverride: tempHome.path);
    Process process;

    if (Platform.isWindows) {
      process = await defaultService.processStart('cmd', ['/c', 'exit', '0']);
    } else {
      process = await defaultService.processStart('sh', ['-c', 'exit 0']);
    }

    expect(await process.exitCode, 0);
  });
}
