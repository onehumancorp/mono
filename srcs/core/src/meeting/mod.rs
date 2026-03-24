/// Meeting room management for the OHC core.
///
/// Provides creation, participant management, and lifecycle tracking for
/// virtual meeting rooms.  Rooms are scoped per organisation so that data
/// is isolated between tenants.
use async_trait::async_trait;
use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::sync::Mutex;
use thiserror::Error;
use uuid::Uuid;

#[derive(Debug, Error)]
pub enum MeetingError {
    #[error("Room not found: {0}")]
    NotFound(String),
    #[error("Participant already joined: {0}")]
    AlreadyJoined(String),
    #[error("Meeting error: {0}")]
    Internal(String),
}

/// Whether a meeting is currently active or has ended.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum RoomStatus {
    Active,
    Closed,
}

/// A single participant in a meeting room.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Participant {
    pub id: String,
    pub name: String,
    /// `true` for human users, `false` for AI agents.
    pub is_agent: bool,
    pub joined_at: DateTime<Utc>,
}

/// A virtual meeting room.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MeetingRoom {
    pub id: String,
    /// Organisation that owns this room (multi-tenant key).
    pub organization_id: String,
    pub name: String,
    pub status: RoomStatus,
    pub participants: Vec<Participant>,
    pub created_at: DateTime<Utc>,
    pub closed_at: Option<DateTime<Utc>>,
}

impl MeetingRoom {
    pub fn new(org_id: impl Into<String>, name: impl Into<String>) -> Self {
        Self {
            id: Uuid::new_v4().to_string(),
            organization_id: org_id.into(),
            name: name.into(),
            status: RoomStatus::Active,
            participants: vec![],
            created_at: Utc::now(),
            closed_at: None,
        }
    }
}

/// Storage backend for meeting rooms.
#[async_trait]
pub trait RoomStore: Send + Sync {
    async fn create(&self, room: MeetingRoom) -> Result<MeetingRoom, MeetingError>;
    async fn get(&self, id: &str) -> Result<MeetingRoom, MeetingError>;
    async fn list(&self, org_id: &str) -> Result<Vec<MeetingRoom>, MeetingError>;
    async fn update(&self, room: MeetingRoom) -> Result<(), MeetingError>;
}

/// In-memory room store (single-docker / desktop mode).
#[derive(Default)]
pub struct InMemoryRoomStore {
    rooms: Mutex<HashMap<String, MeetingRoom>>,
}

#[async_trait]
impl RoomStore for InMemoryRoomStore {
    async fn create(&self, room: MeetingRoom) -> Result<MeetingRoom, MeetingError> {
        let mut rooms = self.rooms.lock().unwrap();
        rooms.insert(room.id.clone(), room.clone());
        Ok(room)
    }

    async fn get(&self, id: &str) -> Result<MeetingRoom, MeetingError> {
        let rooms = self.rooms.lock().unwrap();
        rooms.get(id).cloned().ok_or_else(|| MeetingError::NotFound(id.to_string()))
    }

    async fn list(&self, org_id: &str) -> Result<Vec<MeetingRoom>, MeetingError> {
        let rooms = self.rooms.lock().unwrap();
        Ok(rooms.values().filter(|r| r.organization_id == org_id).cloned().collect())
    }

    async fn update(&self, room: MeetingRoom) -> Result<(), MeetingError> {
        let mut rooms = self.rooms.lock().unwrap();
        rooms.insert(room.id.clone(), room);
        Ok(())
    }
}

/// High-level meeting room manager.
pub struct MeetingManager<S: RoomStore> {
    store: S,
}

impl<S: RoomStore> MeetingManager<S> {
    pub fn new(store: S) -> Self {
        Self { store }
    }

    /// Open a new meeting room for an organisation.
    pub async fn open_room(
        &self,
        org_id: impl Into<String>,
        name: impl Into<String>,
    ) -> Result<MeetingRoom, MeetingError> {
        let room = MeetingRoom::new(org_id, name);
        self.store.create(room).await
    }

    /// Add a participant to an active room.
    pub async fn join(
        &self,
        room_id: &str,
        participant_id: impl Into<String>,
        name: impl Into<String>,
        is_agent: bool,
    ) -> Result<MeetingRoom, MeetingError> {
        let mut room = self.store.get(room_id).await?;
        let pid = participant_id.into();
        if room.participants.iter().any(|p| p.id == pid) {
            return Err(MeetingError::AlreadyJoined(pid));
        }
        room.participants.push(Participant {
            id: pid,
            name: name.into(),
            is_agent,
            joined_at: Utc::now(),
        });
        self.store.update(room.clone()).await?;
        Ok(room)
    }

    /// Close a meeting room.
    pub async fn close_room(&self, room_id: &str) -> Result<(), MeetingError> {
        let mut room = self.store.get(room_id).await?;
        room.status = RoomStatus::Closed;
        room.closed_at = Some(Utc::now());
        self.store.update(room).await
    }

    /// List all rooms for an organisation (tenant-scoped).
    pub async fn list_for_org(&self, org_id: &str) -> Result<Vec<MeetingRoom>, MeetingError> {
        self.store.list(org_id).await
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[tokio::test]
    async fn open_join_close() {
        let mgr = MeetingManager::new(InMemoryRoomStore::default());
        let room = mgr.open_room("org-1", "standup").await.unwrap();
        assert_eq!(room.status, RoomStatus::Active);

        let room = mgr.join(&room.id, "user-1", "Alice", false).await.unwrap();
        assert_eq!(room.participants.len(), 1);

        mgr.close_room(&room.id).await.unwrap();
        let rooms = mgr.list_for_org("org-1").await.unwrap();
        assert_eq!(rooms[0].status, RoomStatus::Closed);
    }

    #[tokio::test]
    async fn tenant_isolation() {
        let mgr = MeetingManager::new(InMemoryRoomStore::default());
        mgr.open_room("org-a", "room-a").await.unwrap();
        mgr.open_room("org-b", "room-b").await.unwrap();

        assert_eq!(mgr.list_for_org("org-a").await.unwrap().len(), 1);
        assert_eq!(mgr.list_for_org("org-b").await.unwrap().len(), 1);
    }
}
