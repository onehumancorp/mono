import sqlite3
import json
import time
import os

db_path = os.path.expanduser('~/.openclaw/ohc.db')
os.makedirs(os.path.dirname(db_path), exist_ok=True)
conn = sqlite3.connect(db_path)
c = conn.cursor()

c.execute('''CREATE TABLE IF NOT EXISTS agent_missions (
    id TEXT PRIMARY KEY,
    role TEXT NOT NULL,
    task TEXT NOT NULL,
    status TEXT NOT NULL,
    assigned_to TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
)''')

backend_task = {
    "ID": "backend_mission_1",
    "Type": "event.task",
    "Content": "Implement the `capability_plugins` and `swarm_memory_embeddings` tables, and dynamic MCP registration as per the new Agentic OS blueprint."
}
c.execute("INSERT OR IGNORE INTO agent_missions (id, role, task, status) VALUES (?, ?, ?, ?)",
          ("backend_mission_1", "backend_dev", json.dumps(backend_task), "PENDING"))

ui_task = {
    "ID": "ui_mission_1",
    "Type": "event.task",
    "Content": "Update the OHC Next.js dashboard with Glassmorphism tokens (`blur(15px)`, `rgba` backgrounds, smooth data transitions)."
}
c.execute("INSERT OR IGNORE INTO agent_missions (id, role, task, status) VALUES (?, ?, ?, ?)",
          ("ui_mission_1", "ui_dev", json.dumps(ui_task), "PENDING"))

conn.commit()
conn.close()
print("Database seeded.")
