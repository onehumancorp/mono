import os
import re

def main():
    for root, _, files in os.walk("srcs"):
        for file in files:
            if not file.endswith(".go"):
                continue

            filepath = os.path.join(root, file)
            with open(filepath, "r") as f:
                content = f.read()

            # We need to preserve the leading whitespace for the inserted line
            # The regex matches the indent (\s*) and then "if err := json.NewDecoder(r.Body).Decode("
            # It replaces it with the indent + "r.Body = http.MaxBytesReader(w, r.Body, 1<<20)\n" + indent + "if err := json.NewDecoder(r.Body).Decode("

            # Use re.sub with a function to dynamically construct the replacement with the right indent
            def replacer(match):
                indent = match.group(1)
                return f"{indent}r.Body = http.MaxBytesReader(w, r.Body, 1<<20)\n{indent}if err := json.NewDecoder(r.Body).Decode("

            new_content = re.sub(r'(^[ \t]*)if err := json\.NewDecoder\(r\.Body\)\.Decode\(', replacer, content, flags=re.MULTILINE)

            if new_content != content:
                print(f"Modifying {filepath}")
                with open(filepath, "w") as f:
                    f.write(new_content)

if __name__ == "__main__":
    main()
