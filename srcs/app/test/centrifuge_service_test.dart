import 'package:flutter_test/flutter_test.dart';
import 'package:ohc_app/services/centrifuge_service.dart';

void main() {
  group('CentrifugeMessage', () {
    test('fromJson parses all fields', () {
      final json = {
        'id': 'msg-1',
        'channel_id': 'general',
        'author_id': 'user-1',
        'author_name': 'Alice',
        'body': 'Hello, world!',
        'sent_at': '2025-01-01T10:00:00.000Z',
      };
      final msg = CentrifugeMessage.fromJson(json);
      expect(msg.id, 'msg-1');
      expect(msg.channelId, 'general');
      expect(msg.authorId, 'user-1');
      expect(msg.authorName, 'Alice');
      expect(msg.body, 'Hello, world!');
      expect(msg.sentAt, DateTime.utc(2025, 1, 1, 10));
    });

    test('fromJson uses defaults for missing fields', () {
      final json = <String, dynamic>{};
      final msg = CentrifugeMessage.fromJson(json);
      expect(msg.id, '');
      expect(msg.channelId, '');
      expect(msg.authorName, 'Unknown');
      expect(msg.body, '');
    });

    test('toJson round-trips', () {
      final now = DateTime.utc(2025, 6, 15, 12, 30);
      final msg = CentrifugeMessage(
        id: 'msg-2',
        channelId: 'support',
        authorId: 'agent-1',
        authorName: 'Agent',
        body: 'How can I help?',
        sentAt: now,
      );
      final json = msg.toJson();
      expect(json['id'], 'msg-2');
      expect(json['channel_id'], 'support');
      expect(json['author_id'], 'agent-1');
      expect(json['author_name'], 'Agent');
      expect(json['body'], 'How can I help?');
      expect(json['sent_at'], now.toIso8601String());
    });
  });

  group('CentrifugeService construction', () {
    test('creates service with correct fields', () {
      final svc = CentrifugeService(
        serverUrl: 'ws://localhost:8000/connection/websocket',
        token: 'test-token',
        userId: 'user-1',
        userName: 'Alice',
      );
      expect(svc.serverUrl, 'ws://localhost:8000/connection/websocket');
      expect(svc.userId, 'user-1');
      expect(svc.userName, 'Alice');
    });
  });
}
