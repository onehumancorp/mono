use crate::models::status::ServiceStatus;
use crate::utils::{platform, shell};
use std::process::Command;
use std::time::{Duration, SystemTime, UNIX_EPOCH};
use sysinfo::{Pid, ProcessRefreshKind, System};

const OPENCLAW_PORT: u16 = 18789;
const LOG_FILE_NAME: &str = "openclaw.log";

fn get_log_path() -> Option<std::path::PathBuf> {
    platform::get_openclaw_dir().map(|d| d.join(LOG_FILE_NAME))
}

/// Find the PID listening on the openclaw port.
fn find_port_pid() -> Option<u32> {
    // Try lsof on unix
    #[cfg(unix)]
    {
        let out = Command::new("lsof")
            .args(["-ti", &format!(":{}", OPENCLAW_PORT)])
            .output()
            .ok()?;
        let stdout = String::from_utf8_lossy(&out.stdout);
        let pid_str = stdout.trim().lines().next()?;
        pid_str.parse::<u32>().ok()
    }
    #[cfg(windows)]
    {
        let out = Command::new("netstat")
            .args(["-ano"])
            .output()
            .ok()?;
        let stdout = String::from_utf8_lossy(&out.stdout);
        for line in stdout.lines() {
            if line.contains(&format!(":{}", OPENCLAW_PORT)) && line.contains("LISTENING") {
                let parts: Vec<&str> = line.split_whitespace().collect();
                if let Some(pid) = parts.last() {
                    return pid.parse::<u32>().ok();
                }
            }
        }
        None
    }
}

fn get_process_info(pid: u32) -> (Option<f64>, Option<f64>, Option<u64>) {
    let mut sys = System::new();
    sys.refresh_processes_specifics(
        sysinfo::ProcessesToUpdate::All,
        ProcessRefreshKind::new().with_memory().with_cpu(),
    );
    if let Some(proc) = sys.process(Pid::from_u32(pid)) {
        let mem_mb = proc.memory() as f64 / 1024.0 / 1024.0;
        let cpu = proc.cpu_usage() as f64;
        let start_time = proc.start_time();
        let now = SystemTime::now()
            .duration_since(UNIX_EPOCH)
            .unwrap_or(Duration::ZERO)
            .as_secs();
        let uptime = now.saturating_sub(start_time);
        (Some(mem_mb), Some(cpu), Some(uptime))
    } else {
        (None, None, None)
    }
}

#[tauri::command]
pub async fn get_service_status() -> Result<ServiceStatus, String> {
    let pid = find_port_pid();
    let running = pid.is_some();
    let (memory_mb, cpu_percent, uptime_seconds) = pid
        .map(get_process_info)
        .unwrap_or((None, None, None));

    Ok(ServiceStatus {
        running,
        pid,
        port: OPENCLAW_PORT,
        uptime_seconds,
        memory_mb,
        cpu_percent,
    })
}

#[tauri::command]
pub async fn start_service() -> Result<String, String> {
    // Check if already running
    if find_port_pid().is_some() {
        return Ok("Service already running".to_string());
    }

    let openclaw_path = shell::find_executable("openclaw")
        .or_else(|| shell::find_executable("openclaw.cmd"))
        .ok_or_else(|| "openclaw not found. Please install it with: npm install -g openclaw".to_string())?;

    let log_path = get_log_path().ok_or("Cannot determine log file path")?;
    std::fs::create_dir_all(log_path.parent().unwrap()).map_err(|e| e.to_string())?;

    let log_file = std::fs::OpenOptions::new()
        .create(true)
        .append(true)
        .open(&log_path)
        .map_err(|e| format!("Cannot open log file: {}", e))?;

    let log_file2 = log_file.try_clone().map_err(|e| e.to_string())?;

    #[cfg(unix)]
    {
        use std::os::unix::process::CommandExt;
        Command::new(&openclaw_path)
            .arg("gateway")
            .stdout(log_file)
            .stderr(log_file2)
            .process_group(0)
            .spawn()
            .map_err(|e| format!("Failed to start service: {}", e))?;
    }

    #[cfg(windows)]
    {
        Command::new(&openclaw_path)
            .arg("gateway")
            .stdout(log_file)
            .stderr(log_file2)
            .creation_flags(0x00000008) // DETACHED_PROCESS
            .spawn()
            .map_err(|e| format!("Failed to start service: {}", e))?;
    }

    // Wait a moment for startup
    tokio::time::sleep(Duration::from_millis(1500)).await;
    Ok("Service started".to_string())
}

#[tauri::command]
pub async fn stop_service() -> Result<String, String> {
    let pid = find_port_pid().ok_or("Service is not running")?;

    #[cfg(unix)]
    {
        Command::new("kill")
            .args(["-TERM", &pid.to_string()])
            .status()
            .map_err(|e| e.to_string())?;
    }

    #[cfg(windows)]
    {
        Command::new("taskkill")
            .args(["/PID", &pid.to_string(), "/F"])
            .status()
            .map_err(|e| e.to_string())?;
    }

    tokio::time::sleep(Duration::from_millis(800)).await;
    Ok(format!("Service stopped (PID {})", pid))
}

#[tauri::command]
pub async fn restart_service() -> Result<String, String> {
    if find_port_pid().is_some() {
        stop_service().await?;
        tokio::time::sleep(Duration::from_millis(500)).await;
    }
    start_service().await
}

#[tauri::command]
pub async fn get_logs(lines: u32) -> Result<Vec<String>, String> {
    let log_path = match get_log_path() {
        Some(p) => p,
        None => return Ok(vec![]),
    };

    if !log_path.exists() {
        return Ok(vec![]);
    }

    let content = std::fs::read_to_string(&log_path).map_err(|e| e.to_string())?;
    let all_lines: Vec<String> = content.lines().map(String::from).collect();
    let total = all_lines.len();
    let start = total.saturating_sub(lines as usize);
    Ok(all_lines[start..].to_vec())
}
