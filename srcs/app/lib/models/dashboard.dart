import 'package:ohc_app/models/agent.dart';

/// Dashboard snapshot domain model.
class DashboardSnapshot {
  final Organization organization;
  final List<dynamic> meetings; // Generic for now, can be specialized later
  final CostSummary costs;
  final List<Agent> agents;
  final List<StatusBucket> statuses;
  final DateTime updatedAt;

  const DashboardSnapshot({
    required this.organization,
    required this.meetings,
    required this.costs,
    required this.agents,
    required this.statuses,
    required this.updatedAt,
  });

  factory DashboardSnapshot.fromJson(Map<String, dynamic> json) {
    return DashboardSnapshot(
      organization: Organization.fromJson(json['organization'] as Map<String, dynamic>? ?? {}),
      meetings: json['meetings'] as List<dynamic>? ?? [],
      costs: CostSummary.fromJson(json['costs'] as Map<String, dynamic>? ?? json['costSummary'] as Map<String, dynamic>? ?? {}),
      agents: (json['agents'] as List<dynamic>?)
              ?.map((e) => Agent.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
      statuses: (json['statuses'] as List<dynamic>?)
              ?.map((e) => StatusBucket.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
      updatedAt: json['updatedAt'] != null || json['updated_at'] != null
          ? DateTime.parse((json['updatedAt'] ?? json['updated_at']) as String)
          : DateTime.now(),
    );
  }
}

class Organization {
  final String id;
  final String name;
  final String domain;
  final List<OrganizationMember> members;

  const Organization({
    required this.id,
    required this.name,
    required this.domain,
    required this.members,
  });

  factory Organization.fromJson(Map<String, dynamic> json) {
    return Organization(
      id: json['id'] as String? ?? '',
      name: json['name'] as String? ?? '',
      domain: json['domain'] as String? ?? '',
      members: (json['members'] as List<dynamic>?)
              ?.map((e) => OrganizationMember.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
    );
  }
}

class OrganizationMember {
  final String id;
  final String name;
  final String role;
  final String? managerId;
  final bool isHuman;

  const OrganizationMember({
    required this.id,
    required this.name,
    required this.role,
    this.managerId,
    required this.isHuman,
  });

  factory OrganizationMember.fromJson(Map<String, dynamic> json) {
    return OrganizationMember(
      id: json['id'] as String? ?? '',
      name: json['name'] as String? ?? '',
      role: json['role'] as String? ?? '',
      managerId: json['managerId'] as String? ?? json['manager_id'] as String?,
      isHuman: json['isHuman'] as bool? ?? json['is_human'] as bool? ?? false,
    );
  }
}

class CostSummary {
  final double totalCostUSD;
  final int totalTokens;
  final List<AgentCost> agents;

  const CostSummary({
    required this.totalCostUSD,
    required this.totalTokens,
    required this.agents,
  });

  factory CostSummary.fromJson(Map<String, dynamic> json) {
    return CostSummary(
      totalCostUSD: (json['totalCostUSD'] ?? json['total_cost_usd'] ?? 0.0).toDouble(),
      totalTokens: json['totalTokens'] ?? json['total_tokens'] ?? 0,
      agents: (json['agents'] as List<dynamic>?)
              ?.map((e) => AgentCost.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
    );
  }
}

class AgentCost {
  final String agentId;
  final double costUSD;
  final int tokenUsed;

  const AgentCost({
    required this.agentId,
    required this.costUSD,
    required this.tokenUsed,
  });

  factory AgentCost.fromJson(Map<String, dynamic> json) {
    return AgentCost(
      agentId: json['agentID'] ?? json['agent_id'] ?? '',
      costUSD: (json['costUSD'] ?? json['cost_usd'] ?? 0.0).toDouble(),
      tokenUsed: json['tokenUsed'] ?? json['token_used'] ?? 0,
    );
  }
}

class StatusBucket {
  final String status;
  final int count;

  const StatusBucket({
    required this.status,
    required this.count,
  });

  factory StatusBucket.fromJson(Map<String, dynamic> json) {
    return StatusBucket(
      status: json['status'] as String? ?? '',
      count: json['count'] as int? ?? 0,
    );
  }
}
