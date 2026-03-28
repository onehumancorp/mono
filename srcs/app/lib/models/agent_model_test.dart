import 'package:flutter_test/flutter_test.dart';
import 'package:ohc_app/models/agent.dart';

void main() {
  group('Agent model', () {
    test('fromJson parses all fields', () {
      final json = {
        'id': 'a1',
        'name': 'Alice',
        'role': 'engineer',
        'status': 'running',
        'organization_id': 'org-1',
        'created_at': '2025-01-01T00:00:00Z',
      };
      final agent = Agent.fromJson(json);
      expect(agent.id, 'a1');
      expect(agent.name, 'Alice');
      expect(agent.role, 'engineer');
      expect(agent.status, 'running');
      expect(agent.organizationId, 'org-1');
      expect(agent.isRunning, isTrue);
      expect(agent.isPending, isFalse);
    });

    test('fromJson uses defaults for missing optional fields', () {
      final json = {'id': 'b2', 'name': 'Bob'};
      final agent = Agent.fromJson(json);
      expect(agent.role, '');
      expect(agent.status, 'pending');
      expect(agent.organizationId, '');
      expect(agent.isPending, isTrue);
    });

    test('toJson round-trips', () {
      final json = {
        'id': 'c3',
        'name': 'Carol',
        'role': 'ceo',
        'status': 'pending',
        'organization_id': 'org-2',
        'created_at': '2025-06-01T12:00:00.000Z',
      };
      final agent = Agent.fromJson(json);
      final out = agent.toJson();
      expect(out['id'], 'c3');
      expect(out['name'], 'Carol');
      expect(out['role'], 'ceo');
    });
  });
}
