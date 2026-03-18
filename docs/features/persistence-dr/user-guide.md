# User Guide: Persistence & Snapshots

## Introduction
The Snapshot Fabric allows you to "Undo" complex organizational changes or recover from a disaster with a single click.

## Usage
### 1. Creating a Snapshot
Before performing a major reorganisation (e.g., merging two departments), click "Create Snapshot" in the Mission Control panel.

### 2. Labeling
Always give your snapshots a descriptive label (e.g. `pre-frontend-refactor`).

### 3. Restoring
If something goes wrong, navigate to the Snapshots log and click "Restore". Your organisation will revert to its previous state within 5 seconds.

## Best Practices
- Create automated "Weekly Snapshots" in the settings.
- Use snapshots to "Test Scenarios" without affecting your long-term production state.

## Troubleshooting
**Snapshot restoration failed**
- Check the database logs for the OHC cluster.
- Ensure you have enough storage space in your Kubernetes cluster.
