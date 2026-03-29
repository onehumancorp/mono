import re

with open('srcs/dashboard/BUILD.bazel', 'r') as f:
    content = f.read()

# remove duplicates manually using set for deps
def fix_deps(match):
    deps = match.group(1).split(',')
    deps = [d.strip() for d in deps if d.strip()]
    unique_deps = sorted(list(set(deps)))
    return 'deps = [\n        ' + ',\n        '.join(unique_deps) + ',\n    ]'

content = re.sub(r'deps\s*=\s*\[(.*?)\]', fix_deps, content, flags=re.DOTALL)

with open('srcs/dashboard/BUILD.bazel', 'w') as f:
    f.write(content)
