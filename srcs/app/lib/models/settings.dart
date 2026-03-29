/// Application settings domain model.
class Settings {
  final String? minimaxApiKey;
  final String? theme;
  final String? backendUrl;
  final bool standaloneMode;

  const Settings({
    this.minimaxApiKey,
    this.theme,
    this.backendUrl,
    this.standaloneMode = false,
  });

  factory Settings.fromJson(Map<String, dynamic> json) {
    return Settings(
      minimaxApiKey: json['minimax_api_key'] as String? ?? json['minimaxApiKey'] as String?,
      theme: json['theme'] as String?,
      backendUrl: json['backend_url'] as String?,
      standaloneMode: json['standalone_mode'] as bool? ?? false,
    );
  }

  Map<String, dynamic> toJson() => {
        'minimax_api_key': minimaxApiKey,
        'theme': theme,
        'backend_url': backendUrl,
        'standalone_mode': standaloneMode,
      };
}
