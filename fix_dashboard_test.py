import re

with open("srcs/dashboard/server_missing_test.go", "r") as f:
    content = f.read()

# Replace the first TestHandleScale
content = content.replace("func TestHandleScale(t *testing.T)", "func TestHandleScale1(t *testing.T)", 1)
content = content.replace("func TestHandleScaleStream(t *testing.T)", "func TestHandleScaleStream1(t *testing.T)", 1)

with open("srcs/dashboard/server_missing_test.go", "w") as f:
    f.write(content)
