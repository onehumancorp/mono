import sqlite3
import os
import json
import uuid
from datetime import datetime, UTC

def init_db():
    db_dir = os.path.expanduser('~/.openclaw')
    os.makedirs(db_dir, exist_ok=True)
    db_path = os.path.join(db_dir, 'ohc.db')

    conn = sqlite3.connect(db_path)
    cursor = conn.cursor()

    # Core SIP tables
    cursor.execute("CREATE TABLE IF NOT EXISTS swarm_memory (key TEXT PRIMARY KEY, value TEXT, updated_at DATETIME);")
    cursor.execute("CREATE TABLE IF NOT EXISTS agent_status (agent_id TEXT PRIMARY KEY, role TEXT, status TEXT, last_heartbeat DATETIME);")
    cursor.execute("CREATE TABLE IF NOT EXISTS agent_missions (id TEXT PRIMARY KEY, role TEXT, task TEXT, status TEXT, assigned_to TEXT, created_at DATETIME, updated_at DATETIME);")

    # New Architectural schemas
    cursor.execute("CREATE TABLE IF NOT EXISTS capability_plugins (plugin_id TEXT PRIMARY KEY, name TEXT NOT NULL, version TEXT NOT NULL, manifest_url TEXT NOT NULL, status TEXT NOT NULL, registered_at DATETIME DEFAULT CURRENT_TIMESTAMP);")
    cursor.execute("CREATE TABLE IF NOT EXISTS swarm_memory_embeddings (memory_id TEXT PRIMARY KEY, context TEXT NOT NULL, vector_embedding BLOB, source_plugin TEXT, created_at DATETIME DEFAULT CURRENT_TIMESTAMP);")

    # Seed architectural state
    now = datetime.now(UTC).strftime('%Y-%m-%d %H:%M:%S')
    state = json.dumps({
        "vision": "Modular Plugin-Based Capability System",
        "aesthetics": "Next-Generation OHC Design System Tokens",
        "roadmap_status": "Updated to deprecate static Skill Blueprints for dynamic Capability Plugin Mesh."
    })
    cursor.execute("INSERT INTO swarm_memory (key, value, updated_at) VALUES (?, ?, ?) ON CONFLICT(key) DO UPDATE SET value=excluded.value, updated_at=excluded.updated_at;", ('architectural_state', state, now))

    # Seed agent status
    cursor.execute("INSERT INTO agent_status (agent_id, role, status, last_heartbeat) VALUES (?, ?, ?, ?) ON CONFLICT(agent_id) DO UPDATE SET status=excluded.status, last_heartbeat=excluded.last_heartbeat;", ('antigravity_l7', 'Principal Product Architect', 'Architectural blueprint updated. Handoff missions created.', now))

    # Helper to insert a mission
    def insert_mission(role, recipient_id, content, metadata):
        msg_id = str(uuid.uuid4())
        mission_id = str(uuid.uuid4())
        task_json = json.dumps({
            "id": msg_id,
            "sender_id": "antigravity_l7",
            "recipient_id": recipient_id,
            "content": content,
            "metadata": metadata
        })
        cursor.execute("INSERT INTO agent_missions (id, role, task, status, assigned_to, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)", (mission_id, role, task_json, 'READY', recipient_id, now, now))

    # Insert missions if none exist
    cursor.execute("SELECT COUNT(*) FROM agent_missions")
    if cursor.fetchone()[0] == 0:
        insert_mission("SWE Agent (Backend)", "backend_dev", "Implement the capability_plugins and swarm_memory_embeddings tables, and dynamic MCP registration as per the new Agentic OS blueprint.", {"task_id": "3.1", "priority": "high", "epic": "3"})
        insert_mission("UI Developer Agent", "ui_dev", "Update the OHC Next.js dashboard with Glassmorphism tokens (blur(15px), rgba backgrounds, smooth data transitions).", {"task_id": "3.2", "priority": "high", "epic": "3"})
        insert_mission("Visualizer Agent", "visualizer", "Generate high-fidelity mockups of the new Capability Dashboard and plugin mesh integration to serve as a ground-truth reference for frontend implementation.", {"task_id": "3.3", "priority": "medium", "epic": "3"})
        print("Missions inserted.")

    conn.commit()
    conn.close()
    print("OHC DB initialized successfully.")

if __name__ == '__main__':
    init_db()
