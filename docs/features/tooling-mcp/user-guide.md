# User Guide: MCP Tool Integrations

## Introduction
MCP (Model Context Protocol) is how your agents "interact with the world". It connects them to tools like GitHub, Slack, and your internal databases.

## Setup
### 1. Browse Integrations
Open the "Marketplace" and filter for "Tools".

### 2. Connect a Tool
Click "Connect" and enter the MCP server URL. For popular tools, we provide pre-configured templates.

### 3. Permissions
Assign the tool to specific agents. Do not give "Delete Database" access to a junior intern agent!

## Advanced Usage
### Building Custom Tools
You can build your own MCP server in any language. Just provide a manifest file that OHC can read.

## Troubleshooting
**Tool is disconnected**
- Check the "Gateway" status icon.
- Ensure the MCP server is running on a reachable network path.
