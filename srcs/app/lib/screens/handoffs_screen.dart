import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:intl/intl.dart';
import 'package:ohc_app/models/handoff.dart';
import 'package:ohc_app/services/api_service.dart';
import 'package:ohc_app/widgets/slide_to_approve.dart';

/// Screen for managing agent-to-human escalation handoffs.
class HandoffsScreen extends ConsumerStatefulWidget {
  const HandoffsScreen({super.key});

  @override
  ConsumerState<HandoffsScreen> createState() => _HandoffsScreenState();
}

class _HandoffsScreenState extends ConsumerState<HandoffsScreen> {
  late Future<List<HandoffPackage>> _handoffsFuture;
  final Set<String> _processingIds = {};

  @override
  void initState() {
    super.initState();
    _refresh();
  }

  void _refresh() {
    setState(() {
      _handoffsFuture = ref.read(apiServiceProvider)!.listHandoffs();
    });
  }

  Future<void> _handleApprove(String id) async {
    setState(() => _processingIds.add(id));
    try {
      await ref.read(apiServiceProvider)!.resolveHandoff(id, 'approved');
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Handoff approved successfully')),
        );
        _refresh();
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Error: $e'), backgroundColor: Colors.red),
        );
      }
    } finally {
      if (mounted) setState(() => _processingIds.remove(id));
    }
  }

  Future<void> _handleReject(String id) async {
    setState(() => _processingIds.add(id));
    try {
      await ref.read(apiServiceProvider)!.resolveHandoff(id, 'rejected');
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Handoff rejected')),
        );
        _refresh();
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Error: $e'), backgroundColor: Colors.red),
        );
      }
    } finally {
      if (mounted) setState(() => _processingIds.remove(id));
    }
  }

  @override
  Widget build(BuildContext context) {
    final colors = Theme.of(context).colorScheme;

    return Scaffold(
      appBar: AppBar(
        title: const Text('Handoffs & Escalations'),
        actions: [
          IconButton(
            onPressed: _refresh,
            icon: const Icon(Icons.refresh),
          ),
        ],
      ),
      body: FutureBuilder<List<HandoffPackage>>(
        future: _handoffsFuture,
        builder: (context, snapshot) {
          if (snapshot.connectionState == ConnectionState.waiting) {
            return const Center(child: CircularProgressIndicator());
          }

          if (snapshot.hasError) {
            return Center(
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  const Icon(Icons.error_outline, size: 60, color: Colors.red),
                  const SizedBox(height: 16),
                  Text('Failed to load handoffs', style: Theme.of(context).textTheme.titleLarge),
                  TextButton(onPressed: _refresh, child: const Text('Try Again')),
                ],
              ),
            );
          }

          final handoffs = snapshot.data ?? [];
          if (handoffs.isEmpty) {
            return Center(
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Icon(Icons.check_circle_outline, size: 64, color: colors.primary.withOpacity(0.5)),
                  const SizedBox(height: 16),
                  const Text('No pending handoffs', style: TextStyle(fontSize: 18, fontWeight: FontWeight.w500)),
                  const Text('Your agents are operating autonomously.', style: TextStyle(color: Colors.grey)),
                ],
              ),
            );
          }

          return ListView.builder(
            padding: const EdgeInsets.all(16),
            itemCount: handoffs.length,
            itemBuilder: (context, index) {
              final handoff = handoffs[index];
              final isProcessing = _processingIds.contains(handoff.id);

              return Card(
                margin: const EdgeInsets.only(bottom: 16),
                child: Padding(
                  padding: const EdgeInsets.all(20),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Row(
                        mainAxisAlignment: MainAxisAlignment.spaceBetween,
                        children: [
                          Container(
                            padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                            decoration: BoxDecoration(
                              color: colors.primaryContainer,
                              borderRadius: BorderRadius.circular(4),
                            ),
                            child: Text(
                              'Intent: ${handoff.intent.toUpperCase()}',
                              style: TextStyle(
                                fontSize: 10,
                                fontWeight: FontWeight.bold,
                                color: colors.onPrimaryContainer,
                              ),
                            ),
                          ),
                          Text(
                            DateFormat.yMMMd().add_jm().format(handoff.createdAt),
                            style: TextStyle(fontSize: 12, color: colors.onSurfaceVariant),
                          ),
                        ],
                      ),
                      const SizedBox(height: 16),
                      Text(
                        'Escalated by Agent: ${handoff.fromAgentId}',
                        style: const TextStyle(fontWeight: FontWeight.bold),
                      ),
                      const SizedBox(height: 8),
                      Text(
                        handoff.currentState,
                        style: const TextStyle(fontSize: 16),
                      ),
                      if (handoff.visualGroundTruth != null) ...[
                        const SizedBox(height: 16),
                        Container(
                          height: 200,
                          width: double.infinity,
                          decoration: BoxDecoration(
                            color: Colors.black12,
                            borderRadius: BorderRadius.circular(8),
                          ),
                          child: const Center(
                            child: Icon(Icons.image_outlined, size: 48, color: Colors.grey),
                          ),
                        ),
                      ],
                      const SizedBox(height: 24),
                      SlideToApprove(
                        disabled: isProcessing,
                        onApprove: () => _handleApprove(handoff.id),
                        onReject: () => _handleReject(handoff.id),
                      ),
                    ],
                  ),
                ),
              );
            },
          );
        },
      ),
    );
  }
}
