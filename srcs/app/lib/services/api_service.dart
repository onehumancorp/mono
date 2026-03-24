import 'dart:convert';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:http/http.dart' as http;
import 'package:ohc_app/models/agent.dart';
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

  // ── Chat ─────────────────────────────────────────────────────────────────

  Future<void> sendMessage(String content, {String? channelId}) async {
    final res = await _client.post(
      Uri.parse('$baseUrl/api/messages'),
      headers: _headers,
      body: jsonEncode({
        'content': content,
        if (channelId != null) 'channel_id': channelId,
      }),
    );
    _checkStatus(res);
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
