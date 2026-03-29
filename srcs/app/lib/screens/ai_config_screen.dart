import 'package:flutter/material.dart';
import 'package:ohc_app/theme.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:ohc_app/models/ai_provider.dart';
import 'package:ohc_app/services/api_service.dart';

final _providersProvider = FutureProvider<List<AiProvider>>((ref) async {
  final api = ref.watch(apiServiceProvider);
  if (api == null) return [];
  return api.listAiProviders();
});

// ── Screen ─────────────────────────────────────────────────────────────────

class AiConfigScreen extends ConsumerWidget {
  const AiConfigScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final snapshot = ref.watch(_providersProvider);
    return Scaffold(
      appBar: AppBar(
        title: const Text('AI Providers'),
        actions: [
          FilledButton.icon(
            icon: const Icon(Icons.add),
            label: const Text('Add Provider'),
            onPressed: () => _showAddDialog(context, ref),
          ),
          const SizedBox(width: 16),
        ],
      ),
      body: snapshot.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => Center(child: Text('Error: $e')),
        data: (providers) => providers.isEmpty
            ? _EmptyProviders(onAdd: () => _showAddDialog(context, ref))
            : _ProviderList(providers: providers, ref: ref),
      ),
    );
  }

  void _showAddDialog(BuildContext context, WidgetRef ref) {
    showDialog(
      context: context,
      builder: (_) => _ProviderDialog(ref: ref),
    );
  }
}

// ── Empty state ────────────────────────────────────────────────────────────

class _EmptyProviders extends StatelessWidget {
  final VoidCallback onAdd;
  const _EmptyProviders({required this.onAdd});

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          const Icon(Icons.psychology, size: 64, color: Colors.grey),
          const SizedBox(height: 16),
          Text('No AI providers configured',
              style: Theme.of(context).textTheme.titleLarge),
          const SizedBox(height: 8),
          const Text('Add an OpenAI-compatible provider to enable AI agents.'),
          const SizedBox(height: 24),
          FilledButton.icon(
            icon: const Icon(Icons.add),
            label: const Text('Add Provider'),
            onPressed: onAdd,
          ),
        ],
      ),
    );
  }
}

// ── Provider list ──────────────────────────────────────────────────────────

class _ProviderList extends StatelessWidget {
  final List<AiProvider> providers;
  final WidgetRef ref;

  const _ProviderList({required this.providers, required this.ref});

  @override
  Widget build(BuildContext context) {
    return ListView.builder(
      padding: const EdgeInsets.all(16),
      itemCount: providers.length,
      itemBuilder: (_, i) => _ProviderGlassCard(provider: providers[i], ref: ref),
    );
  }
}

class _ProviderCard extends StatelessWidget {
  final AiProvider provider;
  final WidgetRef ref;

  const _ProviderGlassCard({required this.provider, required this.ref});

  @override
  Widget build(BuildContext context) {
    return GlassCard(
      margin: const EdgeInsets.only(bottom: 12),
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                const Icon(Icons.psychology, color: Colors.indigo),
                const SizedBox(width: 8),
                Text(provider.name,
                    style: const TextStyle(
                        fontWeight: FontWeight.bold, fontSize: 16)),
                const Spacer(),
                if (provider.isOfficial)
                  Chip(
                    label: const Text('Official'),
                    backgroundColor: Colors.indigo.shade100,
                  ),
                IconButton(
                  icon: const Icon(Icons.edit_outlined),
                  tooltip: 'Edit API key',
                  onPressed: () => _showEditKeyDialog(context),
                ),
              ],
            ),
            const SizedBox(height: 8),
            Text(provider.baseUrl,
                style: Theme.of(context).textTheme.bodySmall),
            if (provider.models.isNotEmpty) ...[
              const SizedBox(height: 8),
              Wrap(
                spacing: 6,
                runSpacing: 4,
                children: provider.models
                    .map((m) => Chip(
                          label: Text(m),
                          visualDensity: VisualDensity.compact,
                        ))
                    .toList(),
              ),
            ],
          ],
        ),
      ),
    );
  }

  void _showEditKeyDialog(BuildContext context) {
    final ctrl = TextEditingController(text: provider.apiKey);
    showDialog(
      context: context,
      builder: (_) => AlertDialog(
        title: Text('API Key — ${provider.name}'),
        content: TextField(
          controller: ctrl,
          obscureText: true,
          decoration: const InputDecoration(
            labelText: 'API Key',
            border: OutlineInputBorder(),
            prefixIcon: Icon(Icons.key),
          ),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('Cancel'),
          ),
          FilledButton(
            onPressed: () async {
              final api = ref.read(apiServiceProvider);
              await api?.saveAiProviderKey(provider.id, ctrl.text.trim());
              ref.invalidate(_providersProvider);
              if (context.mounted) Navigator.pop(context);
            },
            child: const Text('Save'),
          ),
        ],
      ),
    );
  }
}

// ── Add provider dialog ────────────────────────────────────────────────────

const _presetProviders = [
  {'name': 'OpenAI', 'base_url': 'https://api.openai.com/v1', 'models': ['gpt-4o', 'gpt-4o-mini', 'gpt-4-turbo', 'gpt-3.5-turbo']},
  {'name': 'Anthropic', 'base_url': 'https://api.anthropic.com/v1', 'models': ['claude-opus-4-5', 'claude-sonnet-4-5', 'claude-haiku-4-5']},
  {'name': 'Ollama (local)', 'base_url': 'http://localhost:11434/v1', 'models': ['llama3', 'mistral', 'phi3']},
  {'name': 'Custom', 'base_url': '', 'models': []},
];

class _ProviderDialog extends StatefulWidget {
  final WidgetRef ref;
  const _ProviderDialog({required this.ref});

  @override
  State<_ProviderDialog> createState() => _ProviderDialogState();
}

class _ProviderDialogState extends State<_ProviderDialog> {
  int _presetIndex = 0;
  final _nameCtrl = TextEditingController();
  final _urlCtrl = TextEditingController();
  final _keyCtrl = TextEditingController();
  final _modelsCtrl = TextEditingController();
  bool _loading = false;

  @override
  void initState() {
    super.initState();
    _applyPreset(0);
  }

  void _applyPreset(int i) {
    final p = _presetProviders[i];
    _nameCtrl.text = p['name'] as String;
    _urlCtrl.text = p['base_url'] as String;
    _modelsCtrl.text = (p['models'] as List<dynamic>).join(', ');
    setState(() => _presetIndex = i);
  }

  @override
  void dispose() {
    _nameCtrl.dispose();
    _urlCtrl.dispose();
    _keyCtrl.dispose();
    _modelsCtrl.dispose();
    super.dispose();
  }

  Future<void> _submit() async {
    if (_nameCtrl.text.trim().isEmpty) return;
    setState(() => _loading = true);
    try {
      final models = _modelsCtrl.text
          .split(',')
          .map((s) => s.trim())
          .where((s) => s.isNotEmpty)
          .toList();
      final api = widget.ref.read(apiServiceProvider);
      await api?.addAiProvider(
        name: _nameCtrl.text.trim(),
        baseUrl: _urlCtrl.text.trim(),
        apiKey: _keyCtrl.text.trim(),
        models: models,
      );
      widget.ref.invalidate(_providersProvider);
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
      title: const Text('Add AI Provider'),
      content: SizedBox(
        width: 480,
        child: SingleChildScrollView(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              DropdownButtonFormField<int>(
                value: _presetIndex,
                decoration: const InputDecoration(
                  labelText: 'Preset',
                  border: OutlineInputBorder(),
                ),
                items: List.generate(
                  _presetProviders.length,
                  (i) => DropdownMenuItem(
                    value: i,
                    child: Text(_presetProviders[i]['name'] as String),
                  ),
                ),
                onChanged: (i) => i != null ? _applyPreset(i) : null,
              ),
              const SizedBox(height: 12),
              TextField(
                controller: _nameCtrl,
                decoration: const InputDecoration(
                    labelText: 'Provider Name', border: OutlineInputBorder()),
              ),
              const SizedBox(height: 12),
              TextField(
                controller: _urlCtrl,
                decoration: const InputDecoration(
                    labelText: 'Base URL', border: OutlineInputBorder()),
              ),
              const SizedBox(height: 12),
              TextField(
                controller: _keyCtrl,
                obscureText: true,
                decoration: const InputDecoration(
                  labelText: 'API Key',
                  border: OutlineInputBorder(),
                  prefixIcon: Icon(Icons.key),
                ),
              ),
              const SizedBox(height: 12),
              TextField(
                controller: _modelsCtrl,
                decoration: const InputDecoration(
                  labelText: 'Models (comma-separated)',
                  border: OutlineInputBorder(),
                  hintText: 'gpt-4o, gpt-4o-mini',
                ),
              ),
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
