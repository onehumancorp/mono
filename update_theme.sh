#!/bin/bash
# Find and replace font-family in main.dart
sed -i "s/fontFamily: 'Inter'/fontFamily: 'Outfit'/g" srcs/app/lib/main.dart
# We need to implement the design system tokens in main.dart or standard widgets.
