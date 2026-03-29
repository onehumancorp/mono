import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:ohc_app/services/api_service.dart';
import 'dart:convert';
import 'package:http/http.dart' as http;
import 'package:ohc_app/services/auth_service.dart';

class ScalingScreen extends ConsumerStatefulWidget {
  const ScalingScreen({super.key});

  @override
  ConsumerState<ScalingScreen> createState() => _ScalingScreenState();
}

class _ScalingScreenState extends ConsumerState<ScalingScreen> {
  final _roleCtrl = TextEditingController();
  final _countCtrl = TextEditingController(text: '1');
  bool _scaling = false;
  final List<String> _logs = [];
  @override
  void dispose() {
    _roleCtrl.dispose();
    _countCtrl.dispose();
    super.dispose();
  }

  Future<void> _scale() async {
    final role = _roleCtrl.text.trim();
    final countStr = _countCtrl.text.trim();
    if (role.isEmpty || countStr.isEmpty) return;

    final count = int.tryParse(countStr);
    if (count == null) return;

    setState(() {
      _scaling = true;
      _logs.clear();
      _logs.add('Initiating scale for $role to $count...');
    });

    final api = ref.read(apiServiceProvider);
    if (api == null) return;

    try {
      await api.scaleRole(role, count);
      _listenToStream();
    } catch (e) {
      setState(() {
        _logs.add('Error: $e');
        _scaling = false;
      });
    }
  }

  void _listenToStream() async {
    final api = ref.read(apiServiceProvider);
    if (api == null) return;

    try {
      final response = await api.scaleStream();
      if (response.statusCode != 200) {
        setState(() => _logs.add('Stream error: ${response.statusCode}'));
        return;
      }

      response.stream.transform(utf8.decoder).listen((data) {
        if (!mounted) return;
        final lines = data.split('\n');
        for (final line in lines) {
          if (line.startsWith('data: ')) {
            final jsonStr = line.substring(6);
            if (jsonStr.isEmpty) continue;
            try {
               final event = jsonDecode(jsonStr);
               setState(() {
                 _logs.add(event['event'] ?? 'Unknown event');
               });
               if (event['status'] == 'Ready') {
                 setState(() => _scaling = false);
               }
            } catch (_) {}
          }
        }
      }, onError: (error) {
        if (!mounted) return;
         setState(() {
          _logs.add('Stream closed or error: $error');
          _scaling = false;
         });
      }, onDone: () {
         if (!mounted) return;
         setState(() => _scaling = false);
      });
    } catch (e) {
       if (!mounted) return;
       setState(() {
        _logs.add('Error connecting to stream: $e');
        _scaling = false;
       });
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Dynamic Scaling')),
      body: Padding(
        padding: const EdgeInsets.all(24.0),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text('Agent Fleet Scaling', style: Theme.of(context).textTheme.headlineSmall),
            const SizedBox(height: 16),
            Row(
              children: [
                Expanded(
                  child: TextField(
                    controller: _roleCtrl,
                    decoration: const InputDecoration(labelText: 'Role (e.g. Sales, SWE)', border: OutlineInputBorder()),
                  ),
                ),
                const SizedBox(width: 16),
                SizedBox(
                  width: 100,
                  child: TextField(
                    controller: _countCtrl,
                    keyboardType: TextInputType.number,
                    decoration: const InputDecoration(labelText: 'Target Count', border: OutlineInputBorder()),
                  ),
                ),
                const SizedBox(width: 16),
                FilledButton(
                  onPressed: _scaling ? null : _scale,
                  child: _scaling
                      ? const SizedBox(height: 20, width: 20, child: CircularProgressIndicator(color: Colors.white, strokeWidth: 2))
                      : const Text('Scale'),
                ),
              ],
            ),
            const SizedBox(height: 32),
            Text('Scaling Logs', style: Theme.of(context).textTheme.titleLarge),
            const SizedBox(height: 8),
            Expanded(
              child: Container(
                decoration: BoxDecoration(
                  color: Colors.black87,
                  borderRadius: BorderRadius.circular(8),
                ),
                padding: const EdgeInsets.all(16),
                child: ListView.builder(
                  itemCount: _logs.length,
                  itemBuilder: (context, index) {
                    return Padding(
                      padding: const EdgeInsets.only(bottom: 8.0),
                      child: Text(
                        '> ${_logs[index]}',
                        style: const TextStyle(color: Colors.greenAccent, fontFamily: 'monospace'),
                      ),
                    );
                  },
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
