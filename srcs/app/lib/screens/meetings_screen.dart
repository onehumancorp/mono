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
      appBar: AppBar(title: const Text('Meeting Rooms')),
      body: snapshot.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => Center(child: Text('Error: $e')),
        data: (rooms) => rooms.isEmpty
            ? const Center(child: Text('No active meeting rooms.'))
            : ListView.builder(
                padding: const EdgeInsets.all(16),
                itemCount: rooms.length,
                itemBuilder: (_, i) {
                  final room = rooms[i];
                  return Card(
                    margin: const EdgeInsets.only(bottom: 12),
                    child: ListTile(
                      leading: const Icon(Icons.video_call, color: Colors.teal),
                      title: Text(room['name'] as String? ?? 'Meeting Room'),
                      subtitle: Text(room['status'] as String? ?? ''),
                      trailing: FilledButton.icon(
                        icon: const Icon(Icons.login, size: 18),
                        label: const Text('Join'),
                        onPressed: () {},
                      ),
                    ),
                  );
                },
              ),
      ),
    );
  }
}
