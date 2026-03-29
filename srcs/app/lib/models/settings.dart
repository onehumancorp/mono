/// Application settings domain model.
class Settings {
  final String? minimaxApiKey;
  final String? theme;

  const Settings({
    this.minimaxApiKey,
    this.theme,
  });

  factory Settings.fromJson(Map<String, dynamic> json) {
    return Settings(
      minimaxApiKey: json['minimax_api_key'] as String? ?? json['minimaxApiKey'] as String?,
      theme: json['theme'] as String?,
    );
  }

  Map<String, dynamic> toJson() => {
        'minimax_api_key': minimaxApiKey,
        'theme': theme,
      };
}
