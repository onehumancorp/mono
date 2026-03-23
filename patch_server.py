with open("srcs/frontend/server/BUILD.bazel", "r") as f:
    content = f.read()

content = content.replace('embed = [":server"]', 'embed = [":server_lib"]')

with open("srcs/frontend/server/BUILD.bazel", "w") as f:
    f.write(content)
