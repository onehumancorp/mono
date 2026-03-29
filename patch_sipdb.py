with open("srcs/orchestration/sip.go", "r") as f:
    lines = f.readlines()

for i, line in enumerate(lines):
    if 'db, _ := NewSIPDB("/home/jules/.openclaw/ohc.db")' in line:
        lines[i] = '	home, _ := os.UserHomeDir()\n	dbPath := filepath.Join(home, ".openclaw", "ohc.db")\n	db, _ := NewSIPDB(dbPath)\n'

with open("srcs/orchestration/sip.go", "w") as f:
    f.writelines(lines)
