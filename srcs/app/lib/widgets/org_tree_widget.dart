import 'package:flutter/material.dart';
import 'package:ohc_app/models/dashboard.dart';

/// A recursive widget for rendering the organization hierarchy.
class OrgTreeWidget extends StatelessWidget {
  final List<OrganizationMember> members;
  final String? parentId;
  final int depth;

  const OrgTreeWidget({
    super.key,
    required this.members,
    this.parentId,
    this.depth = 0,
  });

  @override
  Widget build(BuildContext context) {
    final children = members.where((m) => m.managerId == parentId).toList();
    if (children.isEmpty) return const SizedBox.shrink();

    final colors = Theme.of(context).colorScheme;

    return Padding(
      padding: EdgeInsets.only(left: depth == 0 ? 0 : 20.0),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: children.map((member) {
          return Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              _OrgMemberRow(member: member),
              OrgTreeWidget(
                members: members,
                parentId: member.id,
                depth: depth + 1,
              ),
            ],
          );
        }).toList(),
      ),
    );
  }
}

class _OrgMemberRow extends StatelessWidget {
  final OrganizationMember member;

  const _OrgMemberRow({required this.member});

  @override
  Widget build(BuildContext context) {
    final colors = Theme.of(context).colorScheme;

    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 8.0),
      child: Row(
        children: [
          // Avatar
          Container(
            width: 36,
            height: 36,
            decoration: BoxDecoration(
              color: member.isHuman
                  ? colors.primaryContainer
                  : colors.secondaryContainer,
              shape: BoxShape.circle,
            ),
            child: Center(
              child: Text(
                _getInitials(member.role.isNotEmpty ? member.role : member.name),
                style: TextStyle(
                  fontSize: 12,
                  fontWeight: FontWeight.bold,
                  color: member.isHuman
                      ? colors.onPrimaryContainer
                      : colors.onSecondaryContainer,
                ),
              ),
            ),
          ),
          const SizedBox(width: 12),
          // Info
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  children: [
                    Text(
                      member.name,
                      style: const TextStyle(
                        fontWeight: FontWeight.w600,
                        fontSize: 14,
                      ),
                    ),
                    if (member.isHuman) ...[
                      const SizedBox(width: 8),
                      Container(
                        padding: const EdgeInsets.symmetric(
                          horizontal: 4,
                          vertical: 2,
                        ),
                        decoration: BoxDecoration(
                          color: colors.primary.withOpacity(0.1),
                          borderRadius: BorderRadius.circular(4),
                        ),
                        child: Text(
                          'YOU',
                          style: TextStyle(
                            fontSize: 10,
                            fontWeight: FontWeight.bold,
                            color: colors.primary,
                          ),
                        ),
                      ),
                    ],
                  ],
                ),
                Text(
                  member.role.replaceAll('_', ' '),
                  style: TextStyle(
                    fontSize: 12,
                    color: colors.onSurfaceVariant.withOpacity(0.7),
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  String _getInitials(String input) {
    if (input.isEmpty) return '';
    final parts = input.split('_');
    if (parts.length >= 2) {
      return (parts[0][0] + parts[1][0]).toUpperCase();
    }
    return input.substring(0, input.length >= 2 ? 2 : 1).toUpperCase();
  }
}
