use std::path::Path;

/// Ensure a directory exists, creating it and all parents if needed.
pub fn ensure_dir(path: &Path) -> Result<(), String> {
    std::fs::create_dir_all(path).map_err(|e| format!("Failed to create directory {}: {}", path.display(), e))
}

/// Read a text file, returning an empty string if it doesn't exist.
pub fn read_text_file(path: &Path) -> Result<String, String> {
    if !path.exists() {
        return Ok(String::new());
    }
    std::fs::read_to_string(path).map_err(|e| e.to_string())
}

/// Write text to a file, creating parent directories as needed.
pub fn write_text_file(path: &Path, content: &str) -> Result<(), String> {
    if let Some(parent) = path.parent() {
        ensure_dir(parent)?;
    }
    std::fs::write(path, content).map_err(|e| e.to_string())
}

/// Read the last `n` lines from a file efficiently.
pub fn tail_file(path: &Path, n: usize) -> Result<Vec<String>, String> {
    let content = read_text_file(path)?;
    let lines: Vec<String> = content.lines().map(String::from).collect();
    let total = lines.len();
    let start = total.saturating_sub(n);
    Ok(lines[start..].to_vec())
}
