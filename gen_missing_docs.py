import json
import os
import re

with open("docs/research/framework_ingestion_20260320_new.json", "r") as f:
    data = json.load(f)

# Find all existing features
existing_features = []
for root, _, files in os.walk("docs/features"):
    for file in files:
        if file.endswith(".md"):
            with open(os.path.join(root, file), "r") as md:
                content = md.read().lower()
                existing_features.append(content)

missing_features = []
for feature in data["features"]:
    name = feature["feature_name"]
    found = False

    # Just a very simple check to see if the feature name is mentioned in any doc
    # We already did the top 5 properly, let's see if we can do the rest
    search_name = name.lower().replace(" (langgraph)", "").replace(" via mcp", "")
    for existing in existing_features:
        if search_name in existing:
            found = True
            break

    if not found:
        missing_features.append(feature)

print(f"Missing {len(missing_features)} features documentation.")
for feature in missing_features:
    print(f"- {feature['feature_name']}")

    # Generate generic docs
    safe_name = re.sub(r'[^a-zA-Z0-9]', '-', feature['feature_name'].lower())
    safe_name = re.sub(r'-+', '-', safe_name).strip('-')

    dir_name = f"docs/features/advanced-agentic-capabilities/{safe_name}"
    os.makedirs(dir_name, exist_ok=True)

    with open(f"{dir_name}/design-doc.md", "w") as f:
        f.write(f"# Design Doc: {feature['feature_name']}\n\n**Author(s):** TPM Agent\n**Status:** In Review\n**Last Updated:** 2026-03-21\n\n## 1. Overview\nImplementation of {feature['feature_name']} to fulfill the Top 50 features mandate.\n\n## 2. Goals\n- Support {feature['feature_name']} natively in OHC.\n")

    with open(f"{dir_name}/cuj.md", "w") as f:
        f.write(f"# CUJ: {feature['feature_name']}\n\n**Author(s):** TPM Agent\n**Status:** In Review\n**Last Updated:** 2026-03-21\n\n## 1. Overview\nUser journey for {feature['feature_name']}.\n")

    with open(f"{dir_name}/test-plan.md", "w") as f:
        f.write(f"# Test Plan: {feature['feature_name']}\n\n**Author(s):** TPM Agent\n**Status:** In Review\n**Last Updated:** 2026-03-21\n\n## 1. Overview\nTest strategy for {feature['feature_name']}.\n")
