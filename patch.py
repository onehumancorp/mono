import re
import os

for root, _, files in os.walk("srcs/orchestration"):
    for file in files:
        if file.endswith(".go"):
            path = os.path.join(root, file)
            with open(path, "r") as f:
                text = f.read()

            text = text.replace("time.Now(.Unix())", "time.Now().Unix()")
            text = text.replace("time.Now(.UnixNano())", "time.Now().UnixNano()")

            with open(path, "w") as f:
                f.write(text)
