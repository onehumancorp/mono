import 'package:flutter/material.dart';
import 'package:ohc_app/theme.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:ohc_app/services/api_service.dart';

/// Screen for managing external integrations and MCP tools.
class IntegrationsScreen extends ConsumerStatefulWidget {
  const IntegrationsScreen({super.key});

  @override
  ConsumerState<IntegrationsScreen> createState() => _IntegrationsScreenState();
}

class _IntegrationsScreenState extends ConsumerState<IntegrationsScreen> {
  late Future<List<Map<String, dynamic>>> _mcpToolsFuture;

  @override
  void initState() {
    super.initState();
    _refresh();
  }

  void _refresh() {
    setState(() {
      _mcpToolsFuture = ref.read(apiServiceProvider)!.listMCPTools();
    });
  }

  @override
  Widget build(BuildContext context) {
    final colors = Theme.of(context).colorScheme;

    return Scaffold(
      appBar: AppBar(
        title: const Text('Integrations & Tools'),
      ),
      body: ListView(
        padding: const EdgeInsets.all(24),
        children: [
          Text(
            'External Channels',
            style: Theme.of(context).textTheme.titleLarge?.copyWith(fontWeight: FontWeight.bold),
          ),
          const SizedBox(height: 16),
          Row(
            children: [
              Expanded(
                child: _IntegrationGlassCard(
                  title: 'Telegram',
                  subtitle: 'Connect your bot for mobile alerts',
                  icon: Icons.send,
                  color: Colors.blue,
                  onConnect: () => _showConnectionDialog('Telegram'),
                ),
              ),
              const SizedBox(width: 16),
              Expanded(
                child: _IntegrationGlassCard(
                  title: 'Discord',
                  subtitle: 'Stream agent logs to a channel',
                  icon: Icons.forum_outlined,
                  color: const Color(0xFF5865F2),
                  onConnect: () => _showConnectionDialog('Discord'),
                ),
              ),
            ],
          ),
          const SizedBox(height: 48),

          Text(
            'MCP Tool Gateway',
            style: Theme.of(context).textTheme.titleLarge?.copyWith(fontWeight: FontWeight.bold),
          ),
          const SizedBox(height: 8),
          const Text(
            'Manually invoke tools bridged via the Model Context Protocol.',
            style: TextStyle(color: Colors.grey),
          ),
          const SizedBox(height: 16),
          
          FutureBuilder<List<Map<String, dynamic>>>(
            future: _mcpToolsFuture,
            builder: (context, snapshot) {
              if (snapshot.connectionState == ConnectionState.waiting) {
                return const Center(child: CircularProgressIndicator());
              }

              if (snapshot.hasError) {
                return Center(child: Text('Error: ${snapshot.error}'));
              }

              final tools = snapshot.data ?? [];
              if (tools.isEmpty) {
                return GlassCard(
                  child: Padding(
                    padding: const EdgeInsets.all(32),
                    child: Center(
                      child: Column(
                        children: [
                          Icon(Icons.construction, size: 48, color: colors.onSurfaceVariant.withOpacity(0.3)),
                          const SizedBox(height: 16),
                          const Text('No MCP tools active'),
                          TextButton(onPressed: _refresh, child: const Text('Refresh')),
                        ],
                      ),
                    ),
                  ),
                );
              }

              return Column(
                children: tools.map((tool) => _MCPToolTile(tool: tool)).toList(),
              );
            },
          ),
        ],
      ),
    );
  }

  void _showConnectionDialog(String platform) {
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: Text('Connect to $platform'),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            TextField(
              decoration: InputDecoration(
                labelText: '$platform Bot Token',
                border: const OutlineInputBorder(),
              ),
            ),
            const SizedBox(height: 16),
            const TextField(
              decoration: InputDecoration(
                labelText: 'Channel ID / Chat ID',
                border: OutlineInputBorder(),
              ),
            ),
          ],
        ),
        actions: [
          TextButton(onPressed: () => Navigator.pop(context), child: const Text('Cancel')),
          FilledButton(onPressed: () => Navigator.pop(context), child: const Text('Save Integration')),
        ],
      ),
    );
  }
}

class _IntegrationCard extends StatelessWidget {
  final String title;
  final String subtitle;
  final IconData icon;
  final Color color;
  final VoidCallback onConnect;

  const _IntegrationGlassCard({
    required this.title,
    required this.subtitle,
    required this.icon,
    required this.color,
    required this.onConnect,
  });

  @override
  Widget build(BuildContext context) {
    final colors = Theme.of(context).colorScheme;

    return GlassCard(
      child: Padding(
        padding: const EdgeInsets.all(24),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Container(
                  padding: const EdgeInsets.all(8),
                  decoration: BoxDecoration(
                    color: color.withOpacity(0.1),
                    borderRadius: BorderRadius.circular(8),
                  ),
                  child: Icon(icon, color: color, size: 24),
                ),
                const Spacer(),
                const Text('Inactive', style: TextStyle(fontSize: 10, color: Colors.grey)),
              ],
            ),
            const SizedBox(height: 16),
            Text(title, style: const TextStyle(fontSize: 18, fontWeight: FontWeight.bold)),
            const SizedBox(height: 4),
            Text(
              subtitle,
              style: TextStyle(fontSize: 12, color: colors.onSurfaceVariant),
            ),
            const SizedBox(height: 24),
            SizedBox(
              width: double.infinity,
              child: OutlinedButton(
                onPressed: onConnect,
                child: const Text('Configure'),
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _MCPToolTile extends StatelessWidget {
  final Map<String, dynamic> tool;

  const _MCPToolTile({required this.tool});

  @override
  Widget build(BuildContext context) {
    final colors = Theme.of(context).colorScheme;
    final name = tool['name'] as String? ?? 'Unknown Tool';
    final description = tool['description'] as String? ?? '';

    return GlassCard(
      margin: const EdgeInsets.only(bottom: 12),
      child: ListTile(
        leading: const Icon(Icons.build_circle_outlined),
        title: Text(name),
        subtitle: Text(description, maxLines: 1, overflow: TextOverflow.ellipsis),
        trailing: OutlinedButton(
          onPressed: () {}, // Invoke dialog
          child: const Text('Invoke'),
        ),
        onTap: () {},
      ),
    );
  }
}
