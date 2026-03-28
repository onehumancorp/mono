#!/bin/bash
# Script to launch the React frontend dev server from Bazel.
set -e

# Ensure node_modules are present if possible, 
# though usually Bazel should handle this.
if [ ! -d "node_modules" ]; then
    npm install
fi

npm run dev
