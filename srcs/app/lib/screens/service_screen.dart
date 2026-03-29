import 'package:flutter/material.dart';
import 'package:ohc_app/theme.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:ohc_app/services/local_manager_service.dart';

class ServiceScreen extends ConsumerStatefulWidget {
  const ServiceScreen({super.key});

  @override
  ConsumerState<ServiceScreen> createState() => _ServiceScreenState();
}

class _ServiceScreenState extends ConsumerState<ServiceScreen> {
  bool _isRunning = false;
  String _doctorOutput = '';
  bool _isLoading = false;

  @override
  void initState() {
    super.initState();
    _checkStatus();
  }

  Future<void> _checkStatus() async {
    final service = ref.read(localManagerServiceProvider);
    final running = await service.isServiceRunning();
    if (mounted) {
      setState(() {
        _isRunning = running;
      });
    }
  }

  Future<void> _toggleService() async {
    setState(() => _isLoading = true);
    final service = ref.read(localManagerServiceProvider);
    try {
      if (_isRunning) {
        await service.stopService();
      } else {
        await service.startService();
      }
      await Future.delayed(const Duration(seconds: 2)); // Give it a moment
      await _checkStatus();
    } finally {
      if (mounted) setState(() => _isLoading = false);
    }
  }

  Future<void> _runDoctor() async {
    setState(() => _doctorOutput = 'Running doctor...');
    final service = ref.read(localManagerServiceProvider);
    final output = await service.runDoctor();
    if (mounted) {
      setState(() {
        _doctorOutput = output;
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Local Service Manager')),
      body: ListView(
        padding: const EdgeInsets.all(24),
        children: [
          _buildStatusGlassCard(),
          const SizedBox(height: 24),
          _buildActions(),
          const SizedBox(height: 24),
          _buildDoctorOutput(),
        ],
      ),
    );
  }

  Widget _buildStatusGlassCard() {
    return GlassCard(
      child: Padding(
        padding: const EdgeInsets.all(24),
        child: Column(
          children: [
            Icon(
              _isRunning ? Icons.check_circle : Icons.error_outline,
              color: _isRunning ? Colors.green : Colors.orange,
              size: 64,
            ),
            const SizedBox(height: 16),
            Text(
              _isRunning ? 'Service is Running' : 'Service is Stopped',
              style: Theme.of(context).textTheme.headlineSmall,
            ),
            const SizedBox(height: 8),
            const Text(
              'Default Port: 18789',
              style: TextStyle(color: Colors.grey),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildActions() {
    return Wrap(
      spacing: 16,
      runSpacing: 16,
      children: [
        ElevatedButton.icon(
          onPressed: _isLoading ? null : _toggleService,
          icon: Icon(_isRunning ? Icons.stop : Icons.play_arrow),
          label: Text(_isRunning ? 'Stop Service' : 'Start Service'),
          style: ElevatedButton.styleFrom(
            backgroundColor: _isRunning ? Colors.red.shade50 : null,
            foregroundColor: _isRunning ? Colors.red : null,
          ),
        ),
        OutlinedButton.icon(
          onPressed: _runDoctor,
          icon: const Icon(Icons.medical_services_outlined),
          label: const Text('Run Health Check'),
        ),
        OutlinedButton.icon(
          onPressed: () => ref.read(localManagerServiceProvider).restartService(),
          icon: const Icon(Icons.refresh),
          label: const Text('Force Restart'),
        ),
      ],
    );
  }

  Widget _buildDoctorOutput() {
    if (_doctorOutput.isEmpty) return const SizedBox.shrink();
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text('Diagnostics', style: Theme.of(context).textTheme.titleMedium),
        const SizedBox(height: 8),
        Container(
          width: double.infinity,
          padding: const EdgeInsets.all(16),
          decoration: BoxDecoration(
            color: Colors.black87,
            borderRadius: BorderRadius.circular(8),
          ),
          child: SelectableText(
            _doctorOutput,
            style: const TextStyle(
              color: Colors.greenAccent,
              fontFamily: 'monospace',
              fontSize: 12,
            ),
          ),
        ),
      ],
    );
  }
}
