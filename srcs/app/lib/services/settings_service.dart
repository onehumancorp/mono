import 'dart:convert';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:shared_preferences/shared_preferences.dart';

/// Client-side settings for the OHC dashboard.
class ClientSettings {
  final String backendUrl;
  final bool standaloneMode;

  const ClientSettings({
    required this.backendUrl,
    required this.standaloneMode,
  });

  ClientSettings copyWith({String? backendUrl, bool? standaloneMode}) {
    return ClientSettings(
      backendUrl: backendUrl ?? this.backendUrl,
      standaloneMode: standaloneMode ?? this.standaloneMode,
    );
  }

  Map<String, dynamic> toJson() => {
        'backendUrl': backendUrl,
        'standaloneMode': standaloneMode,
      };

  factory ClientSettings.fromJson(Map<String, dynamic> json) {
    return ClientSettings(
      backendUrl: json['backendUrl'] as String? ?? 'http://localhost:18789',
      standaloneMode: json['standaloneMode'] as bool? ?? false,
    );
  }
}

final _prefsProvider = FutureProvider<SharedPreferences>(
  (_) => SharedPreferences.getInstance(),
);

final clientSettingsProvider = StateNotifierProvider<ClientSettingsNotifier, AsyncValue<ClientSettings>>((ref) {
  return ClientSettingsNotifier(ref);
});

class ClientSettingsNotifier extends StateNotifier<AsyncValue<ClientSettings>> {
  final Ref _ref;
  static const _key = 'client_settings';

  ClientSettingsNotifier(this._ref) : super(const AsyncLoading()) {
    _load();
  }

  Future<void> _load() async {
    state = const AsyncLoading();
    state = await AsyncValue.guard(() async {
      final prefs = await _ref.watch(_prefsProvider.future);
      final json = prefs.getString(_key);
      if (json == null) {
        // Check environment variable if web/desktop supports it via string.fromEnvironment
        const envUrl = String.fromEnvironment('BACKEND_URL', defaultValue: 'http://localhost:18789');
        return ClientSettings(backendUrl: envUrl, standaloneMode: false);
      }
      return ClientSettings.fromJson(jsonDecode(json) as Map<String, dynamic>);
    });
  }

  Future<void> updateBackendUrl(String url) async {
    final current = state.valueOrNull;
    if (current == null) return;
    state = AsyncData(current.copyWith(backendUrl: url));
    await _save();
  }

  Future<void> updateStandaloneMode(bool enabled) async {
    final current = state.valueOrNull;
    if (current == null) return;
    state = AsyncData(current.copyWith(standaloneMode: enabled));
    await _save();
  }

  Future<void> _save() async {
    final current = state.valueOrNull;
    if (current == null) return;
    final prefs = await _ref.read(_prefsProvider.future);
    await prefs.setString(_key, jsonEncode(current.toJson()));
  }
}
