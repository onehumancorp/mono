import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:ohc_app/services/manager_service.dart';
import 'package:ohc_app/services/installer_service.dart';

class ManagerScreen extends ConsumerStatefulWidget {
  const ManagerScreen({super.key});

  @override
  ConsumerState<ManagerScreen> createState() => _ManagerScreenState();
}

class _ManagerScreenState extends ConsumerState<ManagerScreen> {
  final _manager = ManagerService();
  final _installer = InstallerService();
  
  ServiceStatus? _status;
  Map<String, bool>? _env;
  List<String> _logs = [];
  bool _loading = false;

  @override
  void initState() {
    super.initState();
    _refresh();
  }

  Future<void> _refresh() async {
    setState(() => _loading = true);
    final status = await _manager.getStatus();
    final env = await _installer.checkEnvironment();
    final logs = await _manager.getLogs(20);
    setState(() {
      _status = status;
      _env = env;
      _logs = logs;
      _loading = false;
    });
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Platform Manager'),
        actions: [
          IconButton(
            icon: const Icon(Icons.refresh),
            onPressed: _loading ? null : _refresh,
          ),
        ],
      ),
      body: ListView(
        padding: const EdgeInsets.all(24),
        children: [
          _buildStatusCard(),
          const SizedBox(height: 24),
          _buildEnvironmentCard(),
          const SizedBox(height: 24),
          _buildLogCard(),
        ],
      ),
    );
  }

  Widget _buildStatusCard() {
    final running = _status?.running ?? false;
    return Card(
      elevation: 0,
      color: Theme.of(context).colorScheme.surfaceContainerHighest.withOpacity(0.3),
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(16),
        side: BorderSide(color: Theme.of(context).colorScheme.outlineVariant),
      ),
      child: Padding(
        padding: const EdgeInsets.all(24),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Icon(
                  running ? Icons.check_circle : Icons.error_outline,
                  color: running ? Colors.green : Colors.orange,
                  size: 32,
                ),
                const SizedBox(width: 16),
                Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      running ? 'Service Running' : 'Service Stopped',
                      style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 18),
                    ),
                    if (running)
                      Text('PID: ${_status?.pid} | Port: ${_status?.port}'),
                  ],
                ),
                const Spacer(),
                FilledButton.icon(
                  onPressed: _loading ? null : () async {
                    if (running) {
                      await _manager.stopService();
                    } else {
                      await _manager.startService();
                    }
                    _refresh();
                  },
                  icon: Icon(running ? Icons.stop : Icons.play_arrow),
                  label: Text(running ? 'Stop Service' : 'Start Service'),
                  style: FilledButton.styleFrom(
                    backgroundColor: running ? Colors.red.withOpacity(0.8) : null,
                  ),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildEnvironmentCard() {
    if (_env == null) return const SizedBox();
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        const Text('Environment', style: TextStyle(fontWeight: FontWeight.bold, fontSize: 16)),
        const SizedBox(height: 12),
        Card(
          elevation: 0,
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(16),
            side: BorderSide(color: Theme.of(context).colorScheme.outlineVariant),
          ),
          child: Column(
            children: [
              _buildEnvTile('Node.js', _env!['node']!, onFix: _installer.openNodeDownloadPage),
              const Divider(height: 1),
              _buildEnvTile('npm', _env!['npm']!),
              const Divider(height: 1),
              _buildEnvTile(
                'OpenClaw CLI',
                _env!['openclaw']!,
                onFix: () async {
                  await _installer.installOpenClaw();
                  _refresh();
                },
              ),
            ],
          ),
        ),
      ],
    );
  }

  Widget _buildEnvTile(String name, bool ok, {VoidCallback? onFix}) {
    return ListTile(
      title: Text(name),
      trailing: ok
          ? const Icon(Icons.check, color: Colors.green)
          : TextButton(
              onPressed: onFix,
              child: const Text('Install / Fix'),
            ),
    );
  }

  Widget _buildLogCard() {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        const Text('Recent Logs', style: TextStyle(fontWeight: FontWeight.bold, fontSize: 16)),
        const SizedBox(height: 12),
        Container(
          width: double.infinity,
          padding: const EdgeInsets.all(16),
          decoration: BoxDecoration(
            color: Colors.black.withOpacity(0.05),
            borderRadius: BorderRadius.circular(12),
            border: Border.all(color: Theme.of(context).colorScheme.outlineVariant),
          ),
          child: Text(
            _logs.isEmpty ? 'No logs available' : _logs.join('\n'),
            style: const TextStyle(fontFamily: 'monospace', fontSize: 12),
          ),
        ),
      ],
    );
  }
}
