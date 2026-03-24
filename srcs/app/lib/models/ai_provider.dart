/// AI provider configuration model.
class AiProvider {
  final String id;
  final String name;
  final String baseUrl;
  final String apiKey;
  final List<String> models;
  final bool isOfficial;

  const AiProvider({
    required this.id,
    required this.name,
    required this.baseUrl,
    required this.apiKey,
    required this.models,
    required this.isOfficial,
  });

  factory AiProvider.fromJson(Map<String, dynamic> json) {
    return AiProvider(
      id: json['id'] as String? ?? '',
      name: json['name'] as String? ?? '',
      baseUrl: json['base_url'] as String? ?? '',
      apiKey: json['api_key'] as String? ?? '',
      models: (json['models'] as List<dynamic>?)?.cast<String>() ?? [],
      isOfficial: json['is_official'] as bool? ?? false,
    );
  }

  AiProvider copyWith({String? apiKey}) {
    return AiProvider(
      id: id,
      name: name,
      baseUrl: baseUrl,
      apiKey: apiKey ?? this.apiKey,
      models: models,
      isOfficial: isOfficial,
    );
  }

  Map<String, dynamic> toJson() => {
        'id': id,
        'name': name,
        'base_url': baseUrl,
        'api_key': apiKey,
        'models': models,
        'is_official': isOfficial,
      };
}
