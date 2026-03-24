use serde::{Deserialize, Serialize};
use std::collections::HashMap;

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct AIProvider {
    pub id: String,
    pub name: String,
    pub base_url: String,
    pub api_key: String,
    pub models: Vec<String>,
    pub is_official: bool,
}

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct Agent {
    pub id: String,
    pub name: String,
    pub emoji: String,
    pub description: String,
    pub model_override: String,
    pub channels: Vec<String>,
    pub tools: Vec<String>,
    pub sandbox_mode: bool,
    pub workspace_isolation: bool,
    pub mention_mode: bool,
    pub sub_agent_permissions: Vec<String>,
    pub is_primary: bool,
}

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct Skill {
    pub name: String,
    pub version: String,
    pub description: String,
    pub category: String,
    pub installed: bool,
    pub enabled: bool,
    pub config_schema: Vec<SkillConfigField>,
    pub config_values: HashMap<String, serde_json::Value>,
}

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct SkillConfigField {
    pub key: String,
    pub label: String,
    pub field_type: String,
    pub default_value: String,
    pub options: Vec<String>,
    pub required: bool,
}
