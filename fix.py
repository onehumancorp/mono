with open("srcs/interop/types.go", "r") as f:
    content = f.read()

content = content.replace('if state.Data == nil {\n\t\tstate.Data = make(map[string]interface{})\n\t}', 'if state.Data == nil {\n\t\tstate.Data = make(map[string]interface{})\n\t\t_ = state.Data\n\t}')

with open("srcs/interop/types.go", "w") as f:
    f.write(content)
