import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:ohc_app/screens/login_screen.dart';
import 'package:ohc_app/screens/dashboard_screen.dart';
import 'package:ohc_app/screens/agents_screen.dart';
import 'package:ohc_app/screens/meetings_screen.dart';
import 'package:ohc_app/screens/chat_screen.dart';
import 'package:ohc_app/screens/channels_screen.dart';
import 'package:ohc_app/screens/ai_config_screen.dart';
import 'package:ohc_app/screens/skills_screen.dart';
import 'package:ohc_app/screens/logs_screen.dart';
import 'package:ohc_app/screens/security_screen.dart';
import 'package:ohc_app/screens/settings_screen.dart';
import 'package:ohc_app/screens/service_screen.dart';
import 'package:ohc_app/screens/wizard_screen.dart';
import 'package:ohc_app/screens/scaling_screen.dart';
import 'package:ohc_app/services/auth_service.dart';
import 'package:flutter/material.dart';

final routerProvider = Provider<GoRouter>((ref) {
  final authState = ref.watch(authStateProvider);

  return GoRouter(
    initialLocation: '/login',
    redirect: (context, state) {
      final isLoggedIn = authState.valueOrNull != null;
      final isLoginRoute = state.matchedLocation == '/login';

      if (!isLoggedIn && !isLoginRoute) return '/login';
      if (isLoggedIn && isLoginRoute) return '/dashboard';
      return null;
    },
    routes: [
      GoRoute(
        path: '/login',
        builder: (context, state) => const LoginScreen(),
      ),
      ShellRoute(
        builder: (context, state, child) => AppShell(child: child),
        routes: [
          GoRoute(
            path: '/dashboard',
            builder: (context, state) => const DashboardScreen(),
          ),
          GoRoute(
            path: '/agents',
            builder: (context, state) => const AgentsScreen(),
          ),
          GoRoute(
            path: '/meetings',
            builder: (context, state) => const MeetingsScreen(),
          ),
          GoRoute(
            path: '/chat',
            builder: (context, state) => const ChatScreen(),
          ),
          GoRoute(
            path: '/channels',
            builder: (context, state) => const ChannelsScreen(),
          ),
          GoRoute(
            path: '/ai-config',
            builder: (context, state) => const AiConfigScreen(),
          ),
          GoRoute(
            path: '/skills',
            builder: (context, state) => const SkillsScreen(),
          ),
          GoRoute(
            path: '/logs',
            builder: (context, state) => const LogsScreen(),
          ),
          GoRoute(
            path: '/security',
            builder: (context, state) => const SecurityScreen(),
          ),
          GoRoute(
            path: '/settings',
            builder: (context, state) => const SettingsScreen(),
          ),
          GoRoute(
            path: '/service',
            builder: (context, state) => const ServiceScreen(),
          ),
          GoRoute(
            path: '/wizard',
            builder: (context, state) => const SetupWizardScreen(),
          ),
          GoRoute(
            path: '/scaling',
            builder: (context, state) => const ScalingScreen(),
          ),
        ],
      ),
    ],
  );
});

/// Persistent shell (sidebar + navigation) wrapping all authenticated routes.
class AppShell extends StatelessWidget {
  final Widget child;
  const AppShell({super.key, required this.child});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: Row(
        children: [
          _Sidebar(),
          Expanded(child: child),
        ],
      ),
    );
  }
}

class _Sidebar extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return NavigationDrawer(
      children: [
        const SizedBox(height: 16),
        const Padding(
          padding: EdgeInsets.symmetric(horizontal: 16),
          child: Text(
            'One Human Corp',
            style: TextStyle(fontWeight: FontWeight.bold, fontSize: 16),
          ),
        ),
        const Divider(),
        _NavItem(icon: Icons.dashboard, label: 'Dashboard', path: '/dashboard'),
        _NavItem(icon: Icons.smart_toy, label: 'Agents', path: '/agents'),
        _NavItem(icon: Icons.video_call, label: 'Meetings', path: '/meetings'),
        _NavItem(icon: Icons.chat, label: 'Chat', path: '/chat'),
        _NavItem(icon: Icons.chat_bubble_outline, label: 'Channels', path: '/channels'),
        _NavItem(icon: Icons.trending_up, label: 'Scaling', path: '/scaling'),
        const Divider(),
        _NavItem(icon: Icons.psychology, label: 'AI Providers', path: '/ai-config'),
        _NavItem(icon: Icons.extension, label: 'Skills', path: '/skills'),
        _NavItem(icon: Icons.security, label: 'Security', path: '/security'),
        _NavItem(icon: Icons.terminal, label: 'Logs', path: '/logs'),
        const SizedBox(height: 8),
        _NavItem(icon: Icons.settings, label: 'Settings', path: '/settings'),
        _NavItem(icon: Icons.computer, label: 'Service Management', path: '/service'),
        _NavItem(icon: Icons.auto_fix_high, label: 'Setup Wizard', path: '/wizard'),
        const SizedBox(height: 16),
      ],
    );
  }
}

class _NavItem extends StatelessWidget {
  final IconData icon;
  final String label;
  final String path;

  const _NavItem({
    required this.icon,
    required this.label,
    required this.path,
  });

  @override
  Widget build(BuildContext context) {
    final current = GoRouterState.of(context).matchedLocation;
    final selected = current.startsWith(path);
    return ListTile(
      leading: Icon(icon),
      title: Text(label),
      selected: selected,
      onTap: () => context.go(path),
    );
  }
}
