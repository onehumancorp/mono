import 'package:flutter/material.dart';
import 'package:ohc_app/theme.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:ohc_app/models/security_issue.dart';
import 'package:ohc_app/services/api_service.dart';

final _securityProvider = FutureProvider<List<SecurityIssue>>((ref) async {
  final api = ref.watch(apiServiceProvider);
  if (api == null) return [];
  return api.listSecurityIssues();
});

// ── Screen ─────────────────────────────────────────────────────────────────

class SecurityScreen extends ConsumerWidget {
  const SecurityScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final snapshot = ref.watch(_securityProvider);
    return Scaffold(
      appBar: AppBar(
        title: const Text('Security'),
        actions: [
          IconButton(
            icon: const Icon(Icons.refresh),
            tooltip: 'Re-scan',
            onPressed: () => ref.invalidate(_securityProvider),
          ),
          const SizedBox(width: 8),
        ],
      ),
      body: snapshot.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => Center(child: Text('Error: $e')),
        data: (issues) {
          if (issues.isEmpty) {
            return const _AllClear();
          }
          final open = issues.where((i) => !i.fixed).toList();
          final fixed = issues.where((i) => i.fixed).toList();
          return ListView(
            padding: const EdgeInsets.all(16),
            children: [
              if (open.isNotEmpty) ...[
                _SectionHeader(
                  '${open.length} open issue${open.length != 1 ? 's' : ''}',
                  color: open.any((i) => i.severity == 'high')
                      ? Colors.red
                      : Colors.orange,
                ),
                ...open.map((i) => _IssueGlassCard(issue: i, ref: ref)),
                const SizedBox(height: 16),
              ],
              if (fixed.isNotEmpty) ...[
                _SectionHeader(
                  '${fixed.length} resolved',
                  color: Colors.green,
                ),
                ...fixed.map((i) => _IssueGlassCard(issue: i, ref: ref)),
              ],
            ],
          );
        },
      ),
    );
  }
}

// ── Widgets ────────────────────────────────────────────────────────────────

class _AllClear extends StatelessWidget {
  const _AllClear();

  @override
  Widget build(BuildContext context) {
    return const Center(
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(Icons.verified_user, size: 64, color: Colors.green),
          SizedBox(height: 16),
          Text('No security issues found',
              style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold)),
          SizedBox(height: 8),
          Text('Your configuration looks good.'),
        ],
      ),
    );
  }
}

class _SectionHeader extends StatelessWidget {
  final String text;
  final Color color;
  const _SectionHeader(this.text, {required this.color});

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 8),
      child: Text(
        text,
        style: TextStyle(
          color: color,
          fontWeight: FontWeight.bold,
          fontSize: 14,
        ),
      ),
    );
  }
}

class _IssueCard extends StatefulWidget {
  final SecurityIssue issue;
  final WidgetRef ref;

  const _IssueGlassCard({required this.issue, required this.ref});

  @override
  State<_IssueCard> createState() => _IssueCardState();
}

class _IssueCardState extends State<_IssueCard> {
  bool _busy = false;
  late bool _fixed;

  @override
  void initState() {
    super.initState();
    _fixed = widget.issue.fixed;
  }

  Color _severityColor() {
    switch (widget.issue.severity) {
      case 'high':
        return Colors.red;
      case 'medium':
        return Colors.orange;
      default:
        return Colors.blue;
    }
  }

  Future<void> _fix() async {
    setState(() => _busy = true);
    try {
      final api = widget.ref.read(apiServiceProvider);
      await api?.fixSecurityIssue(widget.issue.id);
      setState(() => _fixed = true);
      widget.ref.invalidate(_securityProvider);
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context)
            .showSnackBar(SnackBar(content: Text('Fix failed: $e')));
      }
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final issue = widget.issue;
    return GlassCard(
      margin: const EdgeInsets.only(bottom: 10),
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Icon(
                  _fixed ? Icons.check_circle : Icons.warning_amber,
                  color: _fixed ? Colors.green : _severityColor(),
                ),
                const SizedBox(width: 8),
                Expanded(
                  child: Text(
                    issue.title,
                    style: const TextStyle(fontWeight: FontWeight.bold),
                  ),
                ),
                Chip(
                  label: Text(issue.severity.toUpperCase()),
                  backgroundColor: _severityColor().withAlpha(30),
                  labelStyle: TextStyle(
                      color: _severityColor(), fontWeight: FontWeight.bold),
                  visualDensity: VisualDensity.compact,
                ),
              ],
            ),
            const SizedBox(height: 8),
            Text(issue.description,
                style: Theme.of(context).textTheme.bodySmall),
            if (issue.detail != null) ...[
              const SizedBox(height: 4),
              Text(issue.detail!,
                  style: Theme.of(context).textTheme.bodySmall?.copyWith(
                        fontFamily: 'monospace',
                        color: Colors.grey,
                      )),
            ],
            if (issue.fixable && !_fixed) ...[
              const SizedBox(height: 12),
              FilledButton.icon(
                icon: _busy
                    ? const SizedBox(
                        width: 16,
                        height: 16,
                        child: CircularProgressIndicator(strokeWidth: 2),
                      )
                    : const Icon(Icons.build, size: 16),
                label: const Text('Auto-fix'),
                onPressed: _busy ? null : _fix,
              ),
            ],
          ],
        ),
      ),
    );
  }
}
