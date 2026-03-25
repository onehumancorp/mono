import 'dart:convert';

import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:mocktail/mocktail.dart';
import 'package:ohc_app/models/agent.dart';
import 'package:ohc_app/models/ai_provider.dart';
import 'package:ohc_app/models/channel.dart';
import 'package:ohc_app/models/security_issue.dart';
import 'package:ohc_app/models/skill.dart';
import 'package:ohc_app/services/api_service.dart';

class MockHttpClient extends Mock implements http.Client {}

class FakeUri extends Fake implements Uri {}

void main() {
  setUpAll(() {
    registerFallbackValue(FakeUri());
  });

  late MockHttpClient mockClient;
  late ApiService api;

  setUp(() {
    mockClient = MockHttpClient();
    api = ApiService(
      baseUrl: 'http://localhost:8080',
      token: 'test-token',
      client: mockClient,
    );
  });

  // ── Agents ────────────────────────────────────────────────────────────────

  group('ApiService agents', () {
    test('listAgents returns list of agents', () async {
      final agentJson = [
        {
          'id': 'a1',
          'name': 'Alice',
          'role': 'engineer',
          'status': 'running',
          'organization_id': 'org-1',
          'created_at': '2025-01-01T00:00:00Z',
        }
      ];
      when(() => mockClient.get(any(), headers: any(named: 'headers')))
          .thenAnswer((_) async => http.Response(jsonEncode(agentJson), 200));

      final agents = await api.listAgents();
      expect(agents, hasLength(1));
      expect(agents.first.id, 'a1');
    });

    test('hireAgent returns new agent', () async {
      final agentJson = {
        'id': 'a2',
        'name': 'Bob',
        'role': 'manager',
        'status': 'pending',
        'organization_id': 'org-1',
        'created_at': '2025-01-02T00:00:00Z',
      };
      when(() => mockClient.post(
            any(),
            headers: any(named: 'headers'),
            body: any(named: 'body'),
          )).thenAnswer((_) async => http.Response(jsonEncode(agentJson), 200));

      final agent = await api.hireAgent('Bob', 'manager');
      expect(agent.id, 'a2');
      expect(agent.name, 'Bob');
    });

    test('fireAgent sends post request', () async {
      when(() => mockClient.post(
            any(),
            headers: any(named: 'headers'),
            body: any(named: 'body'),
          )).thenAnswer((_) async => http.Response('', 200));

      await api.fireAgent('a1');
      verify(() => mockClient.post(any(),
          headers: any(named: 'headers'),
          body: any(named: 'body'))).called(1);
    });

    test('listAgents throws on error status', () async {
      when(() => mockClient.get(any(), headers: any(named: 'headers')))
          .thenAnswer((_) async => http.Response('Forbidden', 403));

      expect(() => api.listAgents(), throwsA(isA<Exception>()));
    });
  });

  // ── Dashboard ─────────────────────────────────────────────────────────────

  group('ApiService dashboard', () {
    test('getDashboard returns map', () async {
      final data = {'active_agents': 5, 'pending_tasks': 2};
      when(() => mockClient.get(any(), headers: any(named: 'headers')))
          .thenAnswer((_) async => http.Response(jsonEncode(data), 200));

      final result = await api.getDashboard();
      expect(result['active_agents'], 5);
    });
  });

  // ── Meetings ──────────────────────────────────────────────────────────────

  group('ApiService meetings', () {
    test('listMeetings returns list', () async {
      final meetings = [
        {'id': 'm1', 'name': 'Standup'}
      ];
      when(() => mockClient.get(any(), headers: any(named: 'headers')))
          .thenAnswer((_) async => http.Response(jsonEncode(meetings), 200));

      final result = await api.listMeetings();
      expect(result, hasLength(1));
      expect(result.first['id'], 'm1');
    });

    test('createMeeting returns new meeting', () async {
      final meeting = {'id': 'm2', 'name': 'Sync'};
      when(() => mockClient.post(
            any(),
            headers: any(named: 'headers'),
            body: any(named: 'body'),
          )).thenAnswer((_) async => http.Response(jsonEncode(meeting), 200));

      final result = await api.createMeeting('Sync');
      expect(result['id'], 'm2');
    });

    test('joinMeeting returns join info', () async {
      final info = {'token': 'lk-token', 'url': 'wss://lk.example.com'};
      when(() => mockClient.post(
            any(),
            headers: any(named: 'headers'),
          )).thenAnswer((_) async => http.Response(jsonEncode(info), 200));

      final result = await api.joinMeeting('m1');
      expect(result['token'], 'lk-token');
    });
  });

  // ── Channels ──────────────────────────────────────────────────────────────

  group('ApiService channels', () {
    test('listChannels returns list', () async {
      final channels = [
        {
          'id': 'ch1',
          'organization_id': 'org-1',
          'name': 'general',
          'backend': 'slack',
          'config': <String, String>{},
          'enabled': true,
          'created_at': '2025-01-01T00:00:00Z',
        }
      ];
      when(() => mockClient.get(any(), headers: any(named: 'headers')))
          .thenAnswer((_) async => http.Response(jsonEncode(channels), 200));

      final result = await api.listChannels();
      expect(result, hasLength(1));
      expect(result.first.name, 'general');
    });

    test('addChannel returns new channel', () async {
      final channelJson = {
        'id': 'ch2',
        'organization_id': 'org-1',
        'name': 'support',
        'backend': 'slack',
        'config': <String, String>{'bot_token': 'xoxb-...'},
        'enabled': true,
        'created_at': '2025-02-01T00:00:00Z',
      };
      when(() => mockClient.post(
            any(),
            headers: any(named: 'headers'),
            body: any(named: 'body'),
          )).thenAnswer((_) async =>
          http.Response(jsonEncode(channelJson), 200));

      final result = await api.addChannel(
        name: 'support',
        backend: 'slack',
        config: {'bot_token': 'xoxb-...'},
      );
      expect(result.id, 'ch2');
    });

    test('deleteChannel calls delete endpoint', () async {
      when(() => mockClient.delete(any(), headers: any(named: 'headers')))
          .thenAnswer((_) async => http.Response('', 200));

      await api.deleteChannel('ch1');
      verify(() => mockClient.delete(any(),
          headers: any(named: 'headers'))).called(1);
    });
  });

  // ── AI Providers ──────────────────────────────────────────────────────────

  group('ApiService AI providers', () {
    test('listAiProviders returns list', () async {
      final providers = [
        {
          'id': 'p1',
          'name': 'OpenAI',
          'base_url': 'https://api.openai.com/v1',
          'api_key': 'sk-...',
          'models': ['gpt-4'],
          'is_official': true,
        }
      ];
      when(() => mockClient.get(any(), headers: any(named: 'headers')))
          .thenAnswer((_) async => http.Response(jsonEncode(providers), 200));

      final result = await api.listAiProviders();
      expect(result, hasLength(1));
      expect(result.first.name, 'OpenAI');
    });

    test('addAiProvider returns new provider', () async {
      final providerJson = {
        'id': 'p2',
        'name': 'Anthropic',
        'base_url': 'https://api.anthropic.com',
        'api_key': 'sk-ant-...',
        'models': ['claude-3'],
        'is_official': false,
      };
      when(() => mockClient.post(
            any(),
            headers: any(named: 'headers'),
            body: any(named: 'body'),
          )).thenAnswer((_) async =>
          http.Response(jsonEncode(providerJson), 200));

      final result = await api.addAiProvider(
        name: 'Anthropic',
        baseUrl: 'https://api.anthropic.com',
        apiKey: 'sk-ant-...',
        models: ['claude-3'],
      );
      expect(result.id, 'p2');
    });

    test('saveAiProviderKey calls patch endpoint', () async {
      when(() => mockClient.patch(
            any(),
            headers: any(named: 'headers'),
            body: any(named: 'body'),
          )).thenAnswer((_) async => http.Response('', 200));

      await api.saveAiProviderKey('p1', 'new-key');
      verify(() => mockClient.patch(any(),
          headers: any(named: 'headers'),
          body: any(named: 'body'))).called(1);
    });
  });

  // ── Skills ────────────────────────────────────────────────────────────────

  group('ApiService skills', () {
    test('listSkills returns list', () async {
      final skills = [
        {
          'name': 'web_search',
          'version': '1.0.0',
          'description': 'Search the web',
          'category': 'official',
          'installed': true,
          'enabled': true,
        }
      ];
      when(() => mockClient.get(any(), headers: any(named: 'headers')))
          .thenAnswer((_) async => http.Response(jsonEncode(skills), 200));

      final result = await api.listSkills();
      expect(result, hasLength(1));
      expect(result.first.name, 'web_search');
    });

    test('installSkill calls install endpoint', () async {
      when(() => mockClient.post(
            any(),
            headers: any(named: 'headers'),
          )).thenAnswer((_) async => http.Response('', 200));

      await api.installSkill('web_search');
      verify(() => mockClient.post(any(),
          headers: any(named: 'headers'))).called(1);
    });

    test('uninstallSkill calls uninstall endpoint', () async {
      when(() => mockClient.post(
            any(),
            headers: any(named: 'headers'),
          )).thenAnswer((_) async => http.Response('', 200));

      await api.uninstallSkill('web_search');
      verify(() => mockClient.post(any(),
          headers: any(named: 'headers'))).called(1);
    });

    test('setSkillEnabled calls patch endpoint', () async {
      when(() => mockClient.patch(
            any(),
            headers: any(named: 'headers'),
            body: any(named: 'body'),
          )).thenAnswer((_) async => http.Response('', 200));

      await api.setSkillEnabled('web_search', true);
      verify(() => mockClient.patch(any(),
          headers: any(named: 'headers'),
          body: any(named: 'body'))).called(1);
    });
  });

  // ── Security ──────────────────────────────────────────────────────────────

  group('ApiService security', () {
    test('listSecurityIssues returns list', () async {
      final issues = [
        {
          'id': 'i1',
          'title': 'Weak password',
          'description': 'Use a stronger password',
          'severity': 'high',
          'fixable': true,
          'fixed': false,
          'category': 'auth',
        }
      ];
      when(() => mockClient.get(any(), headers: any(named: 'headers')))
          .thenAnswer((_) async => http.Response(jsonEncode(issues), 200));

      final result = await api.listSecurityIssues();
      expect(result, hasLength(1));
      expect(result.first.id, 'i1');
    });

    test('fixSecurityIssue calls fix endpoint', () async {
      when(() => mockClient.post(
            any(),
            headers: any(named: 'headers'),
          )).thenAnswer((_) async => http.Response('', 200));

      await api.fixSecurityIssue('i1');
      verify(() => mockClient.post(any(),
          headers: any(named: 'headers'))).called(1);
    });
  });

  // ── Service logs ──────────────────────────────────────────────────────────

  group('ApiService logs', () {
    test('getLogs returns list of strings', () async {
      final logs = ['line 1', 'line 2', 'line 3'];
      when(() => mockClient.get(any(), headers: any(named: 'headers')))
          .thenAnswer((_) async => http.Response(jsonEncode(logs), 200));

      final result = await api.getLogs();
      expect(result, hasLength(3));
      expect(result.first, 'line 1');
    });

    test('getLogs with custom lines param', () async {
      when(() => mockClient.get(any(), headers: any(named: 'headers')))
          .thenAnswer((_) async => http.Response(jsonEncode(<String>[]), 200));

      final result = await api.getLogs(lines: 50);
      expect(result, isEmpty);
    });
  });
}
