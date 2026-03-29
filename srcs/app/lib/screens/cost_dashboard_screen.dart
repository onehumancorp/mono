import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:intl/intl.dart';
import 'package:ohc_app/models/agent.dart';
import 'package:ohc_app/models/dashboard.dart';
import 'package:ohc_app/services/api_service.dart';

/// Screen for financial analytics and token usage monitoring.
class CostDashboardScreen extends ConsumerStatefulWidget {
  const CostDashboardScreen({super.key});

  @override
  ConsumerState<CostDashboardScreen> createState() => _CostDashboardScreenState();
}

class _CostDashboardScreenState extends ConsumerState<CostDashboardScreen> {
  late Future<DashboardSnapshot> _dashboardFuture;

  @override
  void initState() {
    super.initState();
    _refresh();
  }

  void _refresh() {
    setState(() {
      _dashboardFuture = ref.read(apiServiceProvider)!.getDashboard();
    });
  }

  @override
  Widget build(BuildContext context) {
    final colors = Theme.of(context).colorScheme;
    final currencyFormat = NumberFormat.currency(symbol: '\$');

    return Scaffold(
      appBar: AppBar(
        title: const Text('Cost & Token Usage'),
        actions: [
          IconButton(
            onPressed: _refresh,
            icon: const Icon(Icons.refresh),
          ),
        ],
      ),
      body: FutureBuilder<DashboardSnapshot>(
        future: _dashboardFuture,
        builder: (context, snapshot) {
          if (snapshot.connectionState == ConnectionState.waiting) {
            return const Center(child: CircularProgressIndicator());
          }

          if (snapshot.hasError) {
            return Center(child: Text('Error: ${snapshot.error}'));
          }

          final data = snapshot.data!;
          final costs = data.costs;

          return ListView(
            padding: const EdgeInsets.all(24),
            children: [
              // Summary Cards
              Row(
                children: [
                  Expanded(
                    child: _SummaryCard(
                      title: 'Total Spend',
                      value: currencyFormat.format(costs.totalCostUSD),
                      icon: Icons.account_balance_wallet,
                      color: Colors.green,
                    ),
                  ),
                  const SizedBox(width: 16),
                  Expanded(
                    child: _SummaryCard(
                      title: 'Total Tokens',
                      value: NumberFormat.compact().format(costs.totalTokens),
                      icon: Icons.generating_tokens,
                      color: Colors.blue,
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 32),

              // Usage per Agent Chart
              Text(
                'Usage per Agent',
                style: Theme.of(context).textTheme.titleLarge?.copyWith(fontWeight: FontWeight.bold),
              ),
              const SizedBox(height: 16),
              Card(
                child: Padding(
                  padding: const EdgeInsets.all(20),
                  child: Column(
                    children: costs.agents.map((agentCost) {
                      final agent = data.agents.firstWhere(
                        (a) => a.id == agentCost.agentId,
                        orElse: () => Agent(
                          id: agentCost.agentId,
                          name: 'Unknown Agent',
                          role: '',
                          status: '',
                          organizationId: '',
                          createdAt: DateTime.now(),
                        ),
                      );

                      final ratio = costs.totalCostUSD > 0
                          ? agentCost.costUSD / costs.totalCostUSD
                          : 0.0;

                      return Padding(
                        padding: const EdgeInsets.only(bottom: 16),
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Row(
                              mainAxisAlignment: MainAxisAlignment.spaceBetween,
                              children: [
                                Text(agent.name, style: const TextStyle(fontWeight: FontWeight.w500)),
                                Text(currencyFormat.format(agentCost.costUSD)),
                              ],
                            ),
                            const SizedBox(height: 8),
                            Stack(
                              children: [
                                Container(
                                  height: 8,
                                  width: double.infinity,
                                  decoration: BoxDecoration(
                                    color: colors.surfaceContainerHighest,
                                    borderRadius: BorderRadius.circular(4),
                                  ),
                                ),
                                FractionallySizedBox(
                                  widthFactor: ratio.clamp(0.0, 1.0),
                                  child: Container(
                                    height: 8,
                                    decoration: BoxDecoration(
                                      color: colors.primary,
                                      borderRadius: BorderRadius.circular(4),
                                    ),
                                  ),
                                ),
                              ],
                            ),
                            const SizedBox(height: 4),
                            Text(
                              '${NumberFormat.compact().format(agentCost.tokenUsed)} tokens',
                              style: TextStyle(fontSize: 10, color: colors.onSurfaceVariant),
                            ),
                          ],
                        ),
                      );
                    }).toList(),
                  ),
                ),
              ),

              const SizedBox(height: 32),
              // Organization Hierarchy Preview
              Text(
                'Organization View',
                style: Theme.of(context).textTheme.titleLarge?.copyWith(fontWeight: FontWeight.bold),
              ),
              const SizedBox(height: 16),
              Card(
                child: Padding(
                  padding: const EdgeInsets.all(20),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Row(
                        children: [
                          const Icon(Icons.business, size: 20),
                          const SizedBox(width: 8),
                          Text(
                            data.organization.name,
                            style: const TextStyle(fontWeight: FontWeight.bold),
                          ),
                          const Spacer(),
                          Text(data.organization.domain),
                        ],
                      ),
                      const Divider(height: 32),
                      ...data.organization.members.take(3).map((m) => ListTile(
                            leading: Icon(m.isHuman ? Icons.person : Icons.smart_toy, size: 20),
                            title: Text(m.name),
                            subtitle: Text(m.role),
                            dense: true,
                          )),
                      if (data.organization.members.length > 3)
                        Center(
                          child: TextButton(
                            onPressed: () {},
                            child: const Text('View Full Org Tree'),
                          ),
                        ),
                    ],
                  ),
                ),
              ),
            ],
          );
        },
      ),
    );
  }
}

class _SummaryCard extends StatelessWidget {
  final String title;
  final String value;
  final IconData icon;
  final Color color;

  const _SummaryCard({
    required this.title,
    required this.value,
    required this.icon,
    required this.color,
  });

  @override
  Widget build(BuildContext context) {
    final colors = Theme.of(context).colorScheme;

    return Card(
      child: Padding(
        padding: const EdgeInsets.all(20),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Icon(icon, color: color, size: 24),
            const SizedBox(height: 12),
            Text(
              title,
              style: TextStyle(
                fontSize: 12,
                color: colors.onSurfaceVariant,
                fontWeight: FontWeight.w500,
              ),
            ),
            const SizedBox(height: 4),
            Text(
              value,
              style: const TextStyle(
                fontSize: 24,
                fontWeight: FontWeight.bold,
              ),
            ),
          ],
        ),
      ),
    );
  }
}
