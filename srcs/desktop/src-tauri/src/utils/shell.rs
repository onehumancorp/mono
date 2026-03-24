use std::path::PathBuf;
use std::process::Command;

/// Find an executable in PATH, returning its full path if found.
pub fn find_executable(name: &str) -> Option<PathBuf> {
    which::which(name).ok()
}

/// Run a command and return its stdout as a String.
pub fn run_command(cmd: &str, args: &[&str]) -> Result<String, String> {
    let output = Command::new(cmd)
        .args(args)
        .output()
        .map_err(|e| format!("Failed to run '{}': {}", cmd, e))?;

    if output.status.success() {
        Ok(String::from_utf8_lossy(&output.stdout).to_string())
    } else {
        Err(String::from_utf8_lossy(&output.stderr).to_string())
    }
}

/// Run a shell command and return combined output.
pub fn run_shell(command: &str) -> Result<String, String> {
    #[cfg(unix)]
    let output = Command::new("sh")
        .args(["-c", command])
        .output()
        .map_err(|e| e.to_string())?;

    #[cfg(windows)]
    let output = Command::new("cmd")
        .args(["/C", command])
        .output()
        .map_err(|e| e.to_string())?;

    let stdout = String::from_utf8_lossy(&output.stdout).to_string();
    let stderr = String::from_utf8_lossy(&output.stderr).to_string();

    if output.status.success() {
        Ok(stdout)
    } else {
        Err(if stderr.is_empty() { stdout } else { stderr })
    }
}
