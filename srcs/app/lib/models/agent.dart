/// Agent domain model.
class Agent {
  final String id;
  final String name;
  final String role;
  final String status;
  final String organizationId;
  final DateTime createdAt;

  const Agent({
    required this.id,
    required this.name,
    required this.role,
    required this.status,
    required this.organizationId,
    required this.createdAt,
  });

  factory Agent.fromJson(Map<String, dynamic> json) {
    return Agent(
      id: json['id'] as String,
      name: json['name'] as String,
      role: json['role'] as String? ?? '',
      status: json['status'] as String? ?? 'pending',
      organizationId: json['organization_id'] as String? ?? '',
      createdAt: json['created_at'] != null
          ? DateTime.parse(json['created_at'] as String)
          : DateTime.now(),
    );
  }

  Map<String, dynamic> toJson() => {
        'id': id,
        'name': name,
        'role': role,
        'status': status,
        'organization_id': organizationId,
        'created_at': createdAt.toIso8601String(),
      };

  bool get isRunning => status == 'running';
  bool get isPending => status == 'pending';
}
