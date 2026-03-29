/// User domain model for RBAC and identity.
class UserPublic {
  final String id;
  final String username;
  final String email;
  final List<String> roles;
  final bool active;
  final DateTime createdAt;

  const UserPublic({
    required this.id,
    required this.username,
    required this.email,
    required this.roles,
    required this.active,
    required this.createdAt,
  });

  factory UserPublic.fromJson(Map<String, dynamic> json) {
    return UserPublic(
      id: json['id'] as String,
      username: json['username'] as String,
      email: json['email'] as String? ?? '',
      roles: (json['roles'] as List<dynamic>?)?.cast<String>() ?? [],
      active: json['active'] as bool? ?? true,
      createdAt: json['created_at'] != null
          ? DateTime.parse(json['created_at'] as String)
          : DateTime.now(),
    );
  }

  Map<String, dynamic> toJson() => {
        'id': id,
        'username': username,
        'email': email,
        'roles': roles,
        'active': active,
        'created_at': createdAt.toIso8601String(),
      };

  bool get isAdmin => roles.contains('admin');
}
