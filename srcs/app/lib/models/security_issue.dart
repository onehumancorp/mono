/// Security advisory or misconfiguration detected on the local system.
class SecurityIssue {
  final String id;
  final String title;
  final String description;
  final String severity; // high | medium | low
  final bool fixable;
  final bool fixed;
  final String category;
  final String? detail;

  const SecurityIssue({
    required this.id,
    required this.title,
    required this.description,
    required this.severity,
    required this.fixable,
    required this.fixed,
    required this.category,
    this.detail,
  });

  factory SecurityIssue.fromJson(Map<String, dynamic> json) {
    return SecurityIssue(
      id: json['id'] as String? ?? '',
      title: json['title'] as String? ?? '',
      description: json['description'] as String? ?? '',
      severity: json['severity'] as String? ?? 'low',
      fixable: json['fixable'] as bool? ?? false,
      fixed: json['fixed'] as bool? ?? false,
      category: json['category'] as String? ?? 'general',
      detail: json['detail'] as String?,
    );
  }

  SecurityIssue copyWith({bool? fixed}) {
    return SecurityIssue(
      id: id,
      title: title,
      description: description,
      severity: severity,
      fixable: fixable,
      fixed: fixed ?? this.fixed,
      category: category,
      detail: detail,
    );
  }
}
