import 'package:flutter/material.dart';
import 'package:ohc_app/theme.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:ohc_app/models/channel.dart';
import 'package:ohc_app/services/api_service.dart';

final _channelsProvider =
    FutureProvider<List<ChatChannel>>((ref) async {
  final api = ref.watch(apiServiceProvider);
  if (api == null) return [];
  return api.listChannels();
});

// ── Definitions for each supported backend ─────────────────────────────────

class _FieldDef {
  final String key;
  final String label;
  final bool secret;
  final String? hint;
  final List<String>? options;

  const _FieldDef({
    required this.key,
    required this.label,
    this.secret = false,
    this.hint,
    this.options,
  });
}

class _ChannelDef {
  final ChatBackendType type;
  final String icon;
  final String description;
  final String guideUrl;
  final List<_FieldDef> fields;

  const _ChannelDef({
    required this.type,
    required this.icon,
    required this.description,
    required this.guideUrl,
    required this.fields,
  });
}

const _channelDefs = <_ChannelDef>[
  _ChannelDef(
    type: ChatBackendType.centrifuge,
    icon: '⚡',
    description:
        'Native real-time chat using Centrifuge WebSocket — no external server required.',
    guideUrl: 'https://github.com/centrifugal/centrifuge',
    fields: [
      _FieldDef(
        key: 'url',
        label: 'Server URL',
        hint: 'ws://localhost:8000/connection/websocket',
      ),
    ],
  ),
  _ChannelDef(
    type: ChatBackendType.telegram,
    icon: '✈️',
    description: 'Connect via Telegram Bot API.',
    guideUrl: 'https://core.telegram.org/bots/tutorial',
    fields: [
      _FieldDef(
        key: 'bot_token',
        label: 'Bot Token',
        secret: true,
        hint: '123456:ABC-DEF1234...',
      ),
      _FieldDef(key: 'allowed_chats', label: 'Allowed Chat IDs', hint: '-100123456,987654'),
    ],
  ),
  _ChannelDef(
    type: ChatBackendType.discord,
    icon: '🎮',
    description: 'Connect via Discord Bot.',
    guideUrl: 'https://discord.com/developers/docs/intro',
    fields: [
      _FieldDef(key: 'bot_token', label: 'Bot Token', secret: true, hint: 'MTxxxxxxx.xxx.xxx'),
      _FieldDef(key: 'server_id', label: 'Server ID'),
      _FieldDef(key: 'channel_id', label: 'Default Channel ID'),
    ],
  ),
  _ChannelDef(
    type: ChatBackendType.slack,
    icon: '💬',
    description: 'Connect via Slack Bot.',
    guideUrl: 'https://api.slack.com/apps',
    fields: [
      _FieldDef(key: 'bot_token', label: 'Bot OAuth Token', secret: true, hint: 'xoxb-...'),
      _FieldDef(key: 'app_token', label: 'App-Level Token', secret: true, hint: 'xapp-...'),
      _FieldDef(key: 'channel', label: 'Default Channel', hint: '#general'),
    ],
  ),
  _ChannelDef(
    type: ChatBackendType.chatwoot,
    icon: '🗨️',
    description: 'Connect to a self-hosted or cloud Chatwoot instance.',
    guideUrl: 'https://www.chatwoot.com/docs',
    fields: [
      _FieldDef(key: 'api_url', label: 'Chatwoot URL', hint: 'https://app.chatwoot.com'),
      _FieldDef(key: 'api_key', label: 'API Key', secret: true),
      _FieldDef(key: 'account_id', label: 'Account ID'),
      _FieldDef(key: 'inbox_id', label: 'Inbox ID'),
    ],
  ),
  _ChannelDef(
    type: ChatBackendType.teams,
    icon: '🏢',
    description: 'Connect via Microsoft Teams Bot Framework.',
    guideUrl: 'https://learn.microsoft.com/en-us/microsoftteams/platform/',
    fields: [
      _FieldDef(key: 'app_id', label: 'App ID'),
      _FieldDef(key: 'app_password', label: 'App Password', secret: true),
    ],
  ),
  _ChannelDef(
    type: ChatBackendType.mattermost,
    icon: '🔵',
    description: 'Connect via Mattermost Bot.',
    guideUrl: 'https://developers.mattermost.com/integrate/reference/bot-accounts/',
    fields: [
      _FieldDef(key: 'server_url', label: 'Mattermost URL', hint: 'https://mattermost.example.com'),
      _FieldDef(key: 'bot_token', label: 'Bot Token', secret: true),
      _FieldDef(key: 'team', label: 'Team Name'),
      _FieldDef(key: 'channel', label: 'Default Channel'),
    ],
  ),
  _ChannelDef(
    type: ChatBackendType.webhook,
    icon: '🔗',
    description: 'Send messages to any endpoint via HTTP POST.',
    guideUrl: '',
    fields: [
      _FieldDef(key: 'url', label: 'Webhook URL'),
      _FieldDef(key: 'secret', label: 'Signing Secret', secret: true),
    ],
  ),
];

// ── Screen ─────────────────────────────────────────────────────────────────

class ChannelsScreen extends ConsumerWidget {
  const ChannelsScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final snapshot = ref.watch(_channelsProvider);
    return Scaffold(
      appBar: AppBar(
        title: const Text('Chat Channels'),
        actions: [
          FilledButton.icon(
            icon: const Icon(Icons.add),
            label: const Text('Add Channel'),
            onPressed: () => _showAddDialog(context, ref),
          ),
          const SizedBox(width: 16),
        ],
      ),
      body: snapshot.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => Center(child: Text('Error: $e')),
        data: (channels) => channels.isEmpty
            ? _EmptyChannels(onAdd: () => _showAddDialog(context, ref))
            : _ChannelList(channels: channels),
      ),
    );
  }

  void _showAddDialog(BuildContext context, WidgetRef ref) {
    showDialog(
      context: context,
      builder: (_) => _AddChannelDialog(ref: ref),
    );
  }
}

// ── Empty state ────────────────────────────────────────────────────────────

class _EmptyChannels extends StatelessWidget {
  final VoidCallback onAdd;
  const _EmptyChannels({required this.onAdd});

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          const Icon(Icons.chat_bubble_outline, size: 64, color: Colors.grey),
          const SizedBox(height: 16),
          Text('No channels yet',
              style: Theme.of(context).textTheme.titleLarge),
          const SizedBox(height: 8),
          const Text('Add a channel to start receiving messages.'),
          const SizedBox(height: 24),
          FilledButton.icon(
            icon: const Icon(Icons.add),
            label: const Text('Add Channel'),
            onPressed: onAdd,
          ),
        ],
      ),
    );
  }
}

// ── Channel list ───────────────────────────────────────────────────────────

class _ChannelList extends StatelessWidget {
  final List<ChatChannel> channels;
  const _ChannelList({required this.channels});

  @override
  Widget build(BuildContext context) {
    return ListView.builder(
      padding: const EdgeInsets.all(16),
      itemCount: channels.length,
      itemBuilder: (_, i) => _ChannelGlassCard(channel: channels[i]),
    );
  }
}

class _ChannelCard extends StatelessWidget {
  final ChatChannel channel;
  const _ChannelGlassCard({required this.channel});

  String _icon() {
    for (final def in _channelDefs) {
      if (def.type == channel.backend.type) return def.icon;
    }
    return '💬';
  }

  @override
  Widget build(BuildContext context) {
    return GlassCard(
      margin: const EdgeInsets.only(bottom: 12),
      child: ListTile(
        leading: Text(_icon(), style: const TextStyle(fontSize: 28)),
        title: Text(channel.name,
            style: const TextStyle(fontWeight: FontWeight.w600)),
        subtitle: Text(channel.backend.displayName),
        trailing: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            Chip(
              label: Text(channel.enabled ? 'Enabled' : 'Disabled'),
              backgroundColor: channel.enabled
                  ? Colors.green.shade100
                  : Colors.grey.shade200,
            ),
          ],
        ),
      ),
    );
  }
}

// ── Add channel dialog ─────────────────────────────────────────────────────

class _AddChannelDialog extends StatefulWidget {
  final WidgetRef ref;
  const _AddChannelDialog({required this.ref});

  @override
  State<_AddChannelDialog> createState() => _AddChannelDialogState();
}

class _AddChannelDialogState extends State<_AddChannelDialog> {
  _ChannelDef _selected = _channelDefs.first;
  final _nameCtrl = TextEditingController();
  final Map<String, TextEditingController> _fieldCtrls = {};
  bool _loading = false;

  @override
  void initState() {
    super.initState();
    _initFields();
  }

  void _initFields() {
    for (final c in _fieldCtrls.values) {
      c.dispose();
    }
    _fieldCtrls.clear();
    for (final f in _selected.fields) {
      _fieldCtrls[f.key] = TextEditingController();
    }
  }

  @override
  void dispose() {
    _nameCtrl.dispose();
    for (final c in _fieldCtrls.values) {
      c.dispose();
    }
    super.dispose();
  }

  Future<void> _submit() async {
    if (_nameCtrl.text.trim().isEmpty) return;
    setState(() => _loading = true);
    try {
      final config = {
        for (final e in _fieldCtrls.entries)
          if (e.value.text.isNotEmpty) e.key: e.value.text.trim(),
      };
      final api = widget.ref.read(apiServiceProvider);
      await api?.addChannel(
        name: _nameCtrl.text.trim(),
        backend: _selected.type.name,
        config: config,
      );
      widget.ref.invalidate(_channelsProvider);
      if (mounted) Navigator.pop(context);
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context)
            .showSnackBar(SnackBar(content: Text('Error: $e')));
      }
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return AlertDialog(
      title: const Text('Add Chat Channel'),
      content: SizedBox(
        width: 480,
        child: SingleChildScrollView(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              // Backend selector
              DropdownButtonFormField<_ChannelDef>(
                value: _selected,
                decoration: const InputDecoration(
                  labelText: 'Backend',
                  border: OutlineInputBorder(),
                ),
                items: _channelDefs
                    .map((d) => DropdownMenuItem(
                          value: d,
                          child: Row(
                            children: [
                              Text(d.icon),
                              const SizedBox(width: 8),
                              Text(d.type.name == 'centrifuge'
                                  ? 'Native (Centrifuge)'
                                  : d.type.name[0].toUpperCase() +
                                      d.type.name.substring(1)),
                            ],
                          ),
                        ))
                    .toList(),
                onChanged: (val) {
                  if (val == null) return;
                  setState(() {
                    _selected = val;
                    _initFields();
                  });
                },
              ),
              const SizedBox(height: 8),
              Text(
                _selected.description,
                style: Theme.of(context).textTheme.bodySmall,
              ),
              const SizedBox(height: 16),
              // Channel name
              TextField(
                controller: _nameCtrl,
                decoration: const InputDecoration(
                  labelText: 'Channel Name',
                  border: OutlineInputBorder(),
                ),
              ),
              const SizedBox(height: 12),
              // Dynamic fields
              ..._selected.fields.map((f) => Padding(
                    padding: const EdgeInsets.only(bottom: 12),
                    child: TextField(
                      controller: _fieldCtrls[f.key],
                      obscureText: f.secret,
                      decoration: InputDecoration(
                        labelText: f.label,
                        hintText: f.hint,
                        border: const OutlineInputBorder(),
                        suffixIcon:
                            f.secret ? const Icon(Icons.lock_outline) : null,
                      ),
                    ),
                  )),
            ],
          ),
        ),
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.pop(context),
          child: const Text('Cancel'),
        ),
        FilledButton(
          onPressed: _loading ? null : _submit,
          child: _loading
              ? const SizedBox(
                  height: 18,
                  width: 18,
                  child: CircularProgressIndicator(strokeWidth: 2),
                )
              : const Text('Add'),
        ),
      ],
    );
  }
}
