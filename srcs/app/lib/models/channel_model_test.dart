import 'package:flutter_test/flutter_test.dart';
import 'package:ohc_app/models/channel.dart';

void main() {
  group('ChatBackend', () {
    test('fromJson parses string backend (slack)', () {
      final b = ChatBackend.fromJson('slack');
      expect(b.type, ChatBackendType.slack);
      expect(b.displayName, 'Slack');
    });

    test('fromJson parses centrifuge backend with url', () {
      final b = ChatBackend.fromJson({
        'centrifuge': 'ws://localhost:8000/connection/websocket',
      });
      expect(b.type, ChatBackendType.centrifuge);
      expect(b.url, 'ws://localhost:8000/connection/websocket');
      expect(b.displayName, 'Native (Centrifuge)');
    });

    test('fromJson parses webhook backend with url', () {
      final b = ChatBackend.fromJson({'webhook': 'https://example.com/hook'});
      expect(b.type, ChatBackendType.webhook);
      expect(b.url, 'https://example.com/hook');
    });

    test('toJson for centrifuge round-trips', () {
      final b = ChatBackend(
          type: ChatBackendType.centrifuge,
          url: 'ws://localhost:8000/connection/websocket');
      final json = b.toJson();
      expect(json['centrifuge'], 'ws://localhost:8000/connection/websocket');
    });

    test('all backend types have display names', () {
      for (final type in ChatBackendType.values) {
        final b = ChatBackend(type: type);
        expect(b.displayName, isNotEmpty);
      }
    });
  });

  group('ChatChannel', () {
    test('fromJson parses all fields', () {
      final json = {
        'id': 'ch-1',
        'organization_id': 'org-1',
        'name': 'general',
        'backend': 'slack',
        'config': {'bot_token': 'xoxb-...'},
        'enabled': true,
        'created_at': '2025-01-01T00:00:00.000Z',
      };
      final ch = ChatChannel.fromJson(json);
      expect(ch.id, 'ch-1');
      expect(ch.organizationId, 'org-1');
      expect(ch.name, 'general');
      expect(ch.backend.type, ChatBackendType.slack);
      expect(ch.enabled, isTrue);
      expect(ch.config['bot_token'], 'xoxb-...');
    });

    test('toJson round-trips', () {
      final ch = ChatChannel(
        id: 'ch-2',
        organizationId: 'org-2',
        name: 'support',
        backend: const ChatBackend(type: ChatBackendType.telegram),
        config: const {'bot_token': 'abc123'},
        enabled: false,
        createdAt: DateTime.utc(2025, 3, 1),
      );
      final json = ch.toJson();
      expect(json['id'], 'ch-2');
      expect(json['name'], 'support');
      expect(json['enabled'], isFalse);
    });
  });
}
