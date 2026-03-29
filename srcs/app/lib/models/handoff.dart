/// Handoff domain model.
class HandoffPackage {
  final String id;
  final String fromAgentId;
  final String toHumanRole;
  final String intent;
  final int failedAttempts;
  final String currentState;
  final String? visualGroundTruth;
  final String status;
  final DateTime createdAt;

  const HandoffPackage({
    required this.id,
    required this.fromAgentId,
    required this.toHumanRole,
    required this.intent,
    required this.failedAttempts,
    required this.currentState,
    this.visualGroundTruth,
    required this.status,
    required this.createdAt,
  });

  factory HandoffPackage.fromJson(Map<String, dynamic> json) {
    return HandoffPackage(
      id: json['id'] as String,
      fromAgentId: json['from_agent_id'] as String? ?? json['fromAgentId'] as String? ?? '',
      toHumanRole: json['to_human_role'] as String? ?? json['toHumanRole'] as String? ?? '',
      intent: json['intent'] as String? ?? '',
      failedAttempts: json['failed_attempts'] as int? ?? json['failedAttempts'] as int? ?? 0,
      currentState: json['current_state'] as String? ?? json['currentState'] as String? ?? '',
      visualGroundTruth: json['visual_ground_truth'] as String? ?? json['visualGroundTruth'] as String?,
      status: json['status'] as String? ?? 'pending',
      createdAt: json['created_at'] != null
          ? DateTime.parse(json['created_at'] as String)
          : DateTime.now(),
    );
  }

  Map<String, dynamic> toJson() => {
        'id': id,
        'from_agent_id': fromAgentId,
        'to_human_role': toHumanRole,
        'intent': intent,
        'failed_attempts': failedAttempts,
        'current_state': currentState,
        'visual_ground_truth': visualGroundTruth,
        'status': status,
        'created_at': createdAt.toIso8601String(),
      };

  bool get isPending => status == 'pending';
  bool get isResolved => status == 'resolved';
}
