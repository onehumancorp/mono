import 'dart:io';
import 'package:url_launcher/url_launcher.dart';

class InstallerService {
  Future<Map<String, bool>> checkEnvironment() async {
    final nodeOk = await _checkExecutable('node');
    final npmOk = await _checkExecutable('npm');
    final openclawOk = await _checkExecutable('openclaw');
    
    return {
      'node': nodeOk,
      'npm': npmOk,
      'openclaw': openclawOk,
    };
  }

  Future<bool> _checkExecutable(String cmd) async {
    try {
      final result = await Process.run(cmd, ['--version']);
      return result.exitCode == 0;
    } catch (_) {
      return false;
    }
  }

  Future<void> openNodeDownloadPage() async {
    final url = Uri.parse('https://nodejs.org/en/download');
    if (await canLaunchUrl(url)) {
      await launchUrl(url, mode: LaunchMode.externalApplication);
    }
  }

  Future<String?> installOpenClaw() async {
    try {
      final result = await Process.run('npm', ['install', '-g', 'openclaw']);
      if (result.exitCode == 0) {
        return null; // Success
      } else {
        return result.stderr as String;
      }
    } catch (e) {
      return e.toString();
    }
  }

  Future<String?> updateOpenClaw() async {
    try {
      final result = await Process.run('npm', ['install', '-g', 'openclaw@latest']);
      if (result.exitCode == 0) {
        return null; // Success
      } else {
        return result.stderr as String;
      }
    } catch (e) {
      return e.toString();
    }
  }
}
