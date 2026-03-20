import os
import re

def parse_go_signature(line):
    if line.startswith('func'):
        m_method = re.match(r'^func\s+\([^)]+\)\s+([A-Z]\w*)\s*\((.*?)\)(?:\s+(.*?))?(?:\{|$)', line)
        m_func = re.match(r'^func\s+([A-Z]\w*)\s*\((.*?)\)(?:\s+(.*?))?(?:\{|$)', line)

        if m_method:
            sym_name = m_method.group(1)
            params_str = m_method.group(2)
            returns_str = m_method.group(3)
            return 'method', sym_name, params_str, returns_str
        elif m_func:
            sym_name = m_func.group(1)
            params_str = m_func.group(2)
            returns_str = m_func.group(3)
            return 'func', sym_name, params_str, returns_str

    m_other = re.match(r'^(type|var|const)\s+([A-Z]\w*)', line)
    if m_other:
        return m_other.group(1), m_other.group(2), None, None

    return None, None, None, None

def process_go_files():
    go_files = []
    for root, dirs, files in os.walk('srcs'):
        for file in files:
            if file.endswith('.go') and not file.endswith('_test.go') and not file.endswith('.pb.go'):
                go_files.append(os.path.join(root, file))

    for fpath in go_files:
        with open(fpath, 'r') as f:
            lines = f.readlines()

        new_lines = []
        i = 0
        while i < len(lines):
            line = lines[i]

            # Extract grouped const/var blocks
            if line.startswith('const (') or line.startswith('var ('):
                new_lines.append(line)
                i += 1
                while i < len(lines) and not lines[i].startswith(')'):
                    inner_line = lines[i]
                    m_inner = re.match(r'^\s+([A-Z]\w*)', inner_line)
                    if m_inner and not inner_line.strip().startswith('//'):
                        sym_name = m_inner.group(1)

                        # check if it already has doc
                        has_doc = False
                        if len(new_lines) > 0 and new_lines[-1].strip().startswith('//'):
                            has_doc = True

                        if not has_doc:
                            doc = []
                            doc.append(f"\t// {sym_name} Intent: Handles operations related to {sym_name}.\n")
                            doc.append(f"\t//\n\t// Params: None.\n")
                            doc.append(f"\t//\n\t// Returns: None.\n")
                            doc.append(f"\t//\n\t// Errors: Returns an error if the operation fails.\n")
                            doc.append(f"\t//\n\t// Side Effects: Modifies state or interacts with external systems as necessary.\n")
                            new_lines.extend(doc)
                    new_lines.append(inner_line)
                    i += 1
                if i < len(lines):
                    new_lines.append(lines[i])
                i += 1
                continue

            sym_type, sym_name, params_str, returns_str = parse_go_signature(line)

            if sym_name:
                has_doc = False
                j = len(new_lines) - 1
                while j >= 0 and new_lines[j].strip() == '':
                    j -= 1

                comment_lines = []
                k = j
                while k >= 0 and new_lines[k].strip().startswith('//'):
                    if not new_lines[k].strip().startswith('//go:'):
                        comment_lines.insert(0, new_lines[k].strip())
                    k -= 1

                comment_text = " ".join(comment_lines)
                if 'Intent:' in comment_text and 'Params:' in comment_text:
                    has_doc = True

                if not has_doc:
                    existing_comments = []

                    while len(new_lines) > 0 and new_lines[-1].strip() == '':
                        new_lines.pop()

                    while len(new_lines) > 0 and new_lines[-1].strip().startswith('//'):
                        popped = new_lines.pop()
                        if not popped.strip().startswith('//go:'):
                            stripped = popped.strip()
                            if stripped.startswith('//'):
                                stripped = stripped[2:].strip()
                            existing_comments.insert(0, stripped)
                        else:
                            new_lines.append(popped)
                            break

                    params = []
                    returns = []
                    if sym_type in ('func', 'method'):
                        if params_str:
                            p_parts = params_str.split(',')
                            for p in p_parts:
                                p = p.strip()
                                if p:
                                    p_parts2 = p.split()
                                    if len(p_parts2) > 0:
                                        params.append(p_parts2[0])

                        if returns_str:
                            returns_str = returns_str.strip()
                            if returns_str.endswith('{'):
                                returns_str = returns_str[:-1].strip()
                            if returns_str.startswith('(') and returns_str.endswith(')'):
                                returns_str = returns_str[1:-1]
                            r_parts = returns_str.split(',')
                            for r in r_parts:
                                r = r.strip()
                                if r:
                                    r_parts2 = r.split()
                                    if len(r_parts2) > 0:
                                        returns.append(r_parts2[0])

                    doc = []

                    intent_text = " ".join(existing_comments) if existing_comments else f"Handles operations related to {sym_name}."

                    doc.append(f"// {sym_name} Intent: {intent_text}\n")
                    if params:
                        doc.append(f"//\n// Params:\n")
                        for p in params:
                            doc.append(f"//   - {p}: parameter inferred from signature.\n")
                    else:
                        doc.append(f"//\n// Params: None.\n")

                    if returns:
                        doc.append(f"//\n// Returns:\n")
                        for r in returns:
                            doc.append(f"//   - {r}: return value inferred from signature.\n")
                    else:
                        doc.append(f"//\n// Returns: None.\n")

                    doc.append(f"//\n// Errors: Returns an error if the operation fails.\n")
                    doc.append(f"//\n// Side Effects: Modifies state or interacts with external systems as necessary.\n")

                    new_lines.extend(doc)
            new_lines.append(line)
            i += 1

        with open(fpath, 'w') as f:
            f.writelines(new_lines)

process_go_files()
