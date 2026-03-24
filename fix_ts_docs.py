import os
import re

def process_ts_file(filepath):
    with open(filepath, 'r', encoding='utf-8') as f:
        lines = f.readlines()

    modified = False
    new_lines = []

    for line in lines:
        stripped = line.strip()

        if stripped.startswith('* Summary:'):
            text = stripped.replace('* Summary:', '').strip()
            idx = line.find('*')
            indent = line[:idx]
            new_lines.append(f"{indent}* @summary {text}\n")
            modified = True

        elif stripped.startswith('* Parameters:'):
            text = stripped.replace('* Parameters:', '').strip()
            idx = line.find('*')
            indent = line[:idx]
            if text.lower() == 'none':
                new_lines.append(f"{indent}* @param None\n")
            else:
                params = [p.strip() for p in text.split(',')]
                for p in params:
                    if p:
                        new_lines.append(f"{indent}* @param {p}\n")
            modified = True

        elif stripped.startswith('* Returns:'):
            text = stripped.replace('* Returns:', '').strip()
            idx = line.find('*')
            indent = line[:idx]
            new_lines.append(f"{indent}* @returns {text}\n")
            modified = True

        elif stripped.startswith('* Errors:'):
            text = stripped.replace('* Errors:', '').strip()
            idx = line.find('*')
            indent = line[:idx]
            new_lines.append(f"{indent}* @throws {text}\n")
            modified = True

        elif stripped.startswith('* Side Effects:'):
            text = stripped.replace('* Side Effects:', '').strip()
            idx = line.find('*')
            indent = line[:idx]
            new_lines.append(f"{indent}* @remarks Side Effects: {text}\n")
            modified = True

        else:
            new_lines.append(line)

    if modified:
        with open(filepath, 'w', encoding='utf-8') as f:
            f.writelines(new_lines)

def main():
    for root, dirs, files in os.walk('./srcs/frontend'):
        if 'node_modules' in root or '.git' in root or 'dist' in root:
            continue
        for file in files:
            if file.endswith('.ts') or file.endswith('.tsx'):
                process_ts_file(os.path.join(root, file))

if __name__ == "__main__":
    main()
