/// Agent scheduler for the OHC core.
///
/// The scheduler decides *when* and *how often* agents run.  It supports
/// one-shot tasks, periodic (cron-like) tasks, and event-driven triggers.
/// For single-docker deployments the scheduler runs inside the same Tokio
/// runtime.  For cloud-native deployments the scheduler emits work items
/// that are picked up by the Kubernetes-based orchestration layer.
use async_trait::async_trait;
use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::sync::{Arc, Mutex};
use thiserror::Error;
use uuid::Uuid;

#[derive(Debug, Error)]
pub enum SchedulerError {
    #[error("Task not found: {0}")]
    NotFound(String),
    #[error("Task already exists: {0}")]
    AlreadyExists(String),
    #[error("Scheduler error: {0}")]
    Internal(String),
}

/// How frequently a scheduled task should fire.
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
#[serde(rename_all = "snake_case")]
pub enum Schedule {
    /// Run exactly once at the given UTC time.
    Once { at: DateTime<Utc> },
    /// Run every `seconds` seconds.
    Interval { seconds: u64 },
    /// Cron expression (standard 5-field syntax).
    Cron { expression: String },
}

/// Current state of a scheduled task.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum TaskStatus {
    Pending,
    Running,
    Succeeded,
    Failed,
    Cancelled,
}

/// A single item of work to be executed by an agent.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ScheduledTask {
    pub id: String,
    /// Organisation that owns this task (multi-tenant key).
    pub organization_id: String,
    /// ID of the agent that should handle the task.
    pub agent_id: String,
    pub name: String,
    pub schedule: Schedule,
    pub status: TaskStatus,
    pub created_at: DateTime<Utc>,
    pub last_run_at: Option<DateTime<Utc>>,
    pub next_run_at: Option<DateTime<Utc>>,
    /// Arbitrary payload forwarded to the agent.
    pub payload: serde_json::Value,
}

impl ScheduledTask {
    pub fn new(
        org_id: impl Into<String>,
        agent_id: impl Into<String>,
        name: impl Into<String>,
        schedule: Schedule,
        payload: serde_json::Value,
    ) -> Self {
        let now = Utc::now();
        let next_run_at = match &schedule {
            Schedule::Once { at } => Some(*at),
            Schedule::Interval { .. } => Some(now),
            Schedule::Cron { .. } => None, // resolved by cron engine
        };
        Self {
            id: Uuid::new_v4().to_string(),
            organization_id: org_id.into(),
            agent_id: agent_id.into(),
            name: name.into(),
            schedule,
            status: TaskStatus::Pending,
            created_at: now,
            last_run_at: None,
            next_run_at,
            payload,
        }
    }
}

/// Storage backend for scheduled tasks.
#[async_trait]
pub trait TaskStore: Send + Sync {
    async fn create(&self, task: ScheduledTask) -> Result<ScheduledTask, SchedulerError>;
    async fn get(&self, id: &str) -> Result<ScheduledTask, SchedulerError>;
    async fn list(&self, org_id: &str) -> Result<Vec<ScheduledTask>, SchedulerError>;
    async fn update(&self, task: ScheduledTask) -> Result<(), SchedulerError>;
    async fn delete(&self, id: &str) -> Result<(), SchedulerError>;
    /// Return all pending tasks whose `next_run_at` is in the past.
    async fn due_tasks(&self) -> Result<Vec<ScheduledTask>, SchedulerError>;
}

/// In-memory task store for single-docker / desktop mode.
pub struct InMemoryTaskStore {
    tasks: Mutex<HashMap<String, ScheduledTask>>,
}

impl InMemoryTaskStore {
    pub fn new() -> Arc<Self> {
        Arc::new(Self {
            tasks: Mutex::new(HashMap::new()),
        })
    }
}

impl Default for InMemoryTaskStore {
    fn default() -> Self {
        Self {
            tasks: Mutex::new(HashMap::new()),
        }
    }
}

#[async_trait]
impl TaskStore for InMemoryTaskStore {
    async fn create(&self, task: ScheduledTask) -> Result<ScheduledTask, SchedulerError> {
        let mut tasks = self.tasks.lock().unwrap();
        if tasks.contains_key(&task.id) {
            return Err(SchedulerError::AlreadyExists(task.id));
        }
        tasks.insert(task.id.clone(), task.clone());
        Ok(task)
    }

    async fn get(&self, id: &str) -> Result<ScheduledTask, SchedulerError> {
        let tasks = self.tasks.lock().unwrap();
        tasks.get(id).cloned().ok_or_else(|| SchedulerError::NotFound(id.to_string()))
    }

    async fn list(&self, org_id: &str) -> Result<Vec<ScheduledTask>, SchedulerError> {
        let tasks = self.tasks.lock().unwrap();
        Ok(tasks.values().filter(|t| t.organization_id == org_id).cloned().collect())
    }

    async fn update(&self, task: ScheduledTask) -> Result<(), SchedulerError> {
        let mut tasks = self.tasks.lock().unwrap();
        tasks.insert(task.id.clone(), task);
        Ok(())
    }

    async fn delete(&self, id: &str) -> Result<(), SchedulerError> {
        let mut tasks = self.tasks.lock().unwrap();
        tasks.remove(id).ok_or_else(|| SchedulerError::NotFound(id.to_string()))?;
        Ok(())
    }

    async fn due_tasks(&self) -> Result<Vec<ScheduledTask>, SchedulerError> {
        let now = Utc::now();
        let tasks = self.tasks.lock().unwrap();
        Ok(tasks
            .values()
            .filter(|t| {
                t.status == TaskStatus::Pending
                    && t.next_run_at.map_or(false, |next| next <= now)
            })
            .cloned()
            .collect())
    }
}

/// High-level scheduler that manages task lifecycle.
pub struct Scheduler<S: TaskStore> {
    store: Arc<S>,
}

impl<S: TaskStore> Scheduler<S> {
    pub fn new(store: Arc<S>) -> Self {
        Self { store }
    }

    /// Schedule a new task for an organisation.
    pub async fn schedule(
        &self,
        org_id: impl Into<String>,
        agent_id: impl Into<String>,
        name: impl Into<String>,
        schedule: Schedule,
        payload: serde_json::Value,
    ) -> Result<ScheduledTask, SchedulerError> {
        let task = ScheduledTask::new(org_id, agent_id, name, schedule, payload);
        self.store.create(task).await
    }

    /// Cancel a scheduled task.
    pub async fn cancel(&self, task_id: &str) -> Result<(), SchedulerError> {
        let mut task = self.store.get(task_id).await?;
        task.status = TaskStatus::Cancelled;
        self.store.update(task).await
    }

    /// List all tasks for an organisation (tenant-scoped).
    pub async fn list_for_org(&self, org_id: &str) -> Result<Vec<ScheduledTask>, SchedulerError> {
        self.store.list(org_id).await
    }

    /// Return tasks that are ready to be executed right now.
    pub async fn poll_due(&self) -> Result<Vec<ScheduledTask>, SchedulerError> {
        self.store.due_tasks().await
    }

    /// Mark a task as running and return it.
    pub async fn mark_running(&self, task_id: &str) -> Result<ScheduledTask, SchedulerError> {
        let mut task = self.store.get(task_id).await?;
        task.status = TaskStatus::Running;
        task.last_run_at = Some(Utc::now());
        self.store.update(task.clone()).await?;
        Ok(task)
    }

    /// Mark a task as succeeded and compute its next run time if recurring.
    pub async fn mark_done(&self, task_id: &str, success: bool) -> Result<(), SchedulerError> {
        let mut task = self.store.get(task_id).await?;
        task.status = if success { TaskStatus::Succeeded } else { TaskStatus::Failed };
        if success {
            // Reschedule interval tasks.
            if let Schedule::Interval { seconds } = &task.schedule {
                let next = Utc::now() + chrono::Duration::seconds(*seconds as i64);
                task.next_run_at = Some(next);
                task.status = TaskStatus::Pending;
            }
        }
        self.store.update(task).await
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[tokio::test]
    async fn schedule_and_cancel() {
        let store = InMemoryTaskStore::new();
        let scheduler = Scheduler::new(store);
        let task = scheduler
            .schedule(
                "org-1",
                "agent-1",
                "daily-report",
                Schedule::Interval { seconds: 86400 },
                serde_json::json!({"report": "daily"}),
            )
            .await
            .unwrap();

        assert_eq!(task.status, TaskStatus::Pending);

        // Cancel and verify.
        scheduler.cancel(&task.id).await.unwrap();
        let tasks = scheduler.list_for_org("org-1").await.unwrap();
        assert_eq!(tasks[0].status, TaskStatus::Cancelled);
    }

    #[tokio::test]
    async fn tenant_isolation() {
        let store = InMemoryTaskStore::new();
        let scheduler = Scheduler::new(store);
        scheduler
            .schedule("org-a", "a1", "task-a", Schedule::Interval { seconds: 60 }, serde_json::Value::Null)
            .await
            .unwrap();
        scheduler
            .schedule("org-b", "b1", "task-b", Schedule::Interval { seconds: 60 }, serde_json::Value::Null)
            .await
            .unwrap();

        let a_tasks = scheduler.list_for_org("org-a").await.unwrap();
        let b_tasks = scheduler.list_for_org("org-b").await.unwrap();
        assert_eq!(a_tasks.len(), 1);
        assert_eq!(b_tasks.len(), 1);
        assert_ne!(a_tasks[0].id, b_tasks[0].id);
    }
}
