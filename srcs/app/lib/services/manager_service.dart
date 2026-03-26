import 'dart:io';
import 'package:path/path.dart' as p;
import 'package:riverpod_annotation/riverpod_annotation.dart';

class ServiceStatus {
  final bool running;
  final int? pid;
  final int port;
  final double? memoryMb;
  final double? cpuPercent;
  final int? uptimeSeconds;

  ServiceStatus({
    required this.running,
    this.pid,
    required this.port,
    this.memoryMb,
    this.cpuPercent,
    this.uptimeSeconds,
  });
}

class ManagerService {
  static const int openclawPort = 18789;

  Future<ServiceStatus> getStatus() async {
    final pid = await _findPortPid();
    final running = pid != null;
    
    // Simplification: In a real app we'd use a package like system_info2
    // For now, we return basic status.
    return ServiceStatus(
      running: running,
      pid: pid,
      port: openclawPort,
    );
  }

  Future<int?> _findPortPid() async {
    try {
      if (Platform.isWindows) {
        final result = await Process.run('netstat', ['-ano']);
        final stdout = result.stdout as String;
        for (final line in stdout.split('\n')) {
          if (line.contains(':$openclawPort') && line.contains('LISTENING')) {
            final parts = line.trim().split(RegExp(r'\s+'));
            if (parts.isNotEmpty) {
              return int.tryParse(parts.last);
            }
          }
        }
      } else {
        final result = await Process.run('lsof', ['-ti', ':$openclawPort']);
        final stdout = (result.stdout as String).trim();
        if (stdout.isNotEmpty) {
          return int.tryParse(stdout.split('\n').first);
        }
      }
    } catch (_) {
      // Handle error (e.g. lsof not installed)
    }
    return null;
  }

  Future<void> startService() async {
    if (await _findPortPid() != null) return;

    final home = Platform.environment['HOME'] ?? Platform.environment['USERPROFILE']!;
    final logFile = File(p.join(home, '.openclaw', 'openclaw.log'));
    if (!await logFile.parent.exists()) {
      await logFile.parent.create(recursive: true);
    }

    // Start detached process. For robustness, we'd find the full path to 'openclaw'.
    await Process.start(
      'openclaw',
      ['gateway'],
      mode: ProcessStartMode.detachedWithStdio,
    ).then((process) {
      process.stdout.pipe(logFile.openWrite(mode: FileMode.append));
      process.stderr.pipe(logFile.openWrite(mode: FileMode.append));
    });
  }

  Future<void> stopService() async {
    final pid = await _findPortPid();
    if (pid == null) return;

    if (Platform.isWindows) {
      await Process.run('taskkill', ['/PID', pid.toString(), '/F']);
    } else {
      await Process.run('kill', ['-TERM', pid.toString()]);
    }
  }

  Future<List<String>> getLogs(int lines) async {
    final home = Platform.environment['HOME'] ?? Platform.environment['USERPROFILE']!;
    final logFile = File(p.join(home, '.openclaw', 'openclaw.log'));
    if (!await logFile.exists()) return [];

    final content = await logFile.readAsLines();
    final start = (content.length - lines).clamp(0, content.length);
    return content.sublist(start);
  }
}
