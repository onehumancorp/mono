import 'dart:convert';

import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:mocktail/mocktail.dart';
import 'package:ohc_app/services/auth_service.dart';

class MockHttpClient extends Mock implements http.Client {}

class FakeUri extends Fake implements Uri {}

void main() {
  setUpAll(() {
    registerFallbackValue(FakeUri());
  });

  // ── AuthUser ──────────────────────────────────────────────────────────────

  group('AuthUser', () {
    test('fromJson parses all fields', () {
      final json = {
        'id': 'user-1',
        'email': 'alice@example.com',
        'name': 'Alice Smith',
        'role': 'admin',
        'organization_id': 'org-1',
      };
      final user = AuthUser.fromJson(json, 'tok-abc');
      expect(user.id, 'user-1');
      expect(user.email, 'alice@example.com');
      expect(user.name, 'Alice Smith');
      expect(user.role, 'admin');
      expect(user.organizationId, 'org-1');
      expect(user.token, 'tok-abc');
    });

    test('fromJson uses email as name when name is missing', () {
      final json = {
        'id': 'user-2',
        'email': 'bob@example.com',
      };
      final user = AuthUser.fromJson(json, 'tok-xyz');
      expect(user.name, 'bob@example.com');
    });

    test('fromJson uses defaults for missing optional fields', () {
      final json = {
        'id': 'user-3',
        'email': 'carol@example.com',
      };
      final user = AuthUser.fromJson(json, 'tok-def');
      expect(user.role, 'viewer');
      expect(user.organizationId, '');
    });
  });

  // ── AuthService ───────────────────────────────────────────────────────────

  group('AuthService', () {
    late MockHttpClient mockClient;
    late AuthService service;

    setUp(() {
      mockClient = MockHttpClient();
      service = AuthService(
        baseUrl: 'http://localhost:8080',
        client: mockClient,
      );
    });

    test('login returns AuthUser on HTTP 200', () async {
      final responseBody = jsonEncode({
        'token': 'jwt-token-123',
        'user': {
          'id': 'u1',
          'email': 'alice@example.com',
          'name': 'Alice',
          'role': 'admin',
          'organization_id': 'org-a',
        },
      });

      when(() => mockClient.post(
            any(),
            headers: any(named: 'headers'),
            body: any(named: 'body'),
          )).thenAnswer((_) async =>
          http.Response(responseBody, 200));

      final user = await service.login('alice@example.com', 'password');
      expect(user.id, 'u1');
      expect(user.email, 'alice@example.com');
      expect(user.token, 'jwt-token-123');
      expect(user.role, 'admin');
    });

    test('login throws on non-200 response', () async {
      when(() => mockClient.post(
            any(),
            headers: any(named: 'headers'),
            body: any(named: 'body'),
          )).thenAnswer((_) async => http.Response('Unauthorized', 401));

      expect(
        () => service.login('bad@example.com', 'wrong'),
        throwsA(isA<Exception>()),
      );
    });

    test('logout calls the logout endpoint', () async {
      when(() => mockClient.post(
            any(),
            headers: any(named: 'headers'),
          )).thenAnswer((_) async => http.Response('', 200));

      await service.logout('some-token');

      verify(() => mockClient.post(
            any(),
            headers: any(named: 'headers'),
          )).called(1);
    });
  });
}
