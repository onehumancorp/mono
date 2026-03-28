import 'package:flutter_test/flutter_test.dart';
import 'package:ohc_app/models/ai_provider.dart';

void main() {
  group('AiProvider model', () {
    test('fromJson parses all fields', () {
      final json = {
        'id': 'prov-1',
        'name': 'OpenAI',
        'base_url': 'https://api.openai.com/v1',
        'api_key': 'sk-abc',
        'models': ['gpt-4', 'gpt-3.5-turbo'],
        'is_official': true,
      };
      final provider = AiProvider.fromJson(json);
      expect(provider.id, 'prov-1');
      expect(provider.name, 'OpenAI');
      expect(provider.baseUrl, 'https://api.openai.com/v1');
      expect(provider.apiKey, 'sk-abc');
      expect(provider.models, ['gpt-4', 'gpt-3.5-turbo']);
      expect(provider.isOfficial, isTrue);
    });

    test('fromJson uses defaults for missing optional fields', () {
      final json = <String, dynamic>{};
      final provider = AiProvider.fromJson(json);
      expect(provider.id, '');
      expect(provider.name, '');
      expect(provider.baseUrl, '');
      expect(provider.apiKey, '');
      expect(provider.models, isEmpty);
      expect(provider.isOfficial, isFalse);
    });

    test('toJson round-trips all fields', () {
      final json = {
        'id': 'prov-2',
        'name': 'Anthropic',
        'base_url': 'https://api.anthropic.com',
        'api_key': 'key-xyz',
        'models': ['claude-3'],
        'is_official': false,
      };
      final provider = AiProvider.fromJson(json);
      final out = provider.toJson();
      expect(out['id'], 'prov-2');
      expect(out['name'], 'Anthropic');
      expect(out['base_url'], 'https://api.anthropic.com');
      expect(out['api_key'], 'key-xyz');
      expect(out['models'], ['claude-3']);
      expect(out['is_official'], isFalse);
    });

    test('copyWith changes only the apiKey', () {
      const original = AiProvider(
        id: 'prov-3',
        name: 'Test',
        baseUrl: 'http://test.example.com',
        apiKey: 'old-key',
        models: ['model-a'],
        isOfficial: false,
      );
      final updated = original.copyWith(apiKey: 'new-key');
      expect(updated.apiKey, 'new-key');
      expect(updated.id, 'prov-3');
      expect(updated.name, 'Test');
      expect(updated.baseUrl, 'http://test.example.com');
      expect(updated.models, ['model-a']);
      expect(updated.isOfficial, isFalse);
    });

    test('copyWith without arguments returns same values', () {
      const original = AiProvider(
        id: 'prov-4',
        name: 'Local',
        baseUrl: 'http://localhost:11434',
        apiKey: '',
        models: [],
        isOfficial: false,
      );
      final copy = original.copyWith();
      expect(copy.id, original.id);
      expect(copy.name, original.name);
      expect(copy.apiKey, original.apiKey);
    });
  });
}
