import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:ohc_app/services/api_service.dart';

final _meetingsProvider = FutureProvider<List<Map<String, dynamic>>>((ref) async {
  final api = ref.watch(apiServiceProvider);
  if (api == null) return [];
  return api.listMeetings();
});

class MeetingsScreen extends ConsumerWidget {
  const MeetingsScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final snapshot = ref.watch(_meetingsProvider);
    return Scaffold(
      appBar: AppBar(
        title: const Text('Meeting Rooms'),
        actions: [
          FilledButton.icon(
            icon: const Icon(Icons.add),
            label: const Text('New Room'),
            onPressed: () => _showCreateDialog(context, ref),
          ),
          const SizedBox(width: 16),
        ],
      ),
      body: snapshot.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => Center(child: Text('Error: $e')),
        data: (rooms) => rooms.isEmpty
            ? _EmptyRooms(onCreate: () => _showCreateDialog(context, ref))
            : _RoomList(rooms: rooms, ref: ref),
      ),
    );
  }

  void _showCreateDialog(BuildContext context, WidgetRef ref) {
    showDialog(
      context: context,
      builder: (_) => _CreateRoomDialog(ref: ref),
    );
  }
}

// ── Empty state ────────────────────────────────────────────────────────────

class _EmptyRooms extends StatelessWidget {
  final VoidCallback onCreate;
  const _EmptyRooms({required this.onCreate});

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          const Icon(Icons.video_call, size: 64, color: Colors.grey),
          const SizedBox(height: 16),
          Text('No active meeting rooms.',
              style: Theme.of(context).textTheme.titleLarge),
          const SizedBox(height: 8),
          const Text('Create a meeting room to collaborate with your team.'),
          const SizedBox(height: 24),
          FilledButton.icon(
            icon: const Icon(Icons.add),
            label: const Text('Create Room'),
            onPressed: onCreate,
          ),
        ],
      ),
    );
  }
}

// ── Room list ──────────────────────────────────────────────────────────────

class _RoomList extends StatelessWidget {
  final List<Map<String, dynamic>> rooms;
  final WidgetRef ref;

  const _RoomList({required this.rooms, required this.ref});

  @override
  Widget build(BuildContext context) {
    return ListView.builder(
      padding: const EdgeInsets.all(16),
      itemCount: rooms.length,
      itemBuilder: (_, i) => _RoomCard(room: rooms[i], ref: ref),
    );
  }
}

class _RoomCard extends StatefulWidget {
  final Map<String, dynamic> room;
  final WidgetRef ref;

  const _RoomCard({required this.room, required this.ref});

  @override
  State<_RoomCard> createState() => _RoomCardState();
}

class _RoomCardState extends State<_RoomCard> {
  bool _joining = false;

  Color _statusColor() {
    switch (widget.room['status'] as String? ?? '') {
      case 'active':
        return Colors.green;
      case 'scheduled':
        return Colors.blue;
      case 'ended':
        return Colors.grey;
      default:
        return Colors.orange;
    }
  }

  Future<void> _join() async {
    setState(() => _joining = true);
    try {
      final api = widget.ref.read(apiServiceProvider);
      final info = await api?.joinMeeting(widget.room['id'] as String);
      if (mounted && info != null) {
        _showJoinInfo(context, info);
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context)
            .showSnackBar(SnackBar(content: Text('Join failed: $e')));
      }
    } finally {
      if (mounted) setState(() => _joining = false);
    }
  }

  void _showJoinInfo(BuildContext context, Map<String, dynamic> info) {
    showDialog(
      context: context,
      builder: (_) => AlertDialog(
        title: const Text('Join Meeting'),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            if (info['join_url'] != null) ...[
              const Text('Join URL:',
                  style: TextStyle(fontWeight: FontWeight.bold)),
              const SizedBox(height: 4),
              SelectableText(info['join_url'] as String),
            ],
            if (info['token'] != null) ...[
              const SizedBox(height: 12),
              const Text('Token:', style: TextStyle(fontWeight: FontWeight.bold)),
              const SizedBox(height: 4),
              SelectableText(info['token'] as String,
                  style: const TextStyle(fontFamily: 'monospace', fontSize: 12)),
            ],
          ],
        ),
        actions: [
          FilledButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('Done'),
          ),
        ],
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final room = widget.room;
    final participantCount = room['participant_count'] as int? ?? 0;
    return Card(
      margin: const EdgeInsets.only(bottom: 12),
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Row(
          children: [
            const Icon(Icons.video_call, color: Colors.teal, size: 36),
            const SizedBox(width: 12),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    room['name'] as String? ?? 'Meeting Room',
                    style: const TextStyle(
                        fontWeight: FontWeight.bold, fontSize: 15),
                  ),
                  const SizedBox(height: 4),
                  Row(
                    children: [
                      Container(
                        width: 8,
                        height: 8,
                        margin: const EdgeInsets.only(right: 4),
                        decoration: BoxDecoration(
                          color: _statusColor(),
                          shape: BoxShape.circle,
                        ),
                      ),
                      Text(
                        room['status'] as String? ?? '',
                        style: TextStyle(
                            color: _statusColor(), fontSize: 12),
                      ),
                      const SizedBox(width: 12),
                      if (participantCount > 0) ...[
                        const Icon(Icons.people, size: 14,
                            color: Colors.grey),
                        const SizedBox(width: 4),
                        Text('$participantCount',
                            style: Theme.of(context).textTheme.bodySmall),
                      ],
                    ],
                  ),
                ],
              ),
            ),
            FilledButton.icon(
              icon: _joining
                  ? const SizedBox(
                      width: 16,
                      height: 16,
                      child: CircularProgressIndicator(strokeWidth: 2),
                    )
                  : const Icon(Icons.login, size: 18),
              label: const Text('Join'),
              onPressed: _joining ? null : _join,
            ),
          ],
        ),
      ),
    );
  }
}

// ── Create room dialog ─────────────────────────────────────────────────────

class _CreateRoomDialog extends StatefulWidget {
  final WidgetRef ref;
  const _CreateRoomDialog({required this.ref});

  @override
  State<_CreateRoomDialog> createState() => _CreateRoomDialogState();
}

class _CreateRoomDialogState extends State<_CreateRoomDialog> {
  final _nameCtrl = TextEditingController();
  bool _loading = false;

  @override
  void dispose() {
    _nameCtrl.dispose();
    super.dispose();
  }

  Future<void> _submit() async {
    if (_nameCtrl.text.trim().isEmpty) return;
    setState(() => _loading = true);
    try {
      final api = widget.ref.read(apiServiceProvider);
      await api?.createMeeting(_nameCtrl.text.trim());
      widget.ref.invalidate(_meetingsProvider);
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
      title: const Text('Create Meeting Room'),
      content: TextField(
        controller: _nameCtrl,
        autofocus: true,
        decoration: const InputDecoration(
          labelText: 'Room Name',
          hintText: 'e.g. Stand-up, Sprint Review',
          border: OutlineInputBorder(),
        ),
        onSubmitted: (_) => _submit(),
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
              : const Text('Create'),
        ),
      ],
    );
  }
}
