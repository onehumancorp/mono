#!/bin/bash
set -e

# Janitor cleanup script to remove obsolete data-residue and prune superseded architectural blueprints

echo "Starting cleanup..."

# Remove temporary data residue files (Zero Junk policy)
rm -f backend.log frontend.log *.patch triage_report.html patch.diff

# Prune superseded blueprints from swarm_memory
# This safely deletes any architectural_blueprint_v* that is not the latest version (v3)
sqlite3 ohc.db "DELETE FROM swarm_memory WHERE key LIKE 'architectural_blueprint_v%' AND key != 'architectural_blueprint_v3';"

echo "Cleanup completed successfully."
