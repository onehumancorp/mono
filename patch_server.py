with open("srcs/dashboard/server.go", "r") as f:
    lines = f.readlines()

for i, line in enumerate(lines):
    if "sipdb:" in line and "sipdb," in line:
        lines[i] = "			sipdb:                 orchestration.GetDefaultSIPDB(),\n"

with open("srcs/dashboard/server.go", "w") as f:
    f.writelines(lines)
