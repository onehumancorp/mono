import 'package:flutter_test/flutter_test.dart';
import 'package:ohc_app/models/skill.dart';

void main() {
  group('Skill model', () {
    test('fromJson parses all fields', () {
      final json = {
        'name': 'web_search',
        'version': '1.2.3',
        'description': 'Search the web',
        'category': 'official',
        'installed': true,
        'enabled': true,
      };
      final skill = Skill.fromJson(json);
      expect(skill.name, 'web_search');
      expect(skill.version, '1.2.3');
      expect(skill.description, 'Search the web');
      expect(skill.category, 'official');
      expect(skill.installed, isTrue);
      expect(skill.enabled, isTrue);
    });

    test('fromJson uses defaults for missing optional fields', () {
      final json = <String, dynamic>{};
      final skill = Skill.fromJson(json);
      expect(skill.name, '');
      expect(skill.version, '0.0.0');
      expect(skill.description, '');
      expect(skill.category, 'community');
      expect(skill.installed, isFalse);
      expect(skill.enabled, isFalse);
    });

    test('toJson round-trips', () {
      const skill = Skill(
        name: 'code_runner',
        version: '2.0.0',
        description: 'Run code snippets',
        category: 'official',
        installed: false,
        enabled: false,
      );
      final json = skill.toJson();
      expect(json['name'], 'code_runner');
      expect(json['version'], '2.0.0');
      expect(json['description'], 'Run code snippets');
      expect(json['category'], 'official');
      expect(json['installed'], isFalse);
      expect(json['enabled'], isFalse);
    });

    test('copyWith changes installed and enabled', () {
      const original = Skill(
        name: 'test_skill',
        version: '1.0.0',
        description: 'A test skill',
        category: 'community',
        installed: false,
        enabled: false,
      );
      final installed = original.copyWith(installed: true);
      expect(installed.installed, isTrue);
      expect(installed.enabled, isFalse);

      final enabled = installed.copyWith(enabled: true);
      expect(enabled.installed, isTrue);
      expect(enabled.enabled, isTrue);
    });

    test('copyWith without arguments returns same values', () {
      const original = Skill(
        name: 'n',
        version: '0.1.0',
        description: 'd',
        category: 'community',
        installed: true,
        enabled: false,
      );
      final copy = original.copyWith();
      expect(copy.name, original.name);
      expect(copy.version, original.version);
      expect(copy.installed, original.installed);
      expect(copy.enabled, original.enabled);
    });
  });
}
