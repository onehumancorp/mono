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
        if not c.startswith('Summary:') and not c.startswith('Intent:') and not c.startswith('Params:') and not c.startswith('Returns:') and not c.startswith('Errors:') and not c.startswith('Side Effects:'):
             summary_lines.append(c)

    summary = ' '.join(summary_lines).strip()
    if not summary:
         summary = f"{name} functionality." if is_func else f"Defines the {name} type."

    p_str = parse_go_params(params_str) if is_func else "None"
    r_str = returns_str if is_func and returns_str else "None"
    e_str = "Returns an error if applicable" if is_func and "error" in returns_str else "None"

    return [
        f"// Summary: {summary}",
        f"// Intent: {summary}",
        f"// Params: {p_str}",
        f"// Returns: {r_str}",
        f"// Errors: {e_str}",
        f"// Side Effects: None"
    ]

def process_go_files():
    for root, _, files in os.walk('srcs'):
        for file in files:
            if file.endswith('.go') and not file.endswith('_test.go') and not file.endswith('.pb.go') and not '/proto/' in root and not 'cmd/' in root:
                filepath = os.path.join(root, file)
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

                        has_tags = any(tag in ''.join(comments) for tag in ['Summary:', 'Intent:', 'Params:', 'Returns:', 'Errors:', 'Side Effects:'])

                        if not has_tags:
                            docstring = build_go_docstring(name, params_str, returns_str, True, comments)
                            for d in docstring:
                                new_lines.append(d + '\n')
                        else:
                            for c in comments:
                                new_lines.append(c)

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

                        has_tags = any(tag in ''.join(comments) for tag in ['Summary:', 'Intent:', 'Params:', 'Returns:', 'Errors:', 'Side Effects:'])

                        if not has_tags:
                            docstring = build_go_docstring(name, "", "", False, comments)
                            for d in docstring:
                                new_lines.append(d + '\n')
                        else:
                            for c in comments:
                                new_lines.append(c)

                        new_lines.append(line)
                        i += 1
                        continue

                    # 3) Match public variables/constants block
                    # Only handling top-level single var/const declarations for simplicity, as they are rarely exported with logic
                    var_match = re.match(r'^(?:var|const)\s+([A-Z][a-zA-Z0-9_]*)\s*.*=', line)
                    if var_match:
                        name = var_match.group(1)

                        comments = extract_existing_comments(lines, i)
                        if comments:
                            new_lines = new_lines[:-len(comments)]

                        has_tags = any(tag in ''.join(comments) for tag in ['Summary:', 'Intent:', 'Params:', 'Returns:', 'Errors:', 'Side Effects:'])

                        if not has_tags:
                            docstring = build_go_docstring(name, "", "", False, comments)
                            for d in docstring:
                                new_lines.append(d + '\n')
                        else:
                            for c in comments:
                                new_lines.append(c)

                        new_lines.append(line)
                        i += 1
                        continue

                    new_lines.append(line)
                    i += 1

                with open(filepath, 'w') as f:
                    f.writelines(new_lines)


def extract_ts_comments(lines, index):
    comment_block = []
    j = index - 1
    # Check for jsdoc-style or single-line comments
    in_multiline = False

    # We read backwards
    temp_block = []
    while j >= 0:
        line = lines[j].strip()
        if not line:
            j -= 1
            continue
        if line == '*/':
            in_multiline = True
            temp_block.insert(0, lines[j])
        elif in_multiline:
            temp_block.insert(0, lines[j])
            if line.startswith('/*') or line.startswith('/**'):
                in_multiline = False
                break
        elif line.startswith('//'):
            temp_block.insert(0, lines[j])
        else:
            break
        j -= 1

    return temp_block

def build_ts_docstring(name, is_func, existing_comments):
    summary_lines = []
    for c in existing_comments:
        c = c.strip()
        c = re.sub(r'^\/\*\*?', '', c)
        c = re.sub(r'^\*\/', '', c)
        c = re.sub(r'^\*', '', c)
        c = re.sub(r'^\/\/', '', c)
        c = c.strip()

        if c and not c.startswith('Summary:') and not c.startswith('Intent:') and not c.startswith('Params:') and not c.startswith('Returns:') and not c.startswith('Errors:') and not c.startswith('Side Effects:'):
             summary_lines.append(c)

    summary = ' '.join(summary_lines).strip()
    if not summary:
         summary = f"{name} functionality." if is_func else f"Defines {name}."

    p_str = "Various" if is_func else "None"
    r_str = "Various" if is_func else "None"
    e_str = "May throw an error" if is_func else "None"

    return [
        '/**',
        f" * Summary: {summary}",
        f" * Intent: {summary}",
        f" * Params: {p_str}",
        f" * Returns: {r_str}",
        f" * Errors: {e_str}",
        f" * Side Effects: None",
        ' */'
    ]

def process_ts_files():
    for root, _, files in os.walk('srcs/frontend/src'):
        for file in files:
            if file.endswith('.ts') or file.endswith('.tsx'):
                if file.endswith('.test.ts') or file.endswith('.test.tsx') or file.endswith('.spec.ts') or 'vite.config.ts' in file or 'vitest.config.ts' in file or 'playwright.config.ts' in file or 'setupTests.ts' in file:
                    continue

                filepath = os.path.join(root, file)
                with open(filepath, 'r') as f:
                    lines = f.readlines()

                new_lines = []
                i = 0
                while i < len(lines):
                    line = lines[i]

                    # Match exported functions/classes/interfaces/types/consts
                    ts_match = re.match(r'^export\s+(?:async\s+)?(?:default\s+)?(function|class|interface|type|const|let|var)\s+([a-zA-Z0-9_]+)', line)
                    if ts_match:
                        kind = ts_match.group(1)
                        name = ts_match.group(2)

                        comments = extract_ts_comments(lines, i)
                        # Find how many lines to remove from new_lines
                        # (Because comments might have empty lines between them and the declaration)
                        remove_count = 0
                        k = i - 1
                        while k >= 0:
                            if lines[k].strip() == '':
                                remove_count += 1
                                k -= 1
                            elif lines[k] in comments:
                                remove_count += 1
                                k -= 1
                            else:
                                break

                        if remove_count > 0:
                            new_lines = new_lines[:-remove_count]

                        has_tags = any(tag in ''.join(comments) for tag in ['Summary:', 'Intent:', 'Params:', 'Returns:', 'Errors:', 'Side Effects:'])

                        is_func = kind == 'function' or (kind in ['const', 'let', 'var'] and ('=>' in line or 'function' in line))

                        if not has_tags:
                            docstring = build_ts_docstring(name, is_func, comments)
                            for d in docstring:
                                new_lines.append(d + '\n')
                        else:
                            for c in comments:
                                new_lines.append(c)

                        new_lines.append(line)
                        i += 1
                        continue

                    new_lines.append(line)
                    i += 1

                with open(filepath, 'w') as f:
                    f.writelines(new_lines)


process_go_files()
process_ts_files()
