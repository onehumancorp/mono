import 'dart:async';
import 'dart:convert';

import 'package:centrifuge/centrifuge.dart' as centrifuge;
import 'package:flutter/foundation.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:ohc_app/services/auth_service.dart';

/// A single real-time chat message received over Centrifuge.
class CentrifugeMessage {
  final String id;
  final String channelId;
  final String authorId;
  final String authorName;
  final String body;
  final DateTime sentAt;

  const CentrifugeMessage({
    required this.id,
    required this.channelId,
    required this.authorId,
    required this.authorName,
    required this.body,
    required this.sentAt,
  });

  factory CentrifugeMessage.fromJson(Map<String, dynamic> json) {
    return CentrifugeMessage(
      id: json['id'] as String? ?? '',
      channelId: json['channel_id'] as String? ?? '',
      authorId: json['author_id'] as String? ?? '',
      authorName: json['author_name'] as String? ?? 'Unknown',
      body: json['body'] as String? ?? '',
      sentAt: json['sent_at'] != null
          ? DateTime.parse(json['sent_at'] as String)
          : DateTime.now(),
    );
  }

  Map<String, dynamic> toJson() => {
        'id': id,
        'channel_id': channelId,
        'author_id': authorId,
        'author_name': authorName,
        'body': body,
        'sent_at': sentAt.toIso8601String(),
      };
}

/// Manages a real-time chat connection to a Centrifuge server.
///
/// Centrifuge provides WebSocket-based pub/sub that allows agents and humans to
/// communicate without requiring a separate Chatwoot server.  Each chat room
/// maps to a Centrifuge channel prefixed with `chat:`.
class CentrifugeService {
  final String serverUrl;
  final String token;
  final String userId;
  final String userName;

  centrifuge.Client? _client;
  final Map<String, centrifuge.Subscription> _subscriptions = {};
  final Map<String, StreamController<CentrifugeMessage>> _controllers = {};

  /// Optional factory used to create the centrifuge [Client]. When omitted the
  /// default [centrifuge.createClient] function is used. Inject a custom
  /// factory in tests to avoid real network connections.
  final centrifuge.Client Function(String, centrifuge.ClientConfig)?
      clientFactory;

  CentrifugeService({
    required this.serverUrl,
    required this.token,
    required this.userId,
    required this.userName,
    this.clientFactory,
  });

  /// Connect to the Centrifuge server.
  Future<void> connect() async {
    final factory = clientFactory ?? centrifuge.createClient;
    _client = factory(
      serverUrl,
      centrifuge.ClientConfig(
        token: token,
      ),
    );
    await _client!.connect();
  }

  /// Subscribe to a chat room and return a stream of incoming messages.
  ///
  /// The Centrifuge channel name is `chat:<roomId>`.
  Stream<CentrifugeMessage> subscribe(String roomId) {
    final channel = 'chat:$roomId';
    if (_controllers.containsKey(channel)) {
      return _controllers[channel]!.stream;
    }

    final controller = StreamController<CentrifugeMessage>.broadcast();
    _controllers[channel] = controller;

    final sub = _client!.newSubscription(channel);
    sub.publication.listen((event) {
      try {
        final json = jsonDecode(utf8.decode(event.data)) as Map<String, dynamic>;
        controller.add(CentrifugeMessage.fromJson(json));
      } catch (e) {
        debugPrint('[CentrifugeService] Failed to parse message: $e');
      }
    });

    sub.subscribe();
    _subscriptions[channel] = sub;
    return controller.stream;
  }

  /// Publish a message to a chat room.
  Future<void> publish(String roomId, String body) async {
    final channel = 'chat:$roomId';
    final payload = jsonEncode({
      'author_id': userId,
      'author_name': userName,
      'body': body,
      'sent_at': DateTime.now().toUtc().toIso8601String(),
    });
    await _client?.publish(channel, utf8.encode(payload));
  }

  /// Unsubscribe from a chat room.
  Future<void> unsubscribe(String roomId) async {
    final channel = 'chat:$roomId';
    await _subscriptions[channel]?.unsubscribe();
    _subscriptions.remove(channel);
    await _controllers[channel]?.close();
    _controllers.remove(channel);
  }

  /// Disconnect from the server and release all resources.
  Future<void> disconnect() async {
    for (final sub in _subscriptions.values) {
      await sub.unsubscribe();
    }
    _subscriptions.clear();
    for (final ctrl in _controllers.values) {
      await ctrl.close();
    }
    _controllers.clear();
    await _client?.disconnect();
    _client = null;
  }
}

// ── Providers ──────────────────────────────────────────────────────────────

final centrifugeUrlProvider = Provider<String>(
  (_) => const String.fromEnvironment(
    'CENTRIFUGE_URL',
    defaultValue: 'ws://localhost:8000/connection/websocket',
  ),
);

final centrifugeServiceProvider = Provider<CentrifugeService?>((ref) {
  final user = ref.watch(authStateProvider).valueOrNull;
  if (user == null) return null;
  final url = ref.watch(centrifugeUrlProvider);
  return CentrifugeService(
    serverUrl: url,
    token: user.token,
    userId: user.id,
    userName: user.name,
  );
});
