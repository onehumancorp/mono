import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:ohc_app/services/api_service.dart';

class AgentHireWizardScreen extends ConsumerStatefulWidget {
  const AgentHireWizardScreen({super.key});

  @override
  ConsumerState<AgentHireWizardScreen> createState() => _AgentHireWizardScreenState();
}

class _AgentHireWizardScreenState extends ConsumerState<AgentHireWizardScreen> {
  int _step = 0;
  String _selectedRole = '';
  String _selectedProvider = '';
  final _nameController = TextEditingController();
  bool _isDeploying = false;
  bool _isLoading = true;
  List<String> _roles = [];
  List<AgentProvider> _providers = [];

  @override
  void initState() {
    super.initState();
    _fetchData();
  }

  Future<void> _fetchData() async {
    try {
      final api = ref.read(apiServiceProvider);
      if (api == null) return;
      
      final providers = await api.listAgentProviders();
      final rolesSet = <String>{};
      for (final p in providers) {
        rolesSet.addAll(p.supportedRoles);
      }
      
      if (mounted) {
        setState(() {
          _providers = providers;
          _roles = rolesSet.toList()..sort();
          if (_providers.isNotEmpty) {
            _selectedProvider = _providers.first.type;
          }
          _isLoading = false;
        });
      }
    } catch (e) {
      if (mounted) {
        setState(() => _isLoading = false);
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Failed to load providers: $e'), backgroundColor: Colors.red),
        );
      }
    }
  }

  @override
  void dispose() {
    _nameController.dispose();
    super.dispose();
  }

  String _formatRole(String role) {
    return role.replaceAll('_', ' ').toLowerCase().split(' ').map((word) => word[0].toUpperCase() + word.substring(1)).join(' ');
  }

  Future<void> _handleDeploy() async {
    setState(() => _isDeploying = true);
    try {
      final api = ref.read(apiServiceProvider);
      if (api != null) {
        await api.hireAgent(_nameController.text.trim(), _selectedRole, providerType: _selectedProvider);
        if (mounted) {
          context.go('/agents');
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(content: Text('Agent ${_nameController.text} hired successfully!')),
          );
        }
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Failed to hire agent: $e'), backgroundColor: Colors.red),
        );
      }
    } finally {
      if (mounted) setState(() => _isDeploying = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Hire New Agent'),
        leading: IconButton(
          icon: const Icon(Icons.close),
          onPressed: () => context.go('/agents'),
        ),
      ),
      body: Stepper(
        type: StepperType.horizontal,
        currentStep: _step,
        onStepContinue: () {
          if (_step < 3) {
            setState(() => _step++);
          } else {
            _handleDeploy();
          }
        },
        onStepCancel: () {
          if (_step > 0) {
            setState(() => _step--);
          }
        },
        controlsBuilder: (context, details) {
          return Padding(
            padding: const EdgeInsets.only(top: 24),
            child: Row(
              children: [
                if (_step < 3)
                  ElevatedButton(
                    onPressed: (_step == 0 && _selectedRole.isEmpty) ? null : details.onStepContinue,
                    child: const Text('Next'),
                  )
                else
                  ElevatedButton(
                    onPressed: _isDeploying ? null : _handleDeploy,
                    child: _isDeploying
                        ? const SizedBox(width: 20, height: 20, child: CircularProgressIndicator(strokeWidth: 2))
                        : const Text('Deploy Agent'),
                  ),
                const SizedBox(width: 12),
                if (_step > 0)
                  TextButton(
                    onPressed: details.onStepCancel,
                    child: const Text('Back'),
                  ),
              ],
            ),
          );
        },
        steps: [
          Step(
            title: const Text('Role'),
            isActive: _step >= 0,
            content: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const Text('Step 1 — Select Agent Role', style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold)),
                const SizedBox(height: 8),
                const Text('Choose the primary capability profile for this new agent.'),
                const SizedBox(height: 24),
                Wrap(
                  spacing: 12,
                  runSpacing: 12,
                  children: _roles.map((role) {
                    final isSelected = _selectedRole == role;
                    return ChoiceChip(
                      label: Text(_formatRole(role)),
                      selected: isSelected,
                      onSelected: (selected) {
                        setState(() => _selectedRole = selected ? role : '');
                        if (selected && _nameController.text.isEmpty) {
                          _nameController.text = 'Senior ${_formatRole(role)}';
                        }
                      },
                    );
                  }).toList(),
                ),
              ],
            ),
          ),
          Step(
            title: const Text('Provider'),
            isActive: _step >= 1,
            content: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const Text('Step 2 — Choose AI Provider', style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold)),
                const SizedBox(height: 8),
                const Text('Select the AI backend that will power this agent.'),
                const SizedBox(height: 16),
                if (_isLoading)
                  const Center(child: Padding(
                    padding: EdgeInsets.all(32.0),
                    child: CircularProgressIndicator(),
                  ))
                else if (_providers.isEmpty)
                  const Center(child: Text('No AI providers available. Please configure one in Integrations.'))
                else
                  ..._providers.map((p) => RadioListTile<String>(
                        title: Text(p.label),
                        subtitle: Text(p.description),
                        value: p.type,
                        groupValue: _selectedProvider,
                        onChanged: (val) => setState(() => _selectedProvider = val!),
                      )),
              ],
            ),
          ),
          Step(
            title: const Text('Details'),
            isActive: _step >= 2,
            content: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const Text('Step 3 — Agent Details', style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold)),
                const SizedBox(height: 16),
                TextField(
                  controller: _nameController,
                  decoration: const InputDecoration(
                    labelText: 'Agent Name',
                    border: OutlineInputBorder(),
                    hintText: 'e.g. Senior Software Engineer',
                  ),
                ),
                const SizedBox(height: 16),
                ListTile(
                  leading: const Icon(Icons.info_outline),
                  title: const Text('This name will appear in transcripts and the org chart.'),
                ),
              ],
            ),
          ),
          Step(
            title: const Text('Confirm'),
            isActive: _step >= 3,
            content: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const Text('Step 4 — Confirm Deployment', style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold)),
                const SizedBox(height: 16),
                Card(
                  child: ListTile(
                    leading: CircleAvatar(child: Text(_selectedRole.isNotEmpty ? _selectedRole[0] : '?')),
                    title: Text(_nameController.text),
                    subtitle: Text(_formatRole(_selectedRole)),
                    trailing: Text(_selectedProvider.toUpperCase()),
                  ),
                ),
                const SizedBox(height: 16),
                const Text(
                  'This agent will be immediately provisioned with a SPIFFE identity, connected to the orchestration hub, and assigned to the default org chart branch.',
                  style: TextStyle(color: Colors.grey),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}
