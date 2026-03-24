/// Chat application integration for the OHC core.
///
/// Provides an abstraction over multiple chat back-ends (Chatwoot, Slack,
/// Telegram, Discord, …).  The unified API makes it trivial to add new
/// adapters without touching the rest of the system.  All messages are
/// scoped to an organisation so data never leaks between tenants.
use async_trait::async_trait;
use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::sync::Mutex;
use thiserror::Error;
use uuid::Uuid;

#[derive(Debug, Error)]
pub enum ChatError {
    #[error("Channel not found: {0}")]
    NotFound(String),
    #[error("Send failed: {0}")]
    SendFailed(String),
    #[error("Chat error: {0}")]
    Internal(String),
}

/// Supported chat integration backends.
#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum ChatBackend {
    Chatwoot,
    Slack,
    Telegram,
    Discord,
    Teams,
    Mattermost,
    /// Generic webhook backend.
    Webhook { url: String },
}

/// A single chat channel (e.g. a Slack channel, a Telegram group).
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ChatChannel {
    pub id: String,
    /// Organisation that owns this channel (multi-tenant key).
    pub organization_id: String,
    pub name: String,
    pub backend: ChatBackend,
    /// Backend-specific configuration (token, room ID, etc.).
    pub config: HashMap<String, String>,
    pub enabled: bool,
    pub created_at: DateTime<Utc>,
}

impl ChatChannel {
    pub fn new(
        org_id: impl Into<String>,
        name: impl Into<String>,
        backend: ChatBackend,
    ) -> Self {
        Self {
            id: Uuid::new_v4().to_string(),
            organization_id: org_id.into(),
            name: name.into(),
            backend,
            config: HashMap::new(),
            enabled: true,
            created_at: Utc::now(),
        }
    }
}

/// A message sent or received over a chat channel.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ChatMessage {
    pub id: String,
    pub channel_id: String,
    /// Organisation that owns this message (multi-tenant key).
    pub organization_id: String,
    pub author_id: String,
    pub author_name: String,
    pub body: String,
    pub sent_at: DateTime<Utc>,
}

impl ChatMessage {
    pub fn new(
        channel_id: impl Into<String>,
        org_id: impl Into<String>,
        author_id: impl Into<String>,
        author_name: impl Into<String>,
        body: impl Into<String>,
    ) -> Self {
        Self {
            id: Uuid::new_v4().to_string(),
            channel_id: channel_id.into(),
            organization_id: org_id.into(),
            author_id: author_id.into(),
            author_name: author_name.into(),
            body: body.into(),
            sent_at: Utc::now(),
        }
    }
}

/// Abstract transport used by the chat manager to dispatch messages.
#[async_trait]
pub trait ChatTransport: Send + Sync {
    async fn send(&self, channel: &ChatChannel, message: &ChatMessage) -> Result<(), ChatError>;
}

/// No-op transport for testing.
pub struct NoopTransport;

#[async_trait]
impl ChatTransport for NoopTransport {
    async fn send(&self, _channel: &ChatChannel, _message: &ChatMessage) -> Result<(), ChatError> {
        Ok(())
    }
}

/// Storage backend for channels and messages.
#[async_trait]
pub trait ChatStore: Send + Sync {
    async fn save_channel(&self, channel: ChatChannel) -> Result<ChatChannel, ChatError>;
    async fn get_channel(&self, id: &str) -> Result<ChatChannel, ChatError>;
    async fn list_channels(&self, org_id: &str) -> Result<Vec<ChatChannel>, ChatError>;
    async fn save_message(&self, message: ChatMessage) -> Result<ChatMessage, ChatError>;
    async fn list_messages(
        &self,
        channel_id: &str,
        org_id: &str,
        limit: usize,
    ) -> Result<Vec<ChatMessage>, ChatError>;
}

/// In-memory chat store for single-docker / desktop mode.
#[derive(Default)]
pub struct InMemoryChatStore {
    channels: Mutex<HashMap<String, ChatChannel>>,
    messages: Mutex<Vec<ChatMessage>>,
}

#[async_trait]
impl ChatStore for InMemoryChatStore {
    async fn save_channel(&self, channel: ChatChannel) -> Result<ChatChannel, ChatError> {
        let mut channels = self.channels.lock().unwrap();
        channels.insert(channel.id.clone(), channel.clone());
        Ok(channel)
    }

    async fn get_channel(&self, id: &str) -> Result<ChatChannel, ChatError> {
        let channels = self.channels.lock().unwrap();
        channels.get(id).cloned().ok_or_else(|| ChatError::NotFound(id.to_string()))
    }

    async fn list_channels(&self, org_id: &str) -> Result<Vec<ChatChannel>, ChatError> {
        let channels = self.channels.lock().unwrap();
        Ok(channels.values().filter(|c| c.organization_id == org_id).cloned().collect())
    }

    async fn save_message(&self, message: ChatMessage) -> Result<ChatMessage, ChatError> {
        let mut messages = self.messages.lock().unwrap();
        messages.push(message.clone());
        Ok(message)
    }

    async fn list_messages(
        &self,
        channel_id: &str,
        org_id: &str,
        limit: usize,
    ) -> Result<Vec<ChatMessage>, ChatError> {
        let messages = self.messages.lock().unwrap();
        let mut result: Vec<ChatMessage> = messages
            .iter()
            .filter(|m| m.channel_id == channel_id && m.organization_id == org_id)
            .cloned()
            .collect();
        result.sort_by(|a, b| a.sent_at.cmp(&b.sent_at));
        result.truncate(limit);
        Ok(result)
    }
}

/// High-level chat integration manager.
pub struct ChatManager<S: ChatStore, T: ChatTransport> {
    store: S,
    transport: T,
}

impl<S: ChatStore, T: ChatTransport> ChatManager<S, T> {
    pub fn new(store: S, transport: T) -> Self {
        Self { store, transport }
    }

    /// Register a new chat channel for an organisation.
    pub async fn add_channel(
        &self,
        org_id: impl Into<String>,
        name: impl Into<String>,
        backend: ChatBackend,
    ) -> Result<ChatChannel, ChatError> {
        let channel = ChatChannel::new(org_id, name, backend);
        self.store.save_channel(channel).await
    }

    /// Send a message through the appropriate backend transport.
    pub async fn send(
        &self,
        channel_id: &str,
        org_id: &str,
        author_id: impl Into<String>,
        author_name: impl Into<String>,
        body: impl Into<String>,
    ) -> Result<ChatMessage, ChatError> {
        let channel = self.store.get_channel(channel_id).await?;
        if channel.organization_id != org_id {
            return Err(ChatError::NotFound(channel_id.to_string()));
        }
        let message = ChatMessage::new(channel_id, org_id, author_id, author_name, body);
        self.transport.send(&channel, &message).await?;
        self.store.save_message(message).await
    }

    /// Retrieve recent messages for a channel (tenant-scoped).
    pub async fn messages(
        &self,
        channel_id: &str,
        org_id: &str,
        limit: usize,
    ) -> Result<Vec<ChatMessage>, ChatError> {
        self.store.list_messages(channel_id, org_id, limit).await
    }

    /// List all channels for an organisation.
    pub async fn list_channels(&self, org_id: &str) -> Result<Vec<ChatChannel>, ChatError> {
        self.store.list_channels(org_id).await
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    fn make_manager() -> ChatManager<InMemoryChatStore, NoopTransport> {
        ChatManager::new(InMemoryChatStore::default(), NoopTransport)
    }

    #[tokio::test]
    async fn add_channel_and_send() {
        let mgr = make_manager();
        let ch = mgr.add_channel("org-1", "general", ChatBackend::Slack).await.unwrap();
        let msg = mgr.send(&ch.id, "org-1", "user-1", "Alice", "hello").await.unwrap();
        assert_eq!(msg.body, "hello");
        let msgs = mgr.messages(&ch.id, "org-1", 10).await.unwrap();
        assert_eq!(msgs.len(), 1);
    }

    #[tokio::test]
    async fn cross_tenant_channel_access_denied() {
        let mgr = make_manager();
        let ch = mgr.add_channel("org-a", "secret", ChatBackend::Slack).await.unwrap();
        // org-b must not be able to send to org-a's channel.
        let err = mgr.send(&ch.id, "org-b", "user-2", "Eve", "hack").await;
        assert!(err.is_err());
    }
}
