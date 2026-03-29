#!/bin/bash
set -e

# Instead of using a py_test rule that lacks playwright dependencies,
# use a sh_test that runs the host python which has playwright.
python3 srcs/orchestration/chaos_verifier.py
