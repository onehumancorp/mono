import os
import re

def process_ts_files():
    for root, _, files in os.walk('srcs/frontend/src'):
        for file in files:
            if file.endswith('.ts') or file.endswith('.tsx'):
                if file.endswith('.test.ts') or file.endswith('.test.tsx') or file.endswith('.spec.ts') or 'vite.config.ts' in file or 'vitest.config.ts' in file or 'playwright.config.ts' in file or 'setupTests.ts' in file or 'proto_types.ts' in file:
                    continue

                filepath = os.path.join(root, file)
                # Restore to make sure we don't duplicate
                os.system(f"git restore {filepath}")

                with open(filepath, 'r') as f:
                    lines = f.readlines()

                new_lines = []
                i = 0
                while i < len(lines):
                    line = lines[i]

                    # Match exported functions/classes/interfaces/types/consts
                    ts_match = re.match(r'^export\s+(?:async\s+)?(?:default\s+)?(function|class|interface|type|const|let|var)\s+([a-zA-Z0-9_]+)\s*(?:=\s*(?:async\s*)?(?:\((.*?)\)\s*=>)|(?:\((.*?)\)(?:\s*:\s*(.*?))?\s*\{)|(?:\s*:\s*(.*?)\s*=))?', line)

                    if ts_match:
                        kind = ts_match.group(1)
                        name = ts_match.group(2)

                        arrow_params = ts_match.group(3)
                        func_params = ts_match.group(4)
                        func_returns = ts_match.group(5)

                        params_str = arrow_params if arrow_params is not None else func_params
                        if params_str:
                            cleaned_params = []
                            for p in params_str.split(','):
                                p = p.split(':')[0].strip()
                                if p:
                                    cleaned_params.append(p)
                            params_str = ', '.join(cleaned_params)
                        else:
                            params_str = "None"

                        returns_str = func_returns if func_returns is not None else "None"
                        if returns_str and '{' in returns_str:
                             returns_str = returns_str.split('{')[0].strip()

                        is_func = kind == 'function' or (kind in ['const', 'let', 'var'] and ('=>' in line or 'function' in line))
                        p_str = params_str if is_func and params_str else "None"
                        r_str = returns_str if is_func and returns_str else "None"
                        e_str = "May throw an error" if is_func else "None"

                        # Find existing comments above the line to overwrite it
                        comment_block = []
                        j = i - 1
                        in_multiline = False

                        while j >= 0:
                            prev_line = lines[j].strip()
                            if not prev_line:
                                j -= 1
                                continue
                            if prev_line == '*/':
                                in_multiline = True
                                comment_block.insert(0, lines[j])
                            elif in_multiline:
                                comment_block.insert(0, lines[j])
                                if prev_line.startswith('/*') or prev_line.startswith('/**'):
                                    in_multiline = False
                                    break
                            elif prev_line.startswith('//'):
                                comment_block.insert(0, lines[j])
                            else:
                                break
                            j -= 1

                        # Extract the summary to preserve the original comments details
                        summary_lines = []
                        for c in comment_block:
                            c = c.strip()
                            c = re.sub(r'^\/\*\*?', '', c)
                            c = re.sub(r'^\*\/', '', c)
                            c = re.sub(r'^\*', '', c)
                            c = re.sub(r'^\/\/', '', c)
                            c = c.strip()
                            if c and not c.startswith('@') and not c.startswith('Summary:') and not c.startswith('Intent:') and not c.startswith('Params:') and not c.startswith('Parameters:') and not c.startswith('Returns:') and not c.startswith('Errors:') and not c.startswith('Side Effects:'):
                                summary_lines.append(c)

                        summary = ' '.join(summary_lines).strip()
                        if not summary:
                             summary = f"{name} functionality." if is_func else f"Defines {name}."

                        # remove the original comments lines from new_lines
                        # (calculate how many lines to remove from new_lines by counting backwards from i to j+1)
                        remove_count = i - (j + 1)
                        if remove_count > 0:
                            new_lines = new_lines[:-remove_count]

                        docstring = [
                            '/**\n',
                            f" * Summary: {summary}\n",
                            f" * Parameters: {p_str}\n",
                            f" * Returns: {r_str}\n",
                            f" * Errors: {e_str}\n",
                            f" * Side Effects: None\n",
                            ' */\n'
                        ]

                        # Ensure we don't duplicate `/**`
                        while len(new_lines) > 0 and '/**' in new_lines[-1]:
                            new_lines.pop()

                        new_lines.extend(docstring)
                        new_lines.append(line)
                        i += 1
                        continue

                    new_lines.append(line)
                    i += 1

                with open(filepath, 'w') as f:
                    f.writelines(new_lines)

process_ts_files()
