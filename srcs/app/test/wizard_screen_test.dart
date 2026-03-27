import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:ohc_app/screens/wizard_screen.dart';

void main() {
  group('SetupWizardScreen', () {
    testWidgets('renders wizard with three steps', (tester) async {
      await tester.pumpWidget(
        const ProviderScope(
          child: MaterialApp(
            home: SetupWizardScreen(),
          ),
        ),
      );
      await tester.pump();

      // The step indicator should show labels for all three steps.
      expect(find.text('Server'), findsOneWidget);
      expect(find.text('AI Provider'), findsOneWidget);
      expect(find.text('Real-time'), findsOneWidget);
    });

    testWidgets('shows server step content by default', (tester) async {
      await tester.pumpWidget(
        const ProviderScope(
          child: MaterialApp(
            home: SetupWizardScreen(),
          ),
        ),
      );
      await tester.pump();

      expect(find.text('Server Settings'), findsOneWidget);
      expect(find.text('Next'), findsOneWidget);
    });

    testWidgets('can navigate to AI provider step', (tester) async {
      await tester.pumpWidget(
        const ProviderScope(
          child: MaterialApp(
            home: SetupWizardScreen(),
          ),
        ),
      );
      await tester.pump();

      await tester.tap(find.text('Next'));
      await tester.pumpAndSettle();

      expect(find.textContaining('AI Provider'), findsWidgets);
      expect(find.text('Back'), findsOneWidget);
    });

    testWidgets('can navigate to Centrifuge step', (tester) async {
      await tester.pumpWidget(
        const ProviderScope(
          child: MaterialApp(
            home: SetupWizardScreen(),
          ),
        ),
      );
      await tester.pump();

      // Step 1 → Step 2
      await tester.tap(find.text('Next'));
      await tester.pumpAndSettle();

      // Step 2 → Step 3
      await tester.tap(find.text('Next'));
      await tester.pumpAndSettle();

      expect(find.textContaining('Centrifuge'), findsWidgets);
      expect(find.text('Save Configuration'), findsOneWidget);
    });

    testWidgets('can navigate back from step 2', (tester) async {
      await tester.pumpWidget(
        const ProviderScope(
          child: MaterialApp(
            home: SetupWizardScreen(),
          ),
        ),
      );
      await tester.pump();

      await tester.tap(find.text('Next'));
      await tester.pumpAndSettle();

      await tester.tap(find.text('Back'));
      await tester.pumpAndSettle();

      expect(find.text('Server Settings'), findsOneWidget);
    });

    testWidgets('default listen address is pre-populated', (tester) async {
      await tester.pumpWidget(
        const ProviderScope(
          child: MaterialApp(
            home: SetupWizardScreen(),
          ),
        ),
      );
      await tester.pump();

      expect(find.widgetWithText(TextField, '0.0.0.0:18789'), findsOneWidget);
    });
  });

  group('WizardStatus', () {
    test('fromJson parses all fields', () {
      final json = {
        'configured': true,
        'steps': {
          'server': true,
          'ai_provider': false,
          'centrifuge': true,
        },
      };
      final status = WizardStatus.fromJson(json);
      expect(status.configured, isTrue);
      expect(status.serverStep, isTrue);
      expect(status.aiProviderStep, isFalse);
      expect(status.centrifugeStep, isTrue);
    });

    test('fromJson handles missing fields gracefully', () {
      final status = WizardStatus.fromJson({});
      expect(status.configured, isFalse);
      expect(status.serverStep, isFalse);
      expect(status.aiProviderStep, isFalse);
      expect(status.centrifugeStep, isFalse);
    });

    test('empty() returns all-false status', () {
      final status = WizardStatus.empty();
      expect(status.configured, isFalse);
      expect(status.serverStep, isFalse);
      expect(status.aiProviderStep, isFalse);
      expect(status.centrifugeStep, isFalse);
    });
  });
}
