use std::path::PathBuf;

/// Return the path to the openclaw config directory.
/// - macOS/Linux: `~/.openclaw`
/// - Windows: `%APPDATA%\openclaw`
pub fn get_openclaw_dir() -> Option<PathBuf> {
    #[cfg(target_os = "windows")]
    {
        dirs::data_dir().map(|d| d.join("openclaw"))
    }
    #[cfg(not(target_os = "windows"))]
    {
        dirs::home_dir().map(|d| d.join(".openclaw"))
    }
}

/// Return the platform name as a short string.
pub fn platform_name() -> &'static str {
    #[cfg(target_os = "macos")]
    return "macos";
    #[cfg(target_os = "linux")]
    return "linux";
    #[cfg(target_os = "windows")]
    return "windows";
    #[allow(unreachable_code)]
    "unknown"
}
