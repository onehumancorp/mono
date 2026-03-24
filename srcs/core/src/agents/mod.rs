/// Agent management for the OHC core.
///
/// Responsible for the lifecycle of AI agents: registration, status tracking,
/// capability discovery, and teardown.  Designed to work identically in a
/// single-docker deployment (in-process) and inside a Kubernetes cluster
/// (via the orchestration layer).
use async_trait::async_trait;
use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::sync::{Arc, Mutex};
use thiserror::Error;
use uuid::Uuid;

#[derive(Debug, Error)]
pub enum AgentError {
    #[error("Agent not found: {0}")]
    NotFound(String),
    #[error("Agent already exists: {0}")]
    AlreadyExists(String),
    #[error("Agent error: {0}")]
    Internal(String),
}

/// Lifecycle state of an agent.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum AgentStatus {
    Pending,
    Running,
    Paused,
    Completed,
    Failed,
}

/// Descriptor for a single AI agent.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AgentDescriptor {
    pub id: String,
    pub name: String,
    pub role: String,
    /// Organisation that owns this agent (multi-tenant key).
    pub organization_id: String,
    pub status: AgentStatus,
    pub capabilities: Vec<String>,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
    /// Free-form metadata (model name, API key ref, etc.).
    pub metadata: HashMap<String, String>,
}

impl AgentDescriptor {
    pub fn new(name: impl Into<String>, role: impl Into<String>, org_id: impl Into<String>) -> Self {
        let now = Utc::now();
        Self {
            id: Uuid::new_v4().to_string(),
            name: name.into(),
            role: role.into(),
            organization_id: org_id.into(),
            status: AgentStatus::Pending,
            capabilities: vec![],
            created_at: now,
            updated_at: now,
            metadata: HashMap::new(),
        }
    }
}

/// Trait for agent storage backends (in-memory, SQLite, PostgreSQL…).
#[async_trait]
pub trait AgentStore: Send + Sync {
    async fn create(&self, agent: AgentDescriptor) -> Result<AgentDescriptor, AgentError>;
    async fn get(&self, id: &str) -> Result<AgentDescriptor, AgentError>;
    async fn list(&self, org_id: &str) -> Result<Vec<AgentDescriptor>, AgentError>;
    async fn update_status(&self, id: &str, status: AgentStatus) -> Result<(), AgentError>;
    async fn delete(&self, id: &str) -> Result<(), AgentError>;
}

/// In-memory agent store — sufficient for single-docker / desktop mode.
pub struct InMemoryAgentStore {
    agents: Mutex<HashMap<String, AgentDescriptor>>,
}

impl InMemoryAgentStore {
    pub fn new() -> Arc<Self> {
        Arc::new(Self {
            agents: Mutex::new(HashMap::new()),
        })
    }
}

impl Default for InMemoryAgentStore {
    fn default() -> Self {
        Self {
            agents: Mutex::new(HashMap::new()),
        }
    }
}

#[async_trait]
impl AgentStore for InMemoryAgentStore {
    async fn create(&self, agent: AgentDescriptor) -> Result<AgentDescriptor, AgentError> {
        let mut agents = self.agents.lock().unwrap();
        if agents.contains_key(&agent.id) {
            return Err(AgentError::AlreadyExists(agent.id));
        }
        agents.insert(agent.id.clone(), agent.clone());
        Ok(agent)
    }

    async fn get(&self, id: &str) -> Result<AgentDescriptor, AgentError> {
        let agents = self.agents.lock().unwrap();
        agents.get(id).cloned().ok_or_else(|| AgentError::NotFound(id.to_string()))
    }

    async fn list(&self, org_id: &str) -> Result<Vec<AgentDescriptor>, AgentError> {
        let agents = self.agents.lock().unwrap();
        Ok(agents.values().filter(|a| a.organization_id == org_id).cloned().collect())
    }

    async fn update_status(&self, id: &str, status: AgentStatus) -> Result<(), AgentError> {
        let mut agents = self.agents.lock().unwrap();
        let agent = agents.get_mut(id).ok_or_else(|| AgentError::NotFound(id.to_string()))?;
        agent.status = status;
        agent.updated_at = Utc::now();
        Ok(())
    }

    async fn delete(&self, id: &str) -> Result<(), AgentError> {
        let mut agents = self.agents.lock().unwrap();
        agents.remove(id).ok_or_else(|| AgentError::NotFound(id.to_string()))?;
        Ok(())
    }
}

/// High-level agent manager used by both the desktop app and the backend server.
pub struct AgentManager<S: AgentStore> {
    store: Arc<S>,
}

impl<S: AgentStore> AgentManager<S> {
    pub fn new(store: Arc<S>) -> Self {
        Self { store }
    }

    /// Hire (register) an agent for an organisation.
    pub async fn hire(
        &self,
        name: impl Into<String>,
        role: impl Into<String>,
        org_id: impl Into<String>,
    ) -> Result<AgentDescriptor, AgentError> {
        let agent = AgentDescriptor::new(name, role, org_id);
        self.store.create(agent).await
    }

    /// Fire (remove) an agent.
    pub async fn fire(&self, agent_id: &str) -> Result<(), AgentError> {
        self.store.delete(agent_id).await
    }

    /// List all agents for an organisation (tenant-scoped).
    pub async fn list_for_org(&self, org_id: &str) -> Result<Vec<AgentDescriptor>, AgentError> {
        self.store.list(org_id).await
    }

    pub async fn get(&self, agent_id: &str) -> Result<AgentDescriptor, AgentError> {
        self.store.get(agent_id).await
    }

    pub async fn set_status(&self, agent_id: &str, status: AgentStatus) -> Result<(), AgentError> {
        self.store.update_status(agent_id, status).await
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[tokio::test]
    async fn hire_and_list() {
        let store = InMemoryAgentStore::new();
        let mgr = AgentManager::new(store);
        let agent = mgr.hire("Alice", "engineer", "org-1").await.unwrap();
        assert_eq!(agent.status, AgentStatus::Pending);

        let agents = mgr.list_for_org("org-1").await.unwrap();
        assert_eq!(agents.len(), 1);

        // Agents from a different org are not visible.
        let agents_org2 = mgr.list_for_org("org-2").await.unwrap();
        assert!(agents_org2.is_empty());
    }

    #[tokio::test]
    async fn fire_agent() {
        let store = InMemoryAgentStore::new();
        let mgr = AgentManager::new(store);
        let agent = mgr.hire("Bob", "ceo", "org-1").await.unwrap();
        mgr.fire(&agent.id).await.unwrap();
        assert!(mgr.get(&agent.id).await.is_err());
    }
}
