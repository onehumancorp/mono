import 'package:flutter/material.dart';
import 'package:ohc_app/theme.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:ohc_app/services/api_service.dart';

/// Screen for scaling the agent workforce dynamically.
class ScalingScreen extends ConsumerStatefulWidget {
  const ScalingScreen({super.key});

  @override
  ConsumerState<ScalingScreen> createState() => _ScalingScreenState();
}

class _ScalingScreenState extends ConsumerState<ScalingScreen> {
  String _selectedRole = 'SOFTWARE_ENGINEER';
  double _targetCount = 1;
  bool _isProvisioning = false;
  final List<String> _logs = [];

  final List<String> _roles = [
    'SOFTWARE_ENGINEER',
    'QA_TESTER',
    'DESIGNER',
    'SECURITY_ENGINEER',
    'PRODUCT_MANAGER',
  ];

  Future<void> _handleScale() async {
    setState(() {
      _isProvisioning = true;
      _logs.add('Starting provisioning for $_selectedRole...');
    });

    try {
      await ref.read(apiServiceProvider)!.scaleAgents(_selectedRole, _targetCount.toInt());
      
      // Simulate real-time logs (since we don't have a real stream yet)
      await Future.delayed(const Duration(milliseconds: 500));
      if (mounted) setState(() => _logs.add('Allocating compute resources...'));
      await Future.delayed(const Duration(milliseconds: 800));
      if (mounted) setState(() => _logs.add('Initializing runtime environments...'));
      await Future.delayed(const Duration(milliseconds: 600));
      if (mounted) setState(() => _logs.add('Injecting context and skills...'));
      await Future.delayed(const Duration(milliseconds: 400));
      
      if (mounted) {
        setState(() {
          _logs.add('SUCCESS: $_selectedRole scaled to $_targetCount instances.');
          _isProvisioning = false;
        });
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Scaling successful')),
        );
      }
    } catch (e) {
      if (mounted) {
        setState(() {
          _logs.add('ERROR: $e');
          _isProvisioning = false;
        });
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final colors = Theme.of(context).colorScheme;

    return Scaffold(
      appBar: AppBar(
        title: const Text('Dynamic Scaling'),
      ),
      body: Row(
        children: [
          // Scaling Controls
          Expanded(
            flex: 2,
            child: SingleChildScrollView(
              padding: const EdgeInsets.all(32),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    'Agent Workforce Scaling',
                    style: Theme.of(context).textTheme.headlineSmall?.copyWith(fontWeight: FontWeight.bold),
                  ),
                  const SizedBox(height: 8),
                  const Text(
                    'Provision additional specialized agents to handle peak demand or complex sub-tasks.',
                    style: TextStyle(color: Colors.grey),
                  ),
                  const SizedBox(height: 32),

                  // Step 1: Role
                  _SectionHeader(number: 1, title: 'Select Specialization'),
                  const SizedBox(height: 16),
                  Wrap(
                    spacing: 8,
                    runSpacing: 8,
                    children: _roles.map((role) {
                      final isSelected = _selectedRole == role;
                      return ChoiceChip(
                        label: Text(role.replaceAll('_', ' ')),
                        selected: isSelected,
                        onSelected: _isProvisioning ? null : (val) => setState(() => _selectedRole = role),
                      );
                    }).toList(),
                  ),
                  const SizedBox(height: 48),

                  // Step 2: Capacity
                  _SectionHeader(number: 2, title: 'Define Target Capacity'),
                  const SizedBox(height: 16),
                  GlassCard(
                    child: Padding(
                      padding: const EdgeInsets.all(24),
                      child: Column(
                        children: [
                          Row(
                            mainAxisAlignment: MainAxisAlignment.spaceBetween,
                            children: [
                              const Text('Target Count', style: TextStyle(fontWeight: FontWeight.bold)),
                              Text(
                                '${_targetCount.toInt()}',
                                style: TextStyle(
                                  fontSize: 24,
                                  fontWeight: FontWeight.bold,
                                  color: colors.primary,
                                ),
                              ),
                            ],
                          ),
                          Slider(
                            value: _targetCount,
                            min: 1,
                            max: 10,
                            divisions: 9,
                            onChanged: _isProvisioning ? null : (val) => setState(() => _targetCount = val),
                          ),
                          const Divider(height: 32),
                          Row(
                            children: [
                              const Icon(Icons.info_outline, size: 16, color: Colors.blue),
                              const SizedBox(width: 8),
                              const Text('Estimated Cost Impact:', style: TextStyle(fontSize: 12)),
                              const Spacer(),
                              Text(
                                '\$${(_targetCount * 0.45).toStringAsFixed(2)} / hr',
                                style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 12),
                              ),
                            ],
                          ),
                        ],
                      ),
                    ),
                  ),
                  const SizedBox(height: 48),

                  // Step 3: Action
                  SizedBox(
                    width: double.infinity,
                    height: 56,
                    child: FilledButton.icon(
                      onPressed: _isProvisioning ? null : _handleScale,
                      icon: _isProvisioning
                          ? const SizedBox(
                              width: 20,
                              height: 20,
                              child: CircularProgressIndicator(strokeWidth: 2, color: Colors.white),
                            )
                          : const Icon(Icons.rocket_launch),
                      label: Text(_isProvisioning ? 'Provisioning...' : 'Initiate Scaling'),
                    ),
                  ),
                ],
              ),
            ),
          ),

          // Scaling Logs (Sidebar)
          VerticalDivider(width: 1, color: colors.outlineVariant),
          Expanded(
            flex: 1,
            child: Container(
              color: colors.surfaceContainerLowest,
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Padding(
                    padding: const EdgeInsets.all(16),
                    child: Row(
                      children: [
                        const Icon(Icons.terminal, size: 18),
                        const SizedBox(width: 8),
                        const Text('Provisioning Logs', style: TextStyle(fontWeight: FontWeight.bold)),
                        const Spacer(),
                        if (_isProvisioning)
                          const SizedBox(
                            width: 12,
                            height: 12,
                            child: CircularProgressIndicator(strokeWidth: 2),
                          ),
                      ],
                    ),
                  ),
                  const Divider(height: 1),
                  Expanded(
                    child: ListView.builder(
                      padding: const EdgeInsets.all(16),
                      itemCount: _logs.length,
                      itemBuilder: (context, index) {
                        final log = _logs[index];
                        final isError = log.startsWith('ERROR');
                        final isSuccess = log.startsWith('SUCCESS');

                        return Padding(
                          padding: const EdgeInsets.only(bottom: 8),
                          child: Text(
                            log,
                            style: TextStyle(
                              fontFamily: 'monospace',
                              fontSize: 12,
                              color: isError
                                  ? Colors.red
                                  : isSuccess
                                      ? Colors.green
                                      : colors.onSurfaceVariant,
                            ),
                          ),
                        );
                      },
                    ),
                  ),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class _SectionHeader extends StatelessWidget {
  final int number;
  final String title;

  const _SectionHeader({required this.number, required this.title});

  @override
  Widget build(BuildContext context) {
    final colors = Theme.of(context).colorScheme;
    return Row(
      children: [
        Container(
          width: 24,
          height: 24,
          decoration: BoxDecoration(
            color: colors.primary,
            shape: BoxShape.circle,
          ),
          child: Center(
            child: Text(
              '$number',
              style: const TextStyle(color: Colors.white, fontSize: 12, fontWeight: FontWeight.bold),
            ),
          ),
        ),
        const SizedBox(width: 12),
        Text(
          title,
          style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 16),
        ),
      ],
    );
  }
}
