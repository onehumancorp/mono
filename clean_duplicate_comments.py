import os
import re

def process_go_file(filepath):
    with open(filepath, 'r', encoding='utf-8') as f:
        lines = f.readlines()

    modified = False
    new_lines = []

    # Iterate through the file and look for duplicate comment lines
    # Sometimes if a file had both standard GoDoc AND // Summary: earlier, the previous scripts duplicated them.
    # For instance, if there's two lines in a row or near each other describing the same struct.

    i = 0
    while i < len(lines):
        line = lines[i]

        # Check if the next line is exactly the same
        if i < len(lines) - 1 and lines[i].strip() == lines[i+1].strip() and lines[i].strip().startswith('//'):
            # Skip this line
            modified = True
            i += 1
            continue

        new_lines.append(line)
        i += 1

    if modified:
        with open(filepath, 'w', encoding='utf-8') as f:
            f.writelines(new_lines)

def main():
    for root, dirs, files in os.walk('.'):
        if 'node_modules' in root or '.git' in root or 'bazel-' in root:
            continue
        for file in files:
            if file.endswith('.go'):
                process_go_file(os.path.join(root, file))

if __name__ == "__main__":
    main()
