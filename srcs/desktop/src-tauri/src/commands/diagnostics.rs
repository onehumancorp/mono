use crate::models::status::{DiagnosticResult, SecurityIssue};
use crate::utils::{platform, shell};
use serde_json::{json, Value};
use std::process::Command;

#[tauri::command]
pub async fn run_doctor() -> Result<Vec<DiagnosticResult>, String> {
    let mut results = Vec::new();

    // Check Node.js
    match shell::run_command("node", &["--version"]) {
        Ok(version) => results.push(DiagnosticResult {
            name: "Node.js".to_string(),
            passed: true,
            message: format!("Found {}", version.trim()),
            suggestion: None,
        }),
        Err(_) => results.push(DiagnosticResult {
            name: "Node.js".to_string(),
            passed: false,
            message: "Node.js not found".to_string(),
            suggestion: Some("Install Node.js from https://nodejs.org".to_string()),
        }),
    }

    // Check npm
    match shell::run_command("npm", &["--version"]) {
        Ok(version) => results.push(DiagnosticResult {
            name: "npm".to_string(),
            passed: true,
            message: format!("Found {}", version.trim()),
            suggestion: None,
        }),
        Err(_) => results.push(DiagnosticResult {
            name: "npm".to_string(),
            passed: false,
            message: "npm not found".to_string(),
            suggestion: Some("npm comes with Node.js, reinstall Node.js".to_string()),
        }),
    }

    // Check openclaw
    match shell::run_command("openclaw", &["--version"]) {
        Ok(version) => results.push(DiagnosticResult {
            name: "OpenClaw".to_string(),
            passed: true,
            message: format!("Found {}", version.trim()),
            suggestion: None,
        }),
        Err(_) => results.push(DiagnosticResult {
            name: "OpenClaw".to_string(),
            passed: false,
            message: "openclaw not found".to_string(),
            suggestion: Some("Install with: npm install -g openclaw".to_string()),
        }),
    }

    // Check config file
    let config_path = platform::get_openclaw_dir().map(|d| d.join("openclaw.json"));
    match config_path {
        Some(path) if path.exists() => {
            let content = std::fs::read_to_string(&path).unwrap_or_default();
            let parsed: Result<Value, _> = serde_json::from_str(&content);
            match parsed {
                Ok(_) => results.push(DiagnosticResult {
                    name: "Config File".to_string(),
                    passed: true,
                    message: format!("Valid JSON at {}", path.display()),
                    suggestion: None,
                }),
                Err(e) => results.push(DiagnosticResult {
                    name: "Config File".to_string(),
                    passed: false,
                    message: format!("Invalid JSON: {}", e),
                    suggestion: Some("Fix JSON syntax errors in openclaw.json".to_string()),
                }),
            }
        }
        Some(path) => results.push(DiagnosticResult {
            name: "Config File".to_string(),
            passed: false,
            message: format!("Config not found at {}", path.display()),
            suggestion: Some("Run 'openclaw init' to create default config".to_string()),
        }),
        None => results.push(DiagnosticResult {
            name: "Config File".to_string(),
            passed: false,
            message: "Cannot determine config directory".to_string(),
            suggestion: None,
        }),
    }

    // Check .env file
    let env_path = platform::get_openclaw_dir().map(|d| d.join(".env"));
    match env_path {
        Some(path) if path.exists() => {
            results.push(DiagnosticResult {
                name: "Environment File".to_string(),
                passed: true,
                message: format!(".env found at {}", path.display()),
                suggestion: None,
            });
        }
        Some(path) => {
            results.push(DiagnosticResult {
                name: "Environment File".to_string(),
                passed: false,
                message: format!(".env not found at {}", path.display()),
                suggestion: Some("Create ~/.openclaw/.env with your API keys".to_string()),
            });
        }
        None => {}
    }

    // Check port availability
    let is_port_free = std::net::TcpListener::bind(format!("127.0.0.1:{}", 18789)).is_ok();
    if is_port_free {
        results.push(DiagnosticResult {
            name: "Service Port (18789)".to_string(),
            passed: false,
            message: "Port 18789 is not in use (service not running)".to_string(),
            suggestion: Some("Start the service from the Dashboard".to_string()),
        });
    } else {
        results.push(DiagnosticResult {
            name: "Service Port (18789)".to_string(),
            passed: true,
            message: "Port 18789 is in use (service running)".to_string(),
            suggestion: None,
        });
    }

    Ok(results)
}

#[tauri::command]
pub async fn test_ai_connection(
    provider: String,
    model: String,
    api_key: String,
    base_url: String,
) -> Result<Value, String> {
    let url = if base_url.is_empty() {
        match provider.as_str() {
            "anthropic" => "https://api.anthropic.com/v1/messages".to_string(),
            "openai" => "https://api.openai.com/v1/chat/completions".to_string(),
            "groq" => "https://api.groq.com/openai/v1/chat/completions".to_string(),
            "mistral" => "https://api.mistral.ai/v1/chat/completions".to_string(),
            _ => return Ok(json!({ "success": false, "message": "Unknown provider and no base_url provided" })),
        }
    } else {
        format!("{}/v1/chat/completions", base_url.trim_end_matches('/'))
    };

    // Use a simple HTTP request with reqwest
    let client = reqwest::Client::new();
    let body = json!({
        "model": model,
        "messages": [{"role": "user", "content": "Say 'ok'"}],
        "max_tokens": 10
    });

    let mut req = client.post(&url);
    if !api_key.is_empty() {
        if provider == "anthropic" {
            req = req
                .header("x-api-key", &api_key)
                .header("anthropic-version", "2023-06-01");
        } else {
            req = req.bearer_auth(&api_key);
        }
    }

    match req
        .header("content-type", "application/json")
        .json(&body)
        .send()
        .await
    {
        Ok(resp) => {
            let status = resp.status();
            if status.is_success() {
                Ok(json!({ "success": true, "message": format!("Connection successful (HTTP {})", status) }))
            } else {
                let text = resp.text().await.unwrap_or_default();
                Ok(json!({ "success": false, "message": format!("HTTP {}: {}", status, &text[..text.len().min(200)]) }))
            }
        }
        Err(e) => Ok(json!({ "success": false, "message": format!("Request failed: {}", e) })),
    }
}

#[tauri::command]
pub async fn test_channel(channel: String, _config: Value) -> Result<Value, String> {
    // Test channel connectivity by checking if the configured bot/token is valid
    // This is a stub that checks the saved config and performs a lightweight verification
    let config_result = crate::commands::config::get_channels_config().await;
    let cfg = config_result.unwrap_or(json!({}));
    let ch_cfg = cfg.get(&channel).cloned().unwrap_or(json!({}));

    let enabled = ch_cfg.get("enabled").and_then(|v| v.as_bool()).unwrap_or(false);
    if !enabled {
        return Ok(json!({
            "success": false,
            "message": format!("Channel '{}' is not enabled. Enable it in the Channels tab first.", channel)
        }));
    }

    // Check required fields
    let has_token = match channel.as_str() {
        "telegram" => ch_cfg.get("bot_token").and_then(|v| v.as_str()).map(|s| !s.is_empty()).unwrap_or(false),
        "discord" => ch_cfg.get("bot_token").and_then(|v| v.as_str()).map(|s| !s.is_empty()).unwrap_or(false),
        "slack" => ch_cfg.get("bot_token").and_then(|v| v.as_str()).map(|s| !s.is_empty()).unwrap_or(false),
        "feishu" => ch_cfg.get("app_id").and_then(|v| v.as_str()).map(|s| !s.is_empty()).unwrap_or(false),
        _ => true,
    };

    if !has_token {
        return Ok(json!({
            "success": false,
            "message": format!("Channel '{}' is missing required credentials. Check the Channels tab.", channel)
        }));
    }

    Ok(json!({
        "success": true,
        "message": format!("Channel '{}' configuration looks valid.", channel)
    }))
}

#[tauri::command]
pub async fn get_system_info() -> Result<Value, String> {
    let node_version = shell::run_command("node", &["--version"])
        .ok()
        .map(|v| v.trim().to_string());
    let npm_version = shell::run_command("npm", &["--version"])
        .ok()
        .map(|v| v.trim().to_string());
    let openclaw_version = shell::run_command("openclaw", &["--version"])
        .ok()
        .map(|v| v.trim().to_string());

    let os = std::env::consts::OS.to_string();
    let arch = std::env::consts::ARCH.to_string();
    let platform = format!("{}-{}", os, arch);

    Ok(json!({
        "node_version": node_version,
        "npm_version": npm_version,
        "openclaw_version": openclaw_version,
        "os": os,
        "arch": arch,
        "platform": platform
    }))
}

#[tauri::command]
pub async fn run_security_scan() -> Result<Vec<SecurityIssue>, String> {
    let mut issues = Vec::new();

    // Check if service is bound to 0.0.0.0 (exposed to network)
    let is_exposed = check_port_exposure();
    if is_exposed {
        issues.push(SecurityIssue {
            id: "port_exposure".to_string(),
            title: "Service Exposed to Network".to_string(),
            description: "The OpenClaw gateway is listening on 0.0.0.0, making it accessible from the network.".to_string(),
            severity: "high".to_string(),
            fixable: false,
            fixed: false,
            category: "network".to_string(),
            detail: Some("Consider binding to 127.0.0.1 only in the OpenClaw configuration.".to_string()),
        });
    }

    // Check Gateway Token
    let env_path = platform::get_openclaw_dir().map(|d| d.join(".env"));
    let has_token = env_path
        .as_ref()
        .filter(|p| p.exists())
        .and_then(|p| std::fs::read_to_string(p).ok())
        .map(|content| content.contains("GATEWAY_TOKEN=") && !content.contains("GATEWAY_TOKEN=\n") && !content.contains("GATEWAY_TOKEN= "))
        .unwrap_or(false);

    if !has_token {
        issues.push(SecurityIssue {
            id: "no_gateway_token".to_string(),
            title: "Gateway Token Not Set".to_string(),
            description: "No GATEWAY_TOKEN is configured. Anyone with network access can use the API.".to_string(),
            severity: "high".to_string(),
            fixable: true,
            fixed: false,
            category: "authentication".to_string(),
            detail: Some("Add GATEWAY_TOKEN=<your-secret> to ~/.openclaw/.env".to_string()),
        });
    }

    // Check config file permissions (Unix only)
    #[cfg(unix)]
    {
        use std::os::unix::fs::PermissionsExt;
        if let Some(cfg_path) = platform::get_openclaw_dir().map(|d| d.join("openclaw.json")) {
            if cfg_path.exists() {
                if let Ok(meta) = std::fs::metadata(&cfg_path) {
                    let mode = meta.permissions().mode();
                    if mode & 0o044 != 0 {
                        issues.push(SecurityIssue {
                            id: "config_permissions".to_string(),
                            title: "Config File Permissions Too Open".to_string(),
                            description: "openclaw.json is readable by other users on the system.".to_string(),
                            severity: "medium".to_string(),
                            fixable: true,
                            fixed: false,
                            category: "file_permissions".to_string(),
                            detail: Some(format!("Current permissions: {:o}. Should be 600.", mode & 0o777)),
                        });
                    }
                }
            }
        }
    }

    // Check .env file permissions (Unix only)
    #[cfg(unix)]
    {
        use std::os::unix::fs::PermissionsExt;
        if let Some(env_p) = platform::get_openclaw_dir().map(|d| d.join(".env")) {
            if env_p.exists() {
                if let Ok(meta) = std::fs::metadata(&env_p) {
                    let mode = meta.permissions().mode();
                    if mode & 0o044 != 0 {
                        issues.push(SecurityIssue {
                            id: "env_permissions".to_string(),
                            title: ".env File Permissions Too Open".to_string(),
                            description: ".env file (containing API keys) is readable by other users.".to_string(),
                            severity: "high".to_string(),
                            fixable: true,
                            fixed: false,
                            category: "file_permissions".to_string(),
                            detail: Some(format!("Current permissions: {:o}. Should be 600.", mode & 0o777)),
                        });
                    }
                }
            }
        }
    }

    // Check if skills with high permissions are enabled
    issues.push(SecurityIssue {
        id: "skill_permissions_review".to_string(),
        title: "Review Skill Permissions".to_string(),
        description: "Periodically review which skills are enabled and their required permissions.".to_string(),
        severity: "low".to_string(),
        fixable: false,
        fixed: false,
        category: "permissions".to_string(),
        detail: Some("Disable skills you don't actively use to minimize attack surface.".to_string()),
    });

    Ok(issues)
}

fn check_port_exposure() -> bool {
    // Check if the port is bound on any address other than loopback
    #[cfg(unix)]
    {
        if let Ok(out) = Command::new("lsof")
            .args(["-i", ":18789", "-n", "-P"])
            .output()
        {
            let stdout = String::from_utf8_lossy(&out.stdout);
            return stdout.contains("*:18789") || stdout.contains("0.0.0.0:18789");
        }
    }
    false
}

#[tauri::command]
pub async fn fix_security_issues(issue_ids: Vec<String>) -> Result<Vec<String>, String> {
    let mut fixed = Vec::new();

    for id in &issue_ids {
        match id.as_str() {
            "no_gateway_token" => {
                // Generate a random token and add to .env
                let token: String = (0..32)
                    .map(|_| {
                        let idx = (rand_byte() % 62) as usize;
                        b"abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"[idx] as char
                    })
                    .collect();
                crate::commands::config::save_env_value("GATEWAY_TOKEN".to_string(), token)
                    .await?;
                fixed.push(id.clone());
            }
            "config_permissions" => {
                #[cfg(unix)]
                {
                    use std::os::unix::fs::PermissionsExt;
                    if let Some(path) = platform::get_openclaw_dir().map(|d| d.join("openclaw.json")) {
                        let perms = std::fs::Permissions::from_mode(0o600);
                        std::fs::set_permissions(&path, perms).map_err(|e| e.to_string())?;
                        fixed.push(id.clone());
                    }
                }
            }
            "env_permissions" => {
                #[cfg(unix)]
                {
                    use std::os::unix::fs::PermissionsExt;
                    if let Some(path) = platform::get_openclaw_dir().map(|d| d.join(".env")) {
                        let perms = std::fs::Permissions::from_mode(0o600);
                        std::fs::set_permissions(&path, perms).map_err(|e| e.to_string())?;
                        fixed.push(id.clone());
                    }
                }
            }
            _ => {}
        }
    }

    Ok(fixed)
}

fn rand_byte() -> u8 {
    use std::time::{SystemTime, UNIX_EPOCH};
    static COUNTER: std::sync::atomic::AtomicU64 = std::sync::atomic::AtomicU64::new(0);
    let t = SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .unwrap_or_default()
        .as_nanos() as u64;
    let c = COUNTER.fetch_add(1, std::sync::atomic::Ordering::Relaxed);
    ((t ^ c).wrapping_mul(6364136223846793005).wrapping_add(1442695040888963407) >> 33) as u8
}
