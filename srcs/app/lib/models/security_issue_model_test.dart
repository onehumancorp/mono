import 'package:flutter_test/flutter_test.dart';
import 'package:ohc_app/models/security_issue.dart';

void main() {
  group('SecurityIssue model', () {
    test('fromJson parses all fields including detail', () {
      final json = {
        'id': 'issue-1',
        'title': 'Weak SSH key',
        'description': 'Your SSH key is too short',
        'severity': 'high',
        'fixable': true,
        'fixed': false,
        'category': 'ssh',
        'detail': 'Use at least 4096-bit keys',
      };
      final issue = SecurityIssue.fromJson(json);
      expect(issue.id, 'issue-1');
      expect(issue.title, 'Weak SSH key');
      expect(issue.description, 'Your SSH key is too short');
      expect(issue.severity, 'high');
      expect(issue.fixable, isTrue);
      expect(issue.fixed, isFalse);
      expect(issue.category, 'ssh');
      expect(issue.detail, 'Use at least 4096-bit keys');
    });

    test('fromJson uses defaults for missing fields', () {
      final json = <String, dynamic>{};
      final issue = SecurityIssue.fromJson(json);
      expect(issue.id, '');
      expect(issue.title, '');
      expect(issue.description, '');
      expect(issue.severity, 'low');
      expect(issue.fixable, isFalse);
      expect(issue.fixed, isFalse);
      expect(issue.category, 'general');
      expect(issue.detail, isNull);
    });

    test('fromJson handles null detail', () {
      final json = {
        'id': 'issue-2',
        'title': 'Minor issue',
        'description': 'desc',
        'severity': 'medium',
        'fixable': false,
        'fixed': false,
        'category': 'general',
      };
      final issue = SecurityIssue.fromJson(json);
      expect(issue.detail, isNull);
    });

    test('copyWith changes fixed field', () {
      const original = SecurityIssue(
        id: 'issue-3',
        title: 'Test Issue',
        description: 'A test security issue',
        severity: 'medium',
        fixable: true,
        fixed: false,
        category: 'network',
      );
      final fixed = original.copyWith(fixed: true);
      expect(fixed.fixed, isTrue);
      expect(fixed.id, 'issue-3');
      expect(fixed.title, 'Test Issue');
      expect(fixed.severity, 'medium');
      expect(fixed.fixable, isTrue);
      expect(fixed.category, 'network');
      expect(fixed.detail, isNull);
    });

    test('copyWith preserves detail', () {
      const original = SecurityIssue(
        id: 'issue-4',
        title: 'With Detail',
        description: 'desc',
        severity: 'high',
        fixable: true,
        fixed: false,
        category: 'crypto',
        detail: 'Some detail text',
      );
      final copy = original.copyWith();
      expect(copy.detail, 'Some detail text');
    });

    test('copyWith without arguments returns same values', () {
      const original = SecurityIssue(
        id: 'issue-5',
        title: 'Open Port',
        description: 'Port 22 is open',
        severity: 'low',
        fixable: false,
        fixed: false,
        category: 'network',
      );
      final copy = original.copyWith();
      expect(copy.id, original.id);
      expect(copy.severity, original.severity);
      expect(copy.fixed, original.fixed);
    });
  });
}
