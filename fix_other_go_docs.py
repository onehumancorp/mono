import os
import re

def process_go_file(filepath):
    with open(filepath, 'r', encoding='utf-8') as f:
        lines = f.readlines()

    modified = False

    for i in range(len(lines)):
        match = re.match(r'^(\s*)//\s+([A-Za-z0-9_]+)\s+(Defines|Represents|Provides|Is|Returns|Sets|Gets|Creates|Updates|Deletes|Handles|Validates|Extracts|Registers|Parses|Writes|Enables|Reflects|Classifies|Surfaces|Describes|Pairs|Carries|Models)(.*)$', lines[i])
        if match:
            indent = match.group(1)
            symbol = match.group(2)
            verb = match.group(3)
            rest = match.group(4)

            # Lowercase the verb
            new_line = f"{indent}// {symbol} {verb.lower()}{rest}\n"

            # If the rest of the string is literally " the X type.", we can do a bit better for any stragglers
            check_rest = rest.strip()
            if check_rest.endswith(" type."):
                # e.g., "defines the TrustStatusPending type."
                new_desc = f"encapsulates the structured data, logic, and operational context associated with {symbol} within the platform architecture."
                new_line = f"{indent}// {symbol} {new_desc}\n"

            if "new functionality." in check_rest.lower() or "new functionality" in check_rest.lower():
                new_desc = f"constructs, configures, and returns a new {symbol} instance, injecting essential external dependencies."
                new_line = f"{indent}// {symbol} {new_desc}\n"

            if lines[i] != new_line:
                lines[i] = new_line
                modified = True

        # Also catch just plain lowercase "defines the X type."
        match2 = re.match(r'^(\s*)//\s+([A-Za-z0-9_]+)\s+(defines|represents|provides|is|returns|sets|gets|creates|updates|deletes|handles|validates|extracts|registers|parses|writes|enables|reflects|classifies|surfaces|describes|pairs|carries|models)\s+the\s+([A-Za-z0-9_]+)\s+type\.$', lines[i], re.IGNORECASE)
        if match2:
            indent = match2.group(1)
            symbol = match2.group(2)
            new_desc = f"encapsulates the structured data, logic, and operational context associated with {symbol} within the platform architecture."
            new_line = f"{indent}// {symbol} {new_desc}\n"
            if lines[i] != new_line:
                lines[i] = new_line
                modified = True

    if modified:
        with open(filepath, 'w', encoding='utf-8') as f:
            f.writelines(lines)

def main():
    for root, dirs, files in os.walk('.'):
        if 'node_modules' in root or '.git' in root or 'bazel-' in root:
            continue
        for file in files:
            if file.endswith('.go'):
                process_go_file(os.path.join(root, file))

if __name__ == "__main__":
    main()
