import re

def process_file(filepath):
    with open(filepath, 'r') as f:
        content = f.read()

    # Remove trailing commas from the JSON strings created previously
    # It matches `",\n\t\t}` and replaces with `"\n\t\t}`
    content = re.sub(r'",(\s*)\}\`\)', r'"\1}`)', content)

    with open(filepath, 'w') as f:
        f.write(content)

process_file('srcs/dashboard/server_missing_test.go')
