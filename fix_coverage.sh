#!/bin/bash
/home/jules/go/bin/bazelisk coverage --cache_test_results=no //... && cat bazel-out/_coverage/_coverage_report.dat | awk -F":" '/^SF:/ {file=$2} /^LH:/ {lf=$2; cov=(lh/lf)*100; printf "%-50s %3d / %3d  (%.2f%%)\n", file, lh, lf, cov}' | sort -n -k4
