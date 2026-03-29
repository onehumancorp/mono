import 'dart:convert';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:http/http.dart' as http;
import 'package:ohc_app/models/agent.dart';
import 'package:ohc_app/models/ai_provider.dart';
import 'package:ohc_app/models/channel.dart';
import 'package:ohc_app/models/security_issue.dart';
import 'package:ohc_app/models/skill.dart';
import 'package:ohc_app/services/auth_service.dart';

/// API client for the OHC backend.
class ApiService {
  final String baseUrl;
  final String token;
  final http.Client _client;

  ApiService({
    required this.baseUrl,
    required this.token,
    http.Client? client,
  }) : _client = client ?? http.Client();

  Map<String, String> get _headers => {
        'Content-Type': 'application/json',
        'Authorization': 'Bearer $token',
      };

  // ── Agents ──────────────────────────────────────────────────────────────

  Future<List<Agent>> listAgents() async {
    final res = await _client.get(Uri.parse('$baseUrl/api/agents'), headers: _headers);
    _checkStatus(res);
    final list = jsonDecode(res.body) as List<dynamic>;
    return list.map((e) => Agent.fromJson(e as Map<String, dynamic>)).toList();
  }

  Future<Agent> hireAgent(String name, String role) async {
    final res = await _client.post(
      Uri.parse('$baseUrl/api/agents/hire'),
      headers: _headers,
      body: jsonEncode({'name': name, 'role': role}),
    );
    _checkStatus(res);
    return Agent.fromJson(jsonDecode(res.body) as Map<String, dynamic>);
  }

  Future<void> fireAgent(String agentId) async {
    final res = await _client.post(
      Uri.parse('$baseUrl/api/agents/fire'),
      headers: _headers,
      body: jsonEncode({'agent_id': agentId}),
    );
    _checkStatus(res);
  }

  // ── Dashboard ────────────────────────────────────────────────────────────

  Future<Map<String, dynamic>> getDashboard() async {
    final res = await _client.get(Uri.parse('$baseUrl/api/dashboard'), headers: _headers);
    _checkStatus(res);
    return jsonDecode(res.body) as Map<String, dynamic>;
  }

  // ── Meetings ─────────────────────────────────────────────────────────────

  Future<List<Map<String, dynamic>>> listMeetings() async {
    final res = await _client.get(Uri.parse('$baseUrl/api/meetings'), headers: _headers);
    _checkStatus(res);
    final list = jsonDecode(res.body) as List<dynamic>;
    return list.cast<Map<String, dynamic>>();
  }

  Future<Map<String, dynamic>> createMeeting(String name) async {
    final res = await _client.post(
      Uri.parse('$baseUrl/api/meetings'),
      headers: _headers,
      body: jsonEncode({'name': name}),
    );
    _checkStatus(res);
    return jsonDecode(res.body) as Map<String, dynamic>;
  }

  Future<Map<String, dynamic>> joinMeeting(String meetingId) async {
    final res = await _client.post(
      Uri.parse('$baseUrl/api/meetings/$meetingId/join'),
      headers: _headers,
    );
    _checkStatus(res);
    return jsonDecode(res.body) as Map<String, dynamic>;
  }

  // ── Channels ─────────────────────────────────────────────────────────────

  Future<List<ChatChannel>> listChannels() async {
    final res = await _client.get(Uri.parse('$baseUrl/api/channels'), headers: _headers);
    _checkStatus(res);
    final list = jsonDecode(res.body) as List<dynamic>;
    return list
        .map((e) => ChatChannel.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  Future<ChatChannel> addChannel({
    required String name,
    required String backend,
    required Map<String, String> config,
  }) async {
    final res = await _client.post(
      Uri.parse('$baseUrl/api/channels'),
      headers: _headers,
      body: jsonEncode({'name': name, 'backend': backend, 'config': config}),
    );
    _checkStatus(res);
    return ChatChannel.fromJson(jsonDecode(res.body) as Map<String, dynamic>);
  }

  Future<void> deleteChannel(String channelId) async {
    final res = await _client.delete(
      Uri.parse('$baseUrl/api/channels/$channelId'),
      headers: _headers,
    );
    _checkStatus(res);
  }

  // ── AI Providers ──────────────────────────────────────────────────────────

  Future<List<AiProvider>> listAiProviders() async {
    final res = await _client.get(Uri.parse('$baseUrl/api/ai/providers'), headers: _headers);
    _checkStatus(res);
    final list = jsonDecode(res.body) as List<dynamic>;
    return list
        .map((e) => AiProvider.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  Future<AiProvider> addAiProvider({
    required String name,
    required String baseUrl,
    required String apiKey,
    required List<String> models,
  }) async {
    final res = await _client.post(
      Uri.parse('${this.baseUrl}/api/ai/providers'),
      headers: _headers,
      body: jsonEncode({
        'name': name,
        'base_url': baseUrl,
        'api_key': apiKey,
        'models': models,
      }),
    );
    _checkStatus(res);
    return AiProvider.fromJson(jsonDecode(res.body) as Map<String, dynamic>);
  }

  Future<void> saveAiProviderKey(String providerId, String apiKey) async {
    final res = await _client.patch(
      Uri.parse('$baseUrl/api/ai/providers/$providerId'),
      headers: _headers,
      body: jsonEncode({'api_key': apiKey}),
    );
    _checkStatus(res);
  }

  // ── Skills ────────────────────────────────────────────────────────────────

  Future<List<Skill>> listSkills() async {
    final res = await _client.get(Uri.parse('$baseUrl/api/skills'), headers: _headers);
    _checkStatus(res);
    final list = jsonDecode(res.body) as List<dynamic>;
    return list.map((e) => Skill.fromJson(e as Map<String, dynamic>)).toList();
  }

  Future<void> installSkill(String name) async {
    final res = await _client.post(
      Uri.parse('$baseUrl/api/skills/$name/install'),
      headers: _headers,
    );
    _checkStatus(res);
  }

  Future<void> uninstallSkill(String name) async {
    final res = await _client.post(
      Uri.parse('$baseUrl/api/skills/$name/uninstall'),
      headers: _headers,
    );
    _checkStatus(res);
  }

  Future<void> setSkillEnabled(String name, bool enabled) async {
    final res = await _client.patch(
      Uri.parse('$baseUrl/api/skills/$name'),
      headers: _headers,
      body: jsonEncode({'enabled': enabled}),
    );
    _checkStatus(res);
  }

  // ── Security ──────────────────────────────────────────────────────────────

  Future<List<SecurityIssue>> listSecurityIssues() async {
    final res = await _client.get(Uri.parse('$baseUrl/api/security/issues'), headers: _headers);
    _checkStatus(res);
    final list = jsonDecode(res.body) as List<dynamic>;
    return list
        .map((e) => SecurityIssue.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  Future<void> fixSecurityIssue(String issueId) async {
    final res = await _client.post(
      Uri.parse('$baseUrl/api/security/issues/$issueId/fix'),
      headers: _headers,
    );
    _checkStatus(res);
  }

  // ── Scaling ──────────────────────────────────────────────────────────────

  Future<void> scaleRole(String role, int count) async {
    final res = await _client.post(
      Uri.parse('$baseUrl/api/v1/scale'),
      headers: _headers,
      body: jsonEncode({'role': role, 'count': count}),
    );
    _checkStatus(res);
  }

  Future<http.StreamedResponse> scaleStream() async {
    final request = http.Request('GET', Uri.parse('$baseUrl/api/v1/scale/stream'));
    request.headers.addAll(_headers);
    return await _client.send(request);
  }

  // ── Service logs ──────────────────────────────────────────────────────────

  Future<List<String>> getLogs({int lines = 100}) async {
    final res = await _client.get(
      Uri.parse('$baseUrl/api/service/logs?lines=$lines'),
      headers: _headers,
    );
    _checkStatus(res);
    final list = jsonDecode(res.body) as List<dynamic>;
    return list.cast<String>();
  }

  // ── Helpers ──────────────────────────────────────────────────────────────

  void _checkStatus(http.Response res) {
    if (res.statusCode < 200 || res.statusCode >= 300) {
      throw Exception('API error ${res.statusCode}: ${res.body}');
    }
  }
}

// ── Providers ──────────────────────────────────────────────────────────────

final apiServiceProvider = Provider<ApiService?>((ref) {
  final user = ref.watch(authStateProvider).valueOrNull;
  if (user == null) return null;
  final url = ref.watch(backendUrlProvider);
  return ApiService(baseUrl: url, token: user.token);
});
