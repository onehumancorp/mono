use crate::utils::shell;
use serde_json::{json, Value};
use std::process::Command;

#[tauri::command]
pub async fn check_environment() -> Result<Value, String> {
    let node_ok = shell::run_command("node", &["--version"]).is_ok();
    let npm_ok = shell::run_command("npm", &["--version"]).is_ok();
    let openclaw_ok = shell::run_command("openclaw", &["--version"]).is_ok();

    Ok(json!({
        "node": node_ok,
        "npm": npm_ok,
        "openclaw": openclaw_ok
    }))
}

#[tauri::command]
pub async fn install_nodejs() -> Result<String, String> {
    // Open the Node.js download page in the default browser
    #[cfg(target_os = "macos")]
    {
        Command::new("open")
            .arg("https://nodejs.org/en/download")
            .spawn()
            .map_err(|e| e.to_string())?;
    }
    #[cfg(target_os = "linux")]
    {
        Command::new("xdg-open")
            .arg("https://nodejs.org/en/download")
            .spawn()
            .map_err(|e| e.to_string())?;
    }
    #[cfg(target_os = "windows")]
    {
        Command::new("cmd")
            .args(["/C", "start", "https://nodejs.org/en/download"])
            .spawn()
            .map_err(|e| e.to_string())?;
    }
    Ok("Opening Node.js download page in browser".to_string())
}

#[tauri::command]
pub async fn install_openclaw() -> Result<String, String> {
    let npm_path = shell::find_executable("npm")
        .ok_or_else(|| "npm not found. Please install Node.js first.".to_string())?;

    let out = Command::new(&npm_path)
        .args(["install", "-g", "openclaw"])
        .output()
        .map_err(|e| e.to_string())?;

    if out.status.success() {
        Ok("openclaw installed successfully".to_string())
    } else {
        Err(String::from_utf8_lossy(&out.stderr).to_string())
    }
}

#[tauri::command]
pub async fn check_openclaw_update() -> Result<Value, String> {
    let current_version = shell::run_command("openclaw", &["--version"])
        .unwrap_or_else(|_| "unknown".to_string())
        .trim()
        .to_string();

    // Check npm registry for latest version
    let out = Command::new("npm")
        .args(["view", "openclaw", "version"])
        .output()
        .map_err(|e| e.to_string())?;

    let latest = if out.status.success() {
        String::from_utf8_lossy(&out.stdout).trim().to_string()
    } else {
        "unknown".to_string()
    };

    let has_update = !latest.is_empty()
        && latest != "unknown"
        && current_version != "unknown"
        && latest != current_version;

    Ok(json!({
        "has_update": has_update,
        "current_version": current_version,
        "latest_version": latest
    }))
}

#[tauri::command]
pub async fn update_openclaw() -> Result<String, String> {
    let npm_path = shell::find_executable("npm")
        .ok_or_else(|| "npm not found".to_string())?;

    let out = Command::new(&npm_path)
        .args(["install", "-g", "openclaw@latest"])
        .output()
        .map_err(|e| e.to_string())?;

    if out.status.success() {
        Ok("openclaw updated to latest version".to_string())
    } else {
        Err(String::from_utf8_lossy(&out.stderr).to_string())
    }
}

#[tauri::command]
pub async fn uninstall_openclaw() -> Result<String, String> {
    let npm_path = shell::find_executable("npm")
        .ok_or_else(|| "npm not found".to_string())?;

    let out = Command::new(&npm_path)
        .args(["uninstall", "-g", "openclaw"])
        .output()
        .map_err(|e| e.to_string())?;

    if out.status.success() {
        Ok("openclaw uninstalled".to_string())
    } else {
        Err(String::from_utf8_lossy(&out.stderr).to_string())
    }
}
