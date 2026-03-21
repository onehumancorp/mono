import os
import re

def parse_go_params(params_str):
    if not params_str.strip():
        return "None"

    parts = []
    # simplistic split
    for p in params_str.split(','):
        p = p.strip()
        if p:
            p_parts = p.split(' ')
            if len(p_parts) > 0:
                parts.append(p_parts[0])
    return ", ".join(parts) if parts else "None"

def extract_existing_comments(lines, index):
    # Search backwards from the line before index
    comment_block = []
    j = index - 1
    while j >= 0:
        line = lines[j].strip()
        if line.startswith('//'):
            comment_block.insert(0, lines[j])
            j -= 1
        else:
            break
    return comment_block

def build_go_docstring(name, params_str, returns_str, is_func, existing_comments):
    summary_lines = []
    for c in existing_comments:
        c = c.strip()
        if c.startswith('// '):
            c = c[3:]
        elif c.startswith('//'):
            c = c[2:]
        if not c.startswith('Summary:') and not c.startswith('Intent:') and not c.startswith('Params:') and not c.startswith('Parameters:') and not c.startswith('Returns:') and not c.startswith('Errors:') and not c.startswith('Side Effects:'):
             summary_lines.append(c)

    summary = ' '.join(summary_lines).strip()
    if not summary:
         summary = f"{name} functionality." if is_func else f"Defines the {name} type."

    p_str = parse_go_params(params_str) if is_func else "None"
    r_str = returns_str if is_func and returns_str else "None"
    e_str = "Returns an error if applicable" if is_func and "error" in returns_str else "None"

    return [
        f"// Summary: {summary}\n",
        f"// Parameters: {p_str}\n",
        f"// Returns: {r_str}\n",
        f"// Errors: {e_str}\n",
        f"// Side Effects: None\n"
    ]

def process_go_files():
    for root, _, files in os.walk('srcs'):
        for file in files:
            if file.endswith('.go') and not file.endswith('_test.go') and not file.endswith('.pb.go') and not '/proto/' in root and not 'cmd/' in root:
                filepath = os.path.join(root, file)

                # Restore to make sure we don't duplicate
                os.system(f"git restore {filepath}")

                with open(filepath, 'r') as f:
                    lines = f.readlines()

                new_lines = []
                i = 0
                while i < len(lines):
                    line = lines[i]

                    # 1) Match functions
                    func_match = re.match(r'^func (?:\([a-zA-Z0-9 *\[\]_]+\)\s*)?([A-Z][a-zA-Z0-9_]*)\((.*?)\)(.*?)\s*\{', line)
                    if func_match:
                        name = func_match.group(1)
                        params_str = func_match.group(2)
                        returns_str = func_match.group(3).strip()

                        comments = extract_existing_comments(lines, i)
                        if comments:
                            # Remove old comments from new_lines
                            new_lines = new_lines[:-len(comments)]

                        # We always rebuild to ensure correct format
                        docstring = build_go_docstring(name, params_str, returns_str, True, comments)
                        for d in docstring:
                            new_lines.append(d)

                        new_lines.append(line)
                        i += 1
                        continue

                    # 2) Match types (struct, interface, map, slice, etc.)
                    type_match = re.match(r'^type ([A-Z][a-zA-Z0-9_]*)\s+(?:struct|interface|func|map|\[\]|int|string|bool)', line)
                    if type_match:
                        name = type_match.group(1)

                        comments = extract_existing_comments(lines, i)
                        if comments:
                            new_lines = new_lines[:-len(comments)]

                        docstring = build_go_docstring(name, "", "", False, comments)
                        for d in docstring:
                            new_lines.append(d)

                        new_lines.append(line)
                        i += 1
                        continue

                    # 3) Match public variables/constants block
                    var_match = re.match(r'^(?:var|const)\s+([A-Z][a-zA-Z0-9_]*)\s*.*=', line)
                    if var_match:
                        name = var_match.group(1)

                        comments = extract_existing_comments(lines, i)
                        if comments:
                            new_lines = new_lines[:-len(comments)]

                        docstring = build_go_docstring(name, "", "", False, comments)
                        for d in docstring:
                            new_lines.append(d)

                        new_lines.append(line)
                        i += 1
                        continue

                    # 4) Handle `const (` or `var (` blocks
                    block_match = re.match(r'^(?:var|const)\s+\(', line)
                    if block_match:
                        new_lines.append(line)
                        i += 1
                        while i < len(lines) and not lines[i].startswith(')'):
                            inner_line = lines[i]
                            # Match public exported const/var
                            inner_match = re.match(r'^\s+([A-Z][a-zA-Z0-9_]*)\s*.*=?', inner_line)
                            if inner_match:
                                name = inner_match.group(1)
                                comments = extract_existing_comments(lines, i)
                                if comments:
                                    # Since they are indented, remove them based on line count
                                    new_lines = new_lines[:-len(comments)]

                                docstring = build_go_docstring(name, "", "", False, comments)
                                for d in docstring:
                                    new_lines.append("\t" + d) # Add indent
                            new_lines.append(inner_line)
                            i += 1
                        if i < len(lines):
                            new_lines.append(lines[i]) # The closing ')'
                            i += 1
                        continue

                    new_lines.append(line)
                    i += 1

                with open(filepath, 'w') as f:
                    f.writelines(new_lines)

process_go_files()
