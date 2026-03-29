import 'dart:ui';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:ohc_app/models/skill.dart';
import 'package:ohc_app/services/api_service.dart';

final _skillsProvider = FutureProvider<List<Skill>>((ref) async {
  final api = ref.watch(apiServiceProvider);
  if (api == null) return [];
  return api.listSkills();
});

final _selectedCategoryProvider = StateProvider<String>((ref) => 'all');

// ── Screen ─────────────────────────────────────────────────────────────────

class SkillsScreen extends ConsumerWidget {
  const SkillsScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final snapshot = ref.watch(_skillsProvider);
    final category = ref.watch(_selectedCategoryProvider);

    return Scaffold(
      appBar: AppBar(title: const Text('Skills & Plugins')),
      body: snapshot.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => Center(child: Text('Error: $e')),
        data: (skills) {
          final filtered = category == 'all'
              ? skills
              : skills.where((s) => s.category == category).toList();
          return Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              _CategoryBar(skills: skills),
              Expanded(
                child: filtered.isEmpty
                    ? const Center(child: Text('No skills in this category.'))
                    : _SkillList(skills: filtered, ref: ref),
              ),
            ],
          );
        },
      ),
    );
  }
}

// ── Category filter ────────────────────────────────────────────────────────

class _CategoryBar extends ConsumerWidget {
  final List<Skill> skills;
  const _CategoryBar({required this.skills});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final selected = ref.watch(_selectedCategoryProvider);
    final cats = ['all', 'builtin', 'official', 'community'];

    return SingleChildScrollView(
      scrollDirection: Axis.horizontal,
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      child: Row(
        children: cats.map((c) {
          final count = c == 'all'
              ? skills.length
              : skills.where((s) => s.category == c).length;
          return Padding(
            padding: const EdgeInsets.only(right: 8),
            child: FilterChip(
              label: Text('${c[0].toUpperCase()}${c.substring(1)} ($count)'),
              selected: selected == c,
              onSelected: (_) =>
                  ref.read(_selectedCategoryProvider.notifier).state = c,
            ),
          );
        }).toList(),
      ),
    );
  }
}

// ── Skill list ─────────────────────────────────────────────────────────────

class _SkillList extends StatelessWidget {
  final List<Skill> skills;
  final WidgetRef ref;

  const _SkillList({required this.skills, required this.ref});

  @override
  Widget build(BuildContext context) {
    return ListView.builder(
      padding: const EdgeInsets.all(16),
      itemCount: skills.length,
      itemBuilder: (_, i) => _SkillCard(skill: skills[i], ref: ref),
    );
  }
}

class _SkillCard extends StatefulWidget {
  final Skill skill;
  final WidgetRef ref;

  const _SkillCard({required this.skill, required this.ref});

  @override
  State<_SkillCard> createState() => _SkillCardState();
}

class _SkillCardState extends State<_SkillCard> {
  bool _busy = false;
  late bool _installed;
  late bool _enabled;

  @override
  void initState() {
    super.initState();
    _installed = widget.skill.installed;
    _enabled = widget.skill.enabled;
  }

  Color _categoryColor() {
    switch (widget.skill.category) {
      case 'builtin':
        return Colors.blue;
      case 'official':
        return Colors.green;
      default:
        return Colors.orange;
    }
  }

  Future<void> _toggleInstall() async {
    setState(() => _busy = true);
    try {
      final api = widget.ref.read(apiServiceProvider);
      if (_installed) {
        await api?.uninstallSkill(widget.skill.name);
        setState(() {
          _installed = false;
          _enabled = false;
        });
      } else {
        await api?.installSkill(widget.skill.name);
        setState(() => _installed = true);
      }
      widget.ref.invalidate(_skillsProvider);
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context)
            .showSnackBar(SnackBar(content: Text('Error: $e')));
      }
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  Future<void> _toggleEnable(bool val) async {
    setState(() => _busy = true);
    try {
      final api = widget.ref.read(apiServiceProvider);
      await api?.setSkillEnabled(widget.skill.name, val);
      setState(() => _enabled = val);
      widget.ref.invalidate(_skillsProvider);
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context)
            .showSnackBar(SnackBar(content: Text('Error: $e')));
      }
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final s = widget.skill;
    return ClipRRect(
      borderRadius: BorderRadius.circular(16),
      child: BackdropFilter(
        filter: ImageFilter.compose(
          outer: const ColorFilter.matrix(<double>[
            1.8, 0, 0, 0, 0,
            0, 1.8, 0, 0, 0,
            0, 0, 1.8, 0, 0,
            0, 0, 0, 1, 0
          ]),
          inner: ImageFilter.blur(sigmaX: 15.0, sigmaY: 15.0),
        ),
        child: Container(
          margin: const EdgeInsets.only(bottom: 12),
          decoration: BoxDecoration(
            color: const Color.fromRGBO(255, 255, 255, 0.05),
            borderRadius: BorderRadius.circular(16),
            border: Border.all(
              color: const Color.fromRGBO(255, 255, 255, 0.1),
            ),
          ),
          child: Padding(
            padding: const EdgeInsets.all(16),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  children: [
                    Chip(
                      label: Text(s.category),
                      backgroundColor: _categoryColor().withAlpha(40),
                      labelStyle: TextStyle(color: _categoryColor()),
                      visualDensity: VisualDensity.compact,
                    ),
                    const SizedBox(width: 8),
                    Text(s.name,
                        style: const TextStyle(
                            fontWeight: FontWeight.bold, fontSize: 15)),
                    const SizedBox(width: 4),
                    Text('v${s.version}',
                        style: Theme.of(context).textTheme.bodySmall),
                    const Spacer(),
                    if (_installed)
                      Switch(
                        value: _enabled,
                        onChanged: _busy ? null : _toggleEnable,
                      ),
                    const SizedBox(width: 8),
                    _busy
                        ? const SizedBox(
                            width: 24,
                            height: 24,
                            child: CircularProgressIndicator(strokeWidth: 2),
                          )
                        : OutlinedButton(
                            onPressed: _toggleInstall,
                            child: Text(_installed ? 'Remove' : 'Install'),
                          ),
                  ],
                ),
                if (s.description.isNotEmpty) ...[
                  const SizedBox(height: 8),
                  Text(s.description,
                      style: Theme.of(context).textTheme.bodySmall),
                ],
              ],
            ),
          ),
        ),
      ),
    );
  }
}
