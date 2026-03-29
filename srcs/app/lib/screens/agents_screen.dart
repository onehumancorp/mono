import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
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
            onPressed: () => context.go('/agents/hire'),
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
            ? _EmptyAgents(onHire: () => context.go('/agents/hire'))
            : _AgentList(agents: agents),
      ),
    );
  }
}

class _EmptyAgents extends StatelessWidget {
  final VoidCallback onHire;
  const _EmptyAgents({required this.onHire});

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
          const SizedBox(height: 24),
          FilledButton.icon(
            onPressed: onHire,
            icon: const Icon(Icons.add),
            label: const Text('Hire New Agent'),
          ),
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
