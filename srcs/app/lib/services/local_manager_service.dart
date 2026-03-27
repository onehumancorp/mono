import 'dart:convert';
import 'dart:io';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:path/path.dart' as p;

/// Manages the local OpenClaw service and its configuration.
class LocalManagerService {
  Directory get _openclawDir {
    final home = Platform.environment['HOME'] ?? Platform.environment['USERPROFILE'] ?? '.';
    return Directory(p.join(home, '.openclaw'));
  }

  File get _configFile => File(p.join(_openclawDir.path, 'openclaw.json'));
  File get _envFile => File(p.join(_openclawDir.path, '.env'));

  // ── Service Management ───────────────────────────────────────────────────

  Future<bool> isServiceRunning() async {
    try {
      // Check if something is listening on the default port
      final socket = await Socket.connect('localhost', 18789, timeout: const Duration(milliseconds: 500));
      socket.destroy();
      return true;
    } catch (_) {
      return false;
    }
  }

  Future<void> startService() async {
    if (await isServiceRunning()) return;
    await Process.start('openclaw', ['start', '--daemon'], runInShell: true);
  }

  Future<void> stopService() async {
    await Process.run('openclaw', ['stop'], runInShell: true);
  }

  Future<void> restartService() async {
    await stopService();
    await startService();
  }

  // ── Config Management ────────────────────────────────────────────────────

  Future<Map<String, dynamic>> readConfig() async {
    if (!await _configFile.exists()) {
      return {};
    }
    final content = await _configFile.readAsString();
    return jsonDecode(content) as Map<String, dynamic>;
  }

  Future<void> writeConfig(Map<String, dynamic> config) async {
    if (!await _openclawDir.exists()) {
      await _openclawDir.create(recursive: true);
    }
    const encoder = JsonEncoder.withIndent('  ');
    await _configFile.writeAsString(encoder.convert(config));
  }

  Future<String?> getEnvValue(String key) async {
    if (!await _envFile.exists()) return null;
    final lines = await _envFile.readAsLines();
    for (var line in lines) {
      if (line.startsWith('$key=')) {
        return line.substring(key.length + 1).replaceAll('"', '').trim();
      }
    }
    return null;
  }

  Future<void> saveEnvValue(String key, String value) async {
    if (!await _openclawDir.exists()) {
      await _openclawDir.create(recursive: true);
    }
    List<String> lines = [];
    if (await _envFile.exists()) {
      lines = await _envFile.readAsLines();
    }

    bool found = false;
    for (int i = 0; i < lines.length; i++) {
      if (lines[i].startsWith('$key=')) {
        lines[i] = '$key="$value"';
        found = true;
        break;
      }
    }

    if (!found) {
      lines.add('$key="$value"');
    }

    await _envFile.writeAsString(lines.join('\n') + '\n');
  }

  // ── Diagnostics ──────────────────────────────────────────────────────────

  Future<String> runDoctor() async {
    final result = await Process.run('openclaw', ['doctor'], runInShell: true);
    return result.stdout.toString() + result.stderr.toString();
  }

  Future<Map<String, dynamic>> getSystemInfo() async {
    return {
      'os': Platform.operatingSystem,
      'os_version': Platform.operatingSystemVersion,
      'dart_version': Platform.version,
      'hostname': Platform.localHostname,
      'cpus': Platform.numberOfProcessors,
    };
  }
}

final localManagerServiceProvider = Provider((ref) => LocalManagerService());
