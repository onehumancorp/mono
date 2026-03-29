import 'dart:ui';
import 'package:flutter/material.dart';

class OhcTheme {
  // Premium Glassmorphism Filter
  static final ImageFilter premiumFilter = ImageFilter.compose(
    outer: const ColorFilter.matrix(<double>[
      2.0, 0, 0, 0, 0,
      0, 2.0, 0, 0, 0,
      0, 0, 2.0, 0, 0,
      0, 0, 0, 1, 0,
    ]), // saturate(200%)
    inner: ImageFilter.blur(sigmaX: 20.0, sigmaY: 20.0), // blur(20px)
  );

  // Surface Decoration
  static final BoxDecoration surfaceDecoration = BoxDecoration(
    color: const Color.fromRGBO(255, 255, 255, 0.03),
    border: Border.all(
      color: const Color.fromRGBO(255, 255, 255, 0.08),
      width: 1.0,
    ),
    borderRadius: BorderRadius.circular(16.0),
  );

  // Text Styles
  static const TextStyle premiumText = TextStyle(
    fontFamily: 'Outfit',
    color: Color.fromRGBO(255, 255, 255, 0.9),
    letterSpacing: -0.02,
  );
}
