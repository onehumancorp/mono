import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:ohc_app/models/agent.dart';
import 'package:ohc_app/services/api_service.dart';

final _agentsProvider = FutureProvider<List<Agent>>((ref) async {
  final api = ref.watch(apiServiceProvider);
  if (api == null) return [];
  return api.listAgents();
});

class AgentsScreen extends ConsumerWidget {
  const AgentsScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final snapshot = ref.watch(_agentsProvider);
    return Scaffold(
      appBar: AppBar(
        title: const Text('Agents'),
        actions: [
          FilledButton.icon(
            onPressed: () => _showHireDialog(context, ref),
            icon: const Icon(Icons.add),
            label: const Text('Hire Agent'),
          ),
          const SizedBox(width: 16),
        ],
      ),
      body: snapshot.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => Center(child: Text('Error: $e')),
        data: (agents) => agents.isEmpty
            ? const _EmptyAgents()
            : _AgentList(agents: agents),
      ),
    );
  }

  void _showHireDialog(BuildContext context, WidgetRef ref) {
    showDialog(
      context: context,
      builder: (_) => _HireAgentDialog(ref: ref),
    );
  }
}

class _EmptyAgents extends StatelessWidget {
  const _EmptyAgents();

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          const Icon(Icons.smart_toy, size: 64, color: Colors.grey),
          const SizedBox(height: 16),
          Text(
            'No agents yet',
            style: Theme.of(context).textTheme.titleLarge,
          ),
          const SizedBox(height: 8),
          const Text('Hire your first AI agent to get started.'),
        ],
      ),
    );
  }
}

class _AgentList extends StatelessWidget {
  final List<Agent> agents;
  const _AgentList({required this.agents});

  @override
  Widget build(BuildContext context) {
    return ListView.builder(
      padding: const EdgeInsets.all(16),
      itemCount: agents.length,
      itemBuilder: (_, i) => _AgentCard(agent: agents[i]),
    );
  }
}

class _AgentCard extends StatelessWidget {
  final Agent agent;
  const _AgentCard({required this.agent});

  @override
  Widget build(BuildContext context) {
    return Card(
      margin: const EdgeInsets.only(bottom: 12),
      child: ListTile(
        leading: CircleAvatar(
          backgroundColor: agent.isRunning ? Colors.green : Colors.grey.shade300,
          child: const Icon(Icons.smart_toy, color: Colors.white),
        ),
        title: Text(agent.name, style: const TextStyle(fontWeight: FontWeight.w600)),
        subtitle: Text(agent.role),
        trailing: Chip(
          label: Text(agent.status),
          backgroundColor:
              agent.isRunning ? Colors.green.shade100 : Colors.grey.shade200,
        ),
      ),
    );
  }
}

class _HireAgentDialog extends StatefulWidget {
  final WidgetRef ref;
  const _HireAgentDialog({required this.ref});

  @override
  State<_HireAgentDialog> createState() => _HireAgentDialogState();
}

class _HireAgentDialogState extends State<_HireAgentDialog> {
  final _nameCtrl = TextEditingController();
  final _roleCtrl = TextEditingController();
  bool _loading = false;

  @override
  void dispose() {
    _nameCtrl.dispose();
    _roleCtrl.dispose();
    super.dispose();
  }

  Future<void> _submit() async {
    if (_nameCtrl.text.isEmpty) return;
    setState(() => _loading = true);
    try {
      final api = widget.ref.read(apiServiceProvider);
      await api?.hireAgent(_nameCtrl.text.trim(), _roleCtrl.text.trim());
      widget.ref.invalidate(_agentsProvider);
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
      title: const Text('Hire Agent'),
      content: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          TextField(
            controller: _nameCtrl,
            decoration: const InputDecoration(labelText: 'Agent Name', border: OutlineInputBorder()),
          ),
          const SizedBox(height: 16),
          TextField(
            controller: _roleCtrl,
            decoration: const InputDecoration(labelText: 'Role', border: OutlineInputBorder()),
          ),
        ],
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.pop(context),
          child: const Text('Cancel'),
        ),
        FilledButton(
          onPressed: _loading ? null : _submit,
          child: _loading
              ? const SizedBox(height: 18, width: 18, child: CircularProgressIndicator(strokeWidth: 2))
              : const Text('Hire'),
        ),
      ],
    );
  }
}
