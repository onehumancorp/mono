with open("srcs/frontend/src/App.test.tsx", "r") as f:
    lines = f.readlines()
for i, l in enumerate(lines[-50:]):
    print(f"{len(lines)-50+i+1}: {l.rstrip()}")
