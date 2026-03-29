/// SDLC Pipeline domain model.
class Pipeline {
  final String id;
  final String name;
  final String status;
  final String branch;
  final String? stagingUrl;
  final String? initiatedBy;
  final DateTime createdAt;
  final DateTime updatedAt;

  const Pipeline({
    required this.id,
    required this.name,
    required this.status,
    required this.branch,
    this.stagingUrl,
    this.initiatedBy,
    required this.createdAt,
    required this.updatedAt,
  });

  factory Pipeline.fromJson(Map<String, dynamic> json) {
    return Pipeline(
      id: json['id'] as String? ?? '',
      name: json['name'] as String? ?? 'Unnamed Pipeline',
      status: json['status'] as String? ?? 'pending',
      branch: json['branch'] as String? ?? 'main',
      stagingUrl: json['staging_url'] as String? ?? json['stagingUrl'] as String?,
      initiatedBy: json['initiated_by'] as String? ?? json['initiatedBy'] as String?,
      createdAt: json['created_at_unix'] != null
          ? DateTime.fromMillisecondsSinceEpoch((json['created_at_unix'] as int) * 1000)
          : json['created_at'] != null
              ? DateTime.parse(json['created_at'] as String)
              : DateTime.now(),
      updatedAt: json['updated_at_unix'] != null
          ? DateTime.fromMillisecondsSinceEpoch((json['updated_at_unix'] as int) * 1000)
          : json['updated_at'] != null
              ? DateTime.parse(json['updated_at'] as String)
              : DateTime.now(),
    );
  }

  Map<String, dynamic> toJson() => {
        'id': id,
        'name': name,
        'status': status,
        'branch': branch,
        'staging_url': stagingUrl,
        'initiated_by': initiatedBy,
        'created_at': createdAt.toIso8601String(),
        'updated_at': updatedAt.toIso8601String(),
      };

  bool get isActive => status == 'active' || status == 'running';
  bool get isCompleted => status == 'completed' || status == 'merged';
}
