#!/bin/bash
# The easiest way to deal with App.test.tsx is to just rewrite it to test the main components.
# Let's see what components are in App.tsx.
cat srcs/frontend/src/App.tsx | grep "export default function App" -A 20
