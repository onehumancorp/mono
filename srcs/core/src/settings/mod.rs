/// Settings management for the OHC core.
///
/// Handles loading, persisting, and watching configuration for both
/// single-docker deployments and cloud-native multi-tenant deployments.
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::path::PathBuf;
use std::sync::{Arc, RwLock};
use thiserror::Error;

#[derive(Debug, Error)]
pub enum SettingsError {
    #[error("I/O error: {0}")]
    Io(#[from] std::io::Error),
    #[error("JSON parse error: {0}")]
    Json(#[from] serde_json::Error),
    #[error("Setting key not found: {0}")]
    NotFound(String),
}

/// AI provider configuration.
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub struct AiProvider {
    pub name: String,
    pub api_key: Option<String>,
    pub base_url: Option<String>,
    pub model: String,
    pub enabled: bool,
}

/// Settings for a single OHC deployment (single-docker or desktop).
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AppSettings {
    /// Server listen address, e.g. "0.0.0.0:18789"
    pub listen_addr: String,
    /// Path to the SQLite database file (single-docker mode).
    pub db_path: Option<String>,
    /// PostgreSQL connection string (cloud-native mode).
    pub postgres_url: Option<String>,
    /// Redis connection string.
    pub redis_url: Option<String>,
    /// Configured AI providers.
    pub ai_providers: Vec<AiProvider>,
    /// Additional free-form key/value pairs.
    pub extras: HashMap<String, String>,
}

impl Default for AppSettings {
    fn default() -> Self {
        Self {
            listen_addr: "0.0.0.0:18789".to_string(),
            db_path: Some("ohc.db".to_string()),
            postgres_url: None,
            redis_url: None,
            ai_providers: vec![],
            extras: HashMap::new(),
        }
    }
}

/// Thread-safe settings store.
#[derive(Clone)]
pub struct SettingsStore {
    inner: Arc<RwLock<AppSettings>>,
    path: Option<PathBuf>,
}

impl SettingsStore {
    /// Create an in-memory store with default settings.
    pub fn new() -> Self {
        Self {
            inner: Arc::new(RwLock::new(AppSettings::default())),
            path: None,
        }
    }

    /// Load settings from a JSON file, falling back to defaults on error.
    pub fn from_file(path: impl Into<PathBuf>) -> Result<Self, SettingsError> {
        let path = path.into();
        let settings = if path.exists() {
            let data = std::fs::read_to_string(&path)?;
            serde_json::from_str(&data)?
        } else {
            AppSettings::default()
        };
        Ok(Self {
            inner: Arc::new(RwLock::new(settings)),
            path: Some(path),
        })
    }

    /// Persist current settings to the file (if a path is configured).
    pub fn save(&self) -> Result<(), SettingsError> {
        if let Some(ref path) = self.path {
            let settings = self.inner.read().unwrap().clone();
            let data = serde_json::to_string_pretty(&settings)?;
            std::fs::write(path, data)?;
        }
        Ok(())
    }

    /// Read a snapshot of the current settings.
    pub fn get(&self) -> AppSettings {
        self.inner.read().unwrap().clone()
    }

    /// Replace the entire settings object and optionally persist.
    pub fn set(&self, settings: AppSettings) -> Result<(), SettingsError> {
        *self.inner.write().unwrap() = settings;
        self.save()
    }

    /// Update a single extra key/value.
    pub fn set_extra(&self, key: impl Into<String>, value: impl Into<String>) -> Result<(), SettingsError> {
        let mut settings = self.inner.write().unwrap();
        settings.extras.insert(key.into(), value.into());
        drop(settings);
        self.save()
    }
}

impl Default for SettingsStore {
    fn default() -> Self {
        Self::new()
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn default_settings_listen_addr() {
        let store = SettingsStore::new();
        assert_eq!(store.get().listen_addr, "0.0.0.0:18789");
    }

    #[test]
    fn set_and_get_extra() {
        let store = SettingsStore::new();
        store.set_extra("theme", "dark").unwrap();
        assert_eq!(store.get().extras.get("theme").unwrap(), "dark");
    }

    #[test]
    fn roundtrip_json() {
        let store = SettingsStore::new();
        let mut s = store.get();
        s.extras.insert("key".to_string(), "val".to_string());
        let json = serde_json::to_string(&s).unwrap();
        let s2: AppSettings = serde_json::from_str(&json).unwrap();
        assert_eq!(s2.extras.get("key").unwrap(), "val");
    }
}
