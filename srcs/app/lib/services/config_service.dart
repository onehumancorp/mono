import 'dart:convert';
import 'dart:io';
import 'package:path/path.dart' as p;
import 'package:path_provider/path_provider.dart';

class ConfigService {
  static final ConfigService _instance = ConfigService._internal();
  factory ConfigService() => _instance;
  ConfigService._internal();

  Future<Directory> get _openclawDir async {
    final home = Platform.environment['HOME'] ?? Platform.environment['USERPROFILE'];
    if (home == null) throw Exception("Cannot determine home directory");
    return Directory(p.join(home, '.openclaw'));
  }

  Future<File> get _configFile async {
    final dir = await _openclawDir;
    return File(p.join(dir.path, 'openclaw.json'));
  }

  Future<File> get _envFile async {
    final dir = await _openclawDir;
    return File(p.join(dir.path, '.env'));
  }

  Future<Map<String, dynamic>> getConfig() async {
    final file = await _configFile;
    if (!await file.exists()) return {};
    final content = await file.readAsString();
    return jsonDecode(content) as Map<String, dynamic>;
  }

  Future<void> saveConfig(Map<String, dynamic> config) async {
    final file = await _configFile;
    if (!await file.parent.exists()) {
      await file.parent.create(recursive: true);
    }
    const encoder = JsonEncoder.withIndent('  ');
    await file.writeAsString(encoder.convert(config));
  }

  Future<String?> getEnvValue(String key) async {
    final file = await _envFile;
    if (!await file.exists()) return null;
    final lines = await file.readAsLines();
    for (final line in lines) {
      if (line.startsWith('$key=')) {
        return line.substring(key.length + 1).replaceAll('"', '').trim();
      }
    }
    return null;
  }

  Future<void> saveEnvValue(String key, String value) async {
    final file = await _envFile;
    if (!await file.parent.exists()) {
      await file.parent.create(recursive: true);
    }
    List<String> lines = [];
    if (await file.exists()) {
      lines = await file.readAsLines();
    }

    final prefix = '$key=';
    final newLine = '$key="$value"';
    bool found = false;
    for (int i = 0; i < lines.length; i++) {
      if (lines[i].startsWith(prefix)) {
        lines[i] = newLine;
        found = true;
        break;
      }
    }
    if (!found) {
      lines.add(newLine);
    }
    await file.writeAsString(lines.join('\n') + '\n');
  }

  Future<List<Map<String, dynamic>>> getAgents() async {
    final config = await getConfig();
    final agents = config['agents'] as List<dynamic>? ?? [];
    return agents.cast<Map<String, dynamic>>();
  }

  Future<void> saveAgent(Map<String, dynamic> agent) async {
    final config = await getConfig();
    final agents = (config['agents'] as List<dynamic>? ?? []).cast<Map<String, dynamic>>().toList();
    
    final id = agent['id'];
    final index = agents.indexWhere((a) => a['id'] == id);
    if (index != -1) {
      agents[index] = agent;
    } else {
      agents.add(agent);
    }
    
    config['agents'] = agents;
    await saveConfig(config);
  }
}
