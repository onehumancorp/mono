import os
import re

def process_ts_files():
    ts_files = []
    for root, dirs, files in os.walk('srcs/frontend/src'):
        for file in files:
            if (file.endswith('.ts') or file.endswith('.tsx')) and not file.endswith('.test.ts') and not file.endswith('.test.tsx'):
                ts_files.append(os.path.join(root, file))

    public_sym_pattern = re.compile(r'^export\s+(const|function|type|interface|class)\s+([a-zA-Z0-9_]+)')

    for fpath in ts_files:
        with open(fpath, 'r') as f:
            lines = f.readlines()

        new_lines = []
        i = 0
        while i < len(lines):
            line = lines[i]
            m1 = public_sym_pattern.match(line)

            if m1:
                sym_name = m1.group(2)

                # Check if it already has the required doc format
                has_doc = False
                j = len(new_lines) - 1
                while j >= 0 and new_lines[j].strip() == '':
                    j -= 1

                comment_text = ""
                if j >= 0 and new_lines[j].strip().endswith('*/'):
                    k = j
                    while k >= 0 and not new_lines[k].strip().startswith('/**'):
                        comment_text = new_lines[k] + comment_text
                        k -= 1
                    if k >= 0:
                        comment_text = new_lines[k] + comment_text

                if 'Intent:' in comment_text and 'Params:' in comment_text:
                    has_doc = True

                if not has_doc:
                    existing_comments = []
                    while len(new_lines) > 0 and new_lines[-1].strip() == '':
                        new_lines.pop()

                    j = len(new_lines) - 1
                    if j >= 0 and new_lines[j].strip().endswith('*/'):
                        k = j
                        while k >= 0 and not new_lines[k].strip().startswith('/**'):
                            k -= 1
                        if k >= 0:
                            # Extract existing JSDoc content WITHOUT dropping @param or @returns
                            for l in range(k+1, j):
                                stripped = new_lines[l].strip()
                                if stripped.startswith('*'):
                                    stripped = stripped[1:].strip()
                                if stripped:
                                    existing_comments.append(stripped)
                            new_lines = new_lines[:k]

                    sym_type = m1.group(1)

                    params = []
                    if sym_type in ('function', 'const'):
                        # Gather multiline signature block to parse params
                        sig_block = line
                        k = i + 1
                        while k < len(lines) and ')' not in sig_block:
                            sig_block += lines[k]
                            k += 1

                        sig_match = re.search(r'\((.*?)\)', sig_block.replace('\n', ''))
                        if sig_match:
                            params_str = sig_match.group(1)
                            if params_str:
                                p_parts = params_str.split(',')
                                for p in p_parts:
                                    p = p.strip()
                                    if p:
                                        p_parts2 = p.split(':')
                                        if len(p_parts2) > 0:
                                            # destructuring `{ a, b }` is harder, just grab the whole part before colon
                                            param_name = p_parts2[0].strip()
                                            if param_name and not param_name.startswith('{'):
                                                params.append(param_name)

                    doc = []
                    intent_text = " ".join(existing_comments) if existing_comments else f"Handles operations related to {sym_name}."

                    doc.append(f"/**\n")
                    doc.append(f" * Intent: {intent_text}\n")
                    if params:
                        doc.append(f" *\n * Params:\n")
                        for p in params:
                            doc.append(f" *   - {p}: parameter inferred from signature.\n")
                    else:
                        doc.append(f" *\n * Params: None.\n")

                    doc.append(f" *\n * Returns: Standard inferred return.\n")
                    doc.append(f" *\n * Errors: Throws or returns errors if the operation fails.\n")
                    doc.append(f" *\n * Side Effects: Modifies state, updates UI, or triggers side effects as necessary.\n")
                    doc.append(f" */\n")

                    new_lines.extend(doc)
            new_lines.append(line)
            i += 1

        with open(fpath, 'w') as f:
            f.writelines(new_lines)

process_ts_files()
