import 'dart:convert';
import 'dart:ui';
import 'package:flutter/material.dart';
import '../services/api_service.dart';

class CapabilitiesScreen extends StatefulWidget {
  final ApiService apiService;
  const CapabilitiesScreen({Key? key, required this.apiService}) : super(key: key);

  @override
  State<CapabilitiesScreen> createState() => _CapabilitiesScreenState();
}

class _CapabilitiesScreenState extends State<CapabilitiesScreen> {
  List<dynamic> _plugins = [];
  bool _isLoading = true;

  @override
  void initState() {
    super.initState();
    _fetchCapabilities();
  }

  Future<void> _fetchCapabilities() async {
    try {
      final response = await widget.apiService.get('/api/capability/plugins');
      if (response.statusCode == 200) {
        setState(() {
          _plugins = jsonDecode(response.body);
          _isLoading = false;
        });
      }
    } catch (e) {
      setState(() {
        _isLoading = false;
      });
      print('Error fetching capabilities: $e');
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: Colors.black87,
      appBar: AppBar(
        title: const Text('Plugin Mesh', style: TextStyle(fontFamily: 'Outfit')),
        backgroundColor: Colors.transparent,
        elevation: 0,
      ),
      body: Stack(
        children: [
          // Background graphic
          Positioned.fill(
            child: Container(
              decoration: const BoxDecoration(
                gradient: LinearGradient(
                  colors: [Color(0xFF1E1E2C), Color(0xFF0D0D16)],
                  begin: Alignment.topLeft,
                  end: Alignment.bottomRight,
                ),
              ),
            ),
          ),
          _isLoading
              ? const Center(child: CircularProgressIndicator())
              : ListView.builder(
                  padding: const EdgeInsets.all(16.0),
                  itemCount: _plugins.length,
                  itemBuilder: (context, index) {
                    final plugin = _plugins[index];
                    return _buildGlassCard(plugin);
                  },
                ),
        ],
      ),
    );
  }

  Widget _buildGlassCard(dynamic plugin) {
    return Container(
      margin: const EdgeInsets.only(bottom: 16.0),
      child: ClipRRect(
        borderRadius: BorderRadius.circular(16.0),
        child: BackdropFilter(
          filter: ImageFilter.blur(sigmaX: 15.0, sigmaY: 15.0),
          child: Container(
            padding: const EdgeInsets.all(16.0),
            decoration: BoxDecoration(
              color: const Color.fromRGBO(255, 255, 255, 0.05),
              borderRadius: BorderRadius.circular(16.0),
              border: Border.all(
                color: const Color.fromRGBO(255, 255, 255, 0.1),
                width: 1.0,
              ),
            ),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  plugin['name'] ?? 'Unknown Plugin',
                  style: const TextStyle(
                    fontFamily: 'Outfit',
                    fontSize: 20,
                    fontWeight: FontWeight.bold,
                    color: Colors.white,
                  ),
                ),
                const SizedBox(height: 8),
                Text(
                  'Version: ${plugin['version'] ?? '1.0.0'}',
                  style: const TextStyle(
                    fontFamily: 'Inter',
                    fontSize: 14,
                    color: Colors.white70,
                  ),
                ),
                const SizedBox(height: 8),
                Text(
                  'Status: ${plugin['status'] ?? 'ACTIVE'}',
                  style: TextStyle(
                    fontFamily: 'Inter',
                    fontSize: 14,
                    color: (plugin['status'] == 'ACTIVE') ? Colors.greenAccent : Colors.orangeAccent,
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}
