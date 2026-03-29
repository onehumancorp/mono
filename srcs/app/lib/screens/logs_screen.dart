import 'dart:async';

import 'package:flutter/material.dart';
import 'package:ohc_app/theme.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:ohc_app/services/api_service.dart';

final _logsProvider = FutureProvider.family<List<String>, int>((ref, lines) async {
  final api = ref.watch(apiServiceProvider);
  if (api == null) return [];
  return api.getLogs(lines: lines);
});

class LogsScreen extends ConsumerStatefulWidget {
  const LogsScreen({super.key});

  @override
  ConsumerState<LogsScreen> createState() => _LogsScreenState();
}

class _LogsScreenState extends ConsumerState<LogsScreen> {
  int _lines = 100;
  Timer? _timer;
  final _scrollCtrl = ScrollController();

  @override
  void initState() {
    super.initState();
    // Auto-refresh logs every 3 seconds.
    _timer = Timer.periodic(const Duration(seconds: 3), (_) {
      ref.invalidate(_logsProvider(_lines));
    });
  }

  @override
  void dispose() {
    _timer?.cancel();
    _scrollCtrl.dispose();
    super.dispose();
  }

  void _scrollToBottom() {
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (_scrollCtrl.hasClients) {
        _scrollCtrl.jumpTo(_scrollCtrl.position.maxScrollExtent);
      }
    });
  }

  @override
  Widget build(BuildContext context) {
    final snapshot = ref.watch(_logsProvider(_lines));
    return Scaffold(
      appBar: AppBar(
        title: const Text('Service Logs'),
        actions: [
          // Line count selector
          DropdownButton<int>(
            value: _lines,
            underline: const SizedBox(),
            items: const [
              DropdownMenuItem(value: 50, child: Text('50 lines')),
              DropdownMenuItem(value: 100, child: Text('100 lines')),
              DropdownMenuItem(value: 500, child: Text('500 lines')),
            ],
            onChanged: (v) {
              if (v != null) setState(() => _lines = v);
            },
          ),
          const SizedBox(width: 8),
          IconButton(
            icon: const Icon(Icons.refresh),
            tooltip: 'Refresh',
            onPressed: () => ref.invalidate(_logsProvider(_lines)),
          ),
          const SizedBox(width: 8),
        ],
      ),
      body: snapshot.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => Center(child: Text('Error: $e')),
        data: (lines) {
          _scrollToBottom();
          if (lines.isEmpty) {
            return const Center(child: Text('No logs yet.'));
          }
          return Container(
            color: const Color(0xFF1a1a2e),
            child: ListView.builder(
              controller: _scrollCtrl,
              padding: const EdgeInsets.all(12),
              itemCount: lines.length,
              itemBuilder: (_, i) => _LogLine(line: lines[i], index: i),
            ),
          );
        },
      ),
    );
  }
}

class _LogLine extends StatelessWidget {
  final String line;
  final int index;

  const _LogLine({required this.line, required this.index});

  Color _color() {
    final lower = line.toLowerCase();
    if (lower.contains('error') || lower.contains('fatal')) return Colors.red.shade300;
    if (lower.contains('warn')) return Colors.orange.shade300;
    if (lower.contains('info')) return Colors.green.shade300;
    if (lower.contains('debug')) return Colors.grey.shade400;
    return Colors.grey.shade200;
  }

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 1),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          SizedBox(
            width: 40,
            child: Text(
              '${index + 1}',
              style: TextStyle(
                color: Colors.grey.shade600,
                fontFamily: 'monospace',
                fontSize: 12,
              ),
            ),
          ),
          Expanded(
            child: SelectableText(
              line,
              style: TextStyle(
                color: _color(),
                fontFamily: 'monospace',
                fontSize: 12,
              ),
            ),
          ),
        ],
      ),
    );
  }
}
