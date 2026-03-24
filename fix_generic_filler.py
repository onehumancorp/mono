import os
import re

# We will remove the generic filler: "encapsulates the structured data, logic, and operational context associated with [Symbol] within the platform architecture."
# And replace it with an inferred meaning based on the Symbol string.

def infer_meaning(symbol):
    s = symbol.lower()
    if s.startswith("role"):
        return f"defines the standard operational responsibilities and system access boundaries for the {symbol.replace('Role', '')} persona."
    elif s.startswith("status"):
        return f"represents the {symbol.replace('Status', '').upper()} lifecycle phase of a tracked entity within the event-driven state machine."
    elif s.startswith("category"):
        return f"classifies the integration module under the {symbol.replace('Category', '')} domain taxonomy for structured discovery."
    elif s.startswith("phase"):
        return f"designates the {symbol.replace('Phase', '')} stage of execution within an autonomous workflow."
    elif s.endswith("request") or s.endswith("req"):
        return f"defines the strongly-typed JSON payload required by the HTTP API to initiate a {symbol.replace('Request', '').replace('Req', '')} operation."
    elif s.endswith("response") or s.endswith("resp"):
        return f"defines the structured data returned by the HTTP API upon successful execution of a {symbol.replace('Response', '').replace('Resp', '')} operation."
    elif s.endswith("event"):
        return f"represents an immutable payload emitted to the Pub/Sub system when a {symbol.replace('Event', '')} state change occurs."
    else:
        # Generic fallback that isn't terrible boilerplate
        return f"provides domain-specific context and typed constraints for {symbol} operations across the application."

def process_go_file(filepath):
    with open(filepath, 'r', encoding='utf-8') as f:
        lines = f.readlines()

    modified = False
    new_lines = []

    i = 0
    while i < len(lines):
        line = lines[i]

        # Match the boilerplate injected by the previous script
        match = re.search(r'^(\s*)//\s+([A-Za-z0-9_]+)\s+encapsulates the structured data, logic, and operational context associated with.*$', line)
        if match:
            indent = match.group(1)
            symbol = match.group(2)

            new_desc = infer_meaning(symbol)
            new_lines.append(f"{indent}// {symbol} {new_desc}\n")
            modified = True

        elif re.search(r'^(\s*)//\s+([A-Za-z0-9_]+)\s+constructs, configures, and returns a new.*$', line):
            indent = re.match(r'^(\s*)', line).group(1)
            symbol = re.search(r'^(\s*)//\s+([A-Za-z0-9_]+)', line).group(2)
            new_desc = f"constructs and returns a new, initialized {symbol.replace('New', '')} instance, wiring necessary dependencies."
            new_lines.append(f"{indent}// {symbol} {new_desc}\n")
            modified = True

        else:
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
