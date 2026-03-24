/// Chat channel model — mirrors the Rust core ChatChannel / ChatBackend.
class ChatChannel {
  final String id;
  final String organizationId;
  final String name;
  final ChatBackend backend;
  final Map<String, String> config;
  final bool enabled;
  final DateTime createdAt;

  const ChatChannel({
    required this.id,
    required this.organizationId,
    required this.name,
    required this.backend,
    required this.config,
    required this.enabled,
    required this.createdAt,
  });

  factory ChatChannel.fromJson(Map<String, dynamic> json) {
    return ChatChannel(
      id: json['id'] as String,
      organizationId: json['organization_id'] as String? ?? '',
      name: json['name'] as String,
      backend: ChatBackend.fromJson(json['backend']),
      config: (json['config'] as Map<String, dynamic>?)
              ?.cast<String, String>() ??
          {},
      enabled: json['enabled'] as bool? ?? true,
      createdAt: json['created_at'] != null
          ? DateTime.parse(json['created_at'] as String)
          : DateTime.now(),
    );
  }

  Map<String, dynamic> toJson() => {
        'id': id,
        'organization_id': organizationId,
        'name': name,
        'backend': backend.toJson(),
        'config': config,
        'enabled': enabled,
        'created_at': createdAt.toIso8601String(),
      };
}

/// Supported chat backends — mirrors the Rust ChatBackend enum.
enum ChatBackendType {
  chatwoot,
  slack,
  telegram,
  discord,
  teams,
  mattermost,
  centrifuge,
  webhook,
}

class ChatBackend {
  final ChatBackendType type;

  /// URL used by centrifuge and webhook backends.
  final String? url;

  const ChatBackend({required this.type, this.url});

  factory ChatBackend.fromJson(dynamic json) {
    if (json is String) {
      return ChatBackend(type: _parseType(json));
    }
    if (json is Map<String, dynamic>) {
      if (json.containsKey('centrifuge')) {
        return ChatBackend(
          type: ChatBackendType.centrifuge,
          url: json['centrifuge'] as String?,
        );
      }
      if (json.containsKey('webhook')) {
        return ChatBackend(
          type: ChatBackendType.webhook,
          url: json['webhook'] as String?,
        );
      }
    }
    return const ChatBackend(type: ChatBackendType.webhook);
  }

  Map<String, dynamic> toJson() {
    switch (type) {
      case ChatBackendType.centrifuge:
        return {'centrifuge': url ?? ''};
      case ChatBackendType.webhook:
        return {'webhook': url ?? ''};
      default:
        return {'type': type.name};
    }
  }

  static ChatBackendType _parseType(String s) {
    switch (s) {
      case 'chatwoot':
        return ChatBackendType.chatwoot;
      case 'slack':
        return ChatBackendType.slack;
      case 'telegram':
        return ChatBackendType.telegram;
      case 'discord':
        return ChatBackendType.discord;
      case 'teams':
        return ChatBackendType.teams;
      case 'mattermost':
        return ChatBackendType.mattermost;
      case 'centrifuge':
        return ChatBackendType.centrifuge;
      default:
        return ChatBackendType.webhook;
    }
  }

  String get displayName {
    switch (type) {
      case ChatBackendType.chatwoot:
        return 'Chatwoot';
      case ChatBackendType.slack:
        return 'Slack';
      case ChatBackendType.telegram:
        return 'Telegram';
      case ChatBackendType.discord:
        return 'Discord';
      case ChatBackendType.teams:
        return 'Microsoft Teams';
      case ChatBackendType.mattermost:
        return 'Mattermost';
      case ChatBackendType.centrifuge:
        return 'Native (Centrifuge)';
      case ChatBackendType.webhook:
        return 'Webhook';
    }
  }
}
