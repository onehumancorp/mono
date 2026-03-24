use serde::{Deserialize, Serialize};

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct ServiceStatus {
    pub running: bool,
    pub pid: Option<u32>,
    pub port: u16,
    pub uptime_seconds: Option<u64>,
    pub memory_mb: Option<f64>,
    pub cpu_percent: Option<f64>,
}

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct DiagnosticResult {
    pub name: String,
    pub passed: bool,
    pub message: String,
    pub suggestion: Option<String>,
}

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct SecurityIssue {
    pub id: String,
    pub title: String,
    pub description: String,
    pub severity: String,
    pub fixable: bool,
    pub fixed: bool,
    pub category: String,
    pub detail: Option<String>,
}
