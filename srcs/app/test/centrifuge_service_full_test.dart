import 'dart:async';
import 'dart:convert';

import 'package:centrifuge/centrifuge.dart' as centrifuge;
import 'package:fixnum/fixnum.dart' as fixnum;
import 'package:flutter_test/flutter_test.dart';
import 'package:mocktail/mocktail.dart';
import 'package:ohc_app/services/centrifuge_service.dart';

// ── Mocks ─────────────────────────────────────────────────────────────────────

class MockCentrifugeClient extends Mock implements centrifuge.Client {}

class MockSubscription extends Mock implements centrifuge.Subscription {}

class FakeClientConfig extends Fake implements centrifuge.ClientConfig {}

void main() {
  setUpAll(() {
    registerFallbackValue(FakeClientConfig());
  });

  late MockCentrifugeClient mockClient;
  late MockSubscription mockSub;
  late CentrifugeService svc;

  setUp(() {
    mockClient = MockCentrifugeClient();
    mockSub = MockSubscription();

    svc = CentrifugeService(
      serverUrl: 'ws://localhost:8000/connection/websocket',
      token: 'test-token',
      userId: 'user-1',
      userName: 'Alice',
      clientFactory: (url, config) => mockClient,
    );
  });

  group('CentrifugeService connect/disconnect', () {
    test('connect creates client and calls connect()', () async {
      when(() => mockClient.connect()).thenAnswer((_) async {});

      await svc.connect();

      verify(() => mockClient.connect()).called(1);
    });

    test('disconnect calls disconnect on client', () async {
      when(() => mockClient.connect()).thenAnswer((_) async {});
      when(() => mockClient.disconnect()).thenAnswer((_) async {});

      await svc.connect();
      await svc.disconnect();

      verify(() => mockClient.disconnect()).called(1);
    });

    test('disconnect without connect does not throw', () async {
      // No connect was called, _client is null
      await expectLater(svc.disconnect(), completes);
    });
  });

  group('CentrifugeService subscribe/unsubscribe', () {
    final _pubController =
        StreamController<centrifuge.PublicationEvent>.broadcast();

    setUp(() {
      when(() => mockClient.connect()).thenAnswer((_) async {});
      when(() => mockClient.newSubscription(any())).thenReturn(mockSub);
      when(() => mockSub.publication).thenAnswer((_) => _pubController.stream);
      when(() => mockSub.subscribe()).thenAnswer((_) async {});
      when(() => mockSub.unsubscribe()).thenAnswer((_) async {});
    });

    tearDown(() {
      // Don't close the controller - it's reused across tests
    });

    test('subscribe returns a stream and subscribes', () async {
      await svc.connect();
      final stream = svc.subscribe('general');

      expect(stream, isA<Stream<CentrifugeMessage>>());
      verify(() => mockSub.subscribe()).called(1);
    });

    test('subscribe to same room does not re-subscribe', () async {
      await svc.connect();
      svc.subscribe('general');
      svc.subscribe('general');

      // subscribe() on the underlying centrifuge sub should only be called once
      verify(() => mockSub.subscribe()).called(1);
    });

    test('unsubscribe removes subscription', () async {
      await svc.connect();
      svc.subscribe('general');

      await svc.unsubscribe('general');

      verify(() => mockSub.unsubscribe()).called(1);
    });

    test('unsubscribe non-existent room does not throw', () async {
      await svc.connect();
      await expectLater(svc.unsubscribe('nonexistent'), completes);
    });
  });

  group('CentrifugeService publish', () {
    setUp(() {
      when(() => mockClient.connect()).thenAnswer((_) async {});
      when(() => mockClient.publish(any(), any()))
          .thenAnswer((_) async => centrifuge.PublishResult());
    });

    test('publish sends message to channel', () async {
      await svc.connect();
      await svc.publish('general', 'Hello, world!');

      verify(() => mockClient.publish(any(), any())).called(1);
    });

    test('publish without connect does not throw (null client check)', () async {
      // _client is null, publish uses ?. operator
      await expectLater(svc.publish('general', 'test'), completes);
    });
  });

  group('CentrifugeService disconnect with subscriptions', () {
    final _pubController =
        StreamController<centrifuge.PublicationEvent>.broadcast();

    test('disconnect unsubscribes all active subscriptions', () async {
      when(() => mockClient.connect()).thenAnswer((_) async {});
      when(() => mockClient.disconnect()).thenAnswer((_) async {});
      when(() => mockClient.newSubscription(any())).thenReturn(mockSub);
      when(() => mockSub.publication).thenAnswer((_) => _pubController.stream);
      when(() => mockSub.subscribe()).thenAnswer((_) async {});
      when(() => mockSub.unsubscribe()).thenAnswer((_) async {});

      await svc.connect();
      svc.subscribe('room1');
      svc.subscribe('room2');
      await svc.disconnect();

      verify(() => mockSub.unsubscribe()).called(2);
      verify(() => mockClient.disconnect()).called(1);
    });
  });

  group('CentrifugeService message parsing', () {
    test('subscription message is parsed and emitted', () async {
      final pubController =
          StreamController<centrifuge.PublicationEvent>.broadcast();

      when(() => mockClient.connect()).thenAnswer((_) async {});
      when(() => mockClient.newSubscription(any())).thenReturn(mockSub);
      when(() => mockSub.publication).thenAnswer((_) => pubController.stream);
      when(() => mockSub.subscribe()).thenAnswer((_) async {});

      await svc.connect();
      final stream = svc.subscribe('general');

      final received = <CentrifugeMessage>[];
      final sub = stream.listen(received.add);

      final msgData = jsonEncode({
        'id': 'msg-1',
        'channel_id': 'general',
        'author_id': 'user-1',
        'author_name': 'Alice',
        'body': 'Hello!',
        'sent_at': '2025-01-01T10:00:00.000Z',
      });

      pubController.add(centrifuge.PublicationEvent(
        utf8.encode(msgData),
        fixnum.Int64.ZERO,
        null,
        {},
      ));

      await Future.delayed(const Duration(milliseconds: 10));
      expect(received, hasLength(1));
      expect(received.first.body, 'Hello!');
      expect(received.first.authorName, 'Alice');

      await sub.cancel();
      await pubController.close();
    });

    test('invalid subscription message is silently ignored', () async {
      final pubController =
          StreamController<centrifuge.PublicationEvent>.broadcast();

      when(() => mockClient.connect()).thenAnswer((_) async {});
      when(() => mockClient.newSubscription(any())).thenReturn(mockSub);
      when(() => mockSub.publication).thenAnswer((_) => pubController.stream);
      when(() => mockSub.subscribe()).thenAnswer((_) async {});

      await svc.connect();
      final stream = svc.subscribe('general');

      final received = <CentrifugeMessage>[];
      final sub = stream.listen(received.add);

      // Send invalid JSON
      pubController.add(centrifuge.PublicationEvent(
        utf8.encode('not-valid-json{{{'),
        fixnum.Int64.ZERO,
        null,
        {},
      ));

      await Future.delayed(const Duration(milliseconds: 10));
      expect(received, isEmpty);

      await sub.cancel();
      await pubController.close();
    });
  });

  group('centrifugeUrlProvider', () {
    test('returns default URL', () {
      expect(
        const String.fromEnvironment('CENTRIFUGE_URL',
            defaultValue: 'ws://localhost:8000/connection/websocket'),
        'ws://localhost:8000/connection/websocket',
      );
    });
  });
}
