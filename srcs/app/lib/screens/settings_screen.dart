import 'package:flutter/material.dart';
import 'package:ohc_app/theme.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:ohc_app/services/auth_service.dart';
import 'package:ohc_app/services/settings_service.dart';
import 'package:ohc_app/services/local_manager_service.dart';

class SettingsScreen extends ConsumerWidget {
  const SettingsScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final user = ref.watch(authStateProvider).valueOrNull;
    final clientSettingsAsync = ref.watch(clientSettingsProvider);
    // Trigger lifecycle management
    ref.watch(standaloneManagerProvider);

    return Scaffold(
      appBar: AppBar(title: const Text('Settings')),
      body: clientSettingsAsync.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (err, _) => Center(child: Text('Error: $err')),
        data: (settings) => ListView(
          padding: const EdgeInsets.all(24),
          children: [
            if (user != null) ...[
              ListTile(
                leading: CircleAvatar(child: Text(user.name.substring(0, 1).toUpperCase())),
                title: Text(user.name),
                subtitle: Text(user.email),
              ),
              const Divider(),
            ],
            
            _SectionHeader(title: 'Communication'),
            ListTile(
              leading: const Icon(Icons.link),
              title: const Text('Backend URL'),
              subtitle: Text(settings.backendUrl),
              trailing: IconButton(
                icon: const Icon(Icons.edit),
                onPressed: () => _editBackendUrl(context, ref, settings.backendUrl),
              ),
            ),
            
            SwitchListTile(
              secondary: const Icon(Icons.computer),
              title: const Text('Standalone Mode'),
              subtitle: const Text('App manages local backend lifecycle'),
              value: settings.standaloneMode,
              onChanged: (value) => ref.read(clientSettingsProvider.notifier).updateStandaloneMode(value),
            ),

            if (settings.standaloneMode) ...[
              const Divider(),
              _SectionHeader(title: 'Local Backend'),
              _LocalBackendStatusGlassCard(),
            ],

            const Divider(),
            _SectionHeader(title: 'Account'),
            ListTile(
              leading: const Icon(Icons.business),
              title: const Text('Organization'),
              subtitle: Text(user?.organizationId ?? '—'),
            ),
            ListTile(
              leading: const Icon(Icons.verified_user),
              title: const Text('Role'),
              subtitle: Text(user?.role ?? '—'),
            ),
            const Divider(),
            ListTile(
              leading: const Icon(Icons.logout, color: Colors.red),
              title: const Text('Sign Out', style: TextStyle(color: Colors.red)),
              onTap: () => ref.read(authStateProvider.notifier).logout(),
            ),
          ],
        ),
      ),
    );
  }

  Future<void> _editBackendUrl(BuildContext context, WidgetRef ref, String current) async {
    final controller = TextEditingController(text: current);
    final result = await showDialog<String>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Edit Backend URL'),
        content: TextField(
          controller: controller,
          decoration: const InputDecoration(labelText: 'URL (e.g. http://localhost:8080)'),
        ),
        actions: [
          TextButton(onPressed: () => Navigator.pop(context), child: const Text('Cancel')),
          TextButton(
            onPressed: () => Navigator.pop(context, controller.text),
            child: const Text('Save'),
          ),
        ],
      ),
    );
    if (result != null && result.isNotEmpty) {
      ref.read(clientSettingsProvider.notifier).updateBackendUrl(result);
    }
  }
}

class _SectionHeader extends StatelessWidget {
  final String title;
  const _SectionHeader({required this.title});

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 8.0),
      child: Text(
        title.toUpperCase(),
        style: Theme.of(context).textTheme.labelLarge?.copyWith(
              color: Theme.of(context).colorScheme.primary,
              fontWeight: FontWeight.bold,
            ),
      ),
    );
  }
}

class _LocalBackendStatusCard extends ConsumerWidget {
  const _LocalBackendStatusGlassCard();

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final manager = ref.watch(localManagerServiceProvider);
    
    return FutureBuilder<bool>(
      future: manager.isServiceRunning(),
      builder: (context, snapshot) {
        final running = snapshot.data ?? false;
        return GlassCard(
          child: Padding(
            padding: const EdgeInsets.all(16.0),
            child: Column(
              children: [
                Row(
                  children: [
                    Icon(
                      running ? Icons.check_circle : Icons.error,
                      color: running ? Colors.green : Colors.red,
                    ),
                    const SizedBox(width: 8),
                    Text(running ? 'Service Running' : 'Service Stopped'),
                    const Spacer(),
                    ElevatedButton(
                      onPressed: () async {
                        if (running) {
                          await manager.stopService();
                        } else {
                          await manager.startService();
                        }
                        // Simple refresh hack
                        (context as Element).markNeedsBuild();
                      },
                      child: Text(running ? 'Stop' : 'Start'),
                    ),
                  ],
                ),
                const SizedBox(height: 8),
                OutlinedButton.icon(
                  onPressed: () async {
                    final report = await manager.runDoctor();
                    if (context.mounted) {
                      showDialog(
                        context: context,
                        builder: (context) => AlertDialog(
                          title: const Text('System Doctor'),
                          content: SingleChildScrollView(child: Text(report)),
                          actions: [TextButton(onPressed: () => Navigator.pop(context), child: const Text('Close'))],
                        ),
                      );
                    }
                  },
                  icon: const Icon(Icons.medical_services),
                  label: const Text('Run Doctor Diagnostics'),
                ),
              ],
            ),
          ),
        );
      },
    );
  }
}
