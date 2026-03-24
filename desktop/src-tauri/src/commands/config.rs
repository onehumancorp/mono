use crate::models::config::{Agent, Skill, SkillConfigField};
use crate::utils::platform;
use serde_json::{json, Value};
use std::collections::HashMap;

fn config_path() -> Result<std::path::PathBuf, String> {
    platform::get_openclaw_dir()
        .map(|d| d.join("openclaw.json"))
        .ok_or_else(|| "Cannot determine config directory".to_string())
}

fn env_path() -> Result<std::path::PathBuf, String> {
    platform::get_openclaw_dir()
        .map(|d| d.join(".env"))
        .ok_or_else(|| "Cannot determine config directory".to_string())
}

fn read_config() -> Result<Value, String> {
    let path = config_path()?;
    if !path.exists() {
        return Ok(json!({}));
    }
    let content = std::fs::read_to_string(&path).map_err(|e| e.to_string())?;
    serde_json::from_str(&content).map_err(|e| e.to_string())
}

fn write_config(value: &Value) -> Result<(), String> {
    let path = config_path()?;
    if let Some(parent) = path.parent() {
        std::fs::create_dir_all(parent).map_err(|e| e.to_string())?;
    }
    let content = serde_json::to_string_pretty(value).map_err(|e| e.to_string())?;
    std::fs::write(&path, content).map_err(|e| e.to_string())
}

#[tauri::command]
pub async fn get_config() -> Result<Value, String> {
    read_config()
}

#[tauri::command]
pub async fn save_config(config: Value) -> Result<(), String> {
    write_config(&config)
}

#[tauri::command]
pub async fn get_env_value(key: String) -> Result<Option<String>, String> {
    let path = env_path()?;
    if !path.exists() {
        return Ok(None);
    }
    let content = std::fs::read_to_string(&path).map_err(|e| e.to_string())?;
    for line in content.lines() {
        if let Some(stripped) = line.strip_prefix(&format!("{}=", key)) {
            let val = stripped.trim_matches('"').to_string();
            return Ok(Some(val));
        }
    }
    Ok(None)
}

#[tauri::command]
pub async fn save_env_value(key: String, value: String) -> Result<(), String> {
    let path = env_path()?;
    if let Some(parent) = path.parent() {
        std::fs::create_dir_all(parent).map_err(|e| e.to_string())?;
    }
    let mut lines: Vec<String> = if path.exists() {
        std::fs::read_to_string(&path)
            .map_err(|e| e.to_string())?
            .lines()
            .map(String::from)
            .collect()
    } else {
        vec![]
    };
    let prefix = format!("{}=", key);
    let new_line = format!("{}={}", key, value);
    let mut found = false;
    for line in lines.iter_mut() {
        if line.starts_with(&prefix) {
            *line = new_line.clone();
            found = true;
            break;
        }
    }
    if !found {
        lines.push(new_line);
    }
    std::fs::write(&path, lines.join("\n") + "\n").map_err(|e| e.to_string())
}

#[tauri::command]
pub async fn get_official_providers() -> Result<Value, String> {
    Ok(json!([
        {
            "id": "anthropic",
            "name": "Anthropic",
            "default_base_url": "https://api.anthropic.com",
            "default_models": ["claude-opus-4-5", "claude-sonnet-4-5", "claude-haiku-4-5"],
            "requires_api_key": true,
            "website": "https://console.anthropic.com"
        },
        {
            "id": "openai",
            "name": "OpenAI",
            "default_base_url": "https://api.openai.com",
            "default_models": ["gpt-4o", "gpt-4o-mini", "o1", "o3-mini"],
            "requires_api_key": true,
            "website": "https://platform.openai.com"
        },
        {
            "id": "deepseek",
            "name": "DeepSeek",
            "default_base_url": "https://api.deepseek.com",
            "default_models": ["deepseek-chat", "deepseek-reasoner"],
            "requires_api_key": true,
            "website": "https://platform.deepseek.com"
        },
        {
            "id": "moonshot",
            "name": "Moonshot",
            "default_base_url": "https://api.moonshot.cn",
            "default_models": ["moonshot-v1-8k", "moonshot-v1-32k"],
            "requires_api_key": true,
            "website": "https://platform.moonshot.cn"
        },
        {
            "id": "gemini",
            "name": "Google Gemini",
            "default_base_url": "https://generativelanguage.googleapis.com",
            "default_models": ["gemini-2.0-flash", "gemini-1.5-pro", "gemini-1.5-flash"],
            "requires_api_key": true,
            "website": "https://aistudio.google.com"
        },
        {
            "id": "azure",
            "name": "Azure OpenAI",
            "default_base_url": "",
            "default_models": [],
            "requires_api_key": true,
            "website": "https://azure.microsoft.com/en-us/products/ai-services/openai-service"
        },
        {
            "id": "groq",
            "name": "Groq",
            "default_base_url": "https://api.groq.com",
            "default_models": ["llama-3.3-70b-versatile", "mixtral-8x7b-32768"],
            "requires_api_key": true,
            "website": "https://console.groq.com"
        },
        {
            "id": "mistral",
            "name": "Mistral AI",
            "default_base_url": "https://api.mistral.ai",
            "default_models": ["mistral-large-latest", "mistral-small-latest"],
            "requires_api_key": true,
            "website": "https://console.mistral.ai"
        },
        {
            "id": "cohere",
            "name": "Cohere",
            "default_base_url": "https://api.cohere.ai",
            "default_models": ["command-r-plus", "command-r"],
            "requires_api_key": true,
            "website": "https://dashboard.cohere.com"
        },
        {
            "id": "xai",
            "name": "xAI (Grok)",
            "default_base_url": "https://api.x.ai",
            "default_models": ["grok-beta", "grok-vision-beta"],
            "requires_api_key": true,
            "website": "https://x.ai"
        },
        {
            "id": "ollama",
            "name": "Ollama (Local)",
            "default_base_url": "http://localhost:11434",
            "default_models": [],
            "requires_api_key": false,
            "website": "https://ollama.com"
        },
        {
            "id": "lmstudio",
            "name": "LM Studio (Local)",
            "default_base_url": "http://localhost:1234",
            "default_models": [],
            "requires_api_key": false,
            "website": "https://lmstudio.ai"
        },
        {
            "id": "qwen",
            "name": "Alibaba Qwen",
            "default_base_url": "https://dashscope.aliyuncs.com",
            "default_models": ["qwen-max", "qwen-plus", "qwen-turbo"],
            "requires_api_key": true,
            "website": "https://dashscope.aliyun.com"
        },
        {
            "id": "zhipu",
            "name": "Zhipu AI (GLM)",
            "default_base_url": "https://open.bigmodel.cn",
            "default_models": ["glm-4", "glm-4-flash", "glm-4v"],
            "requires_api_key": true,
            "website": "https://open.bigmodel.cn"
        }
    ]))
}

#[tauri::command]
pub async fn get_ai_config() -> Result<Value, String> {
    let cfg = read_config()?;
    let providers = cfg.get("providers").cloned().unwrap_or(json!([]));
    let primary_model = cfg
        .get("primaryModel")
        .and_then(|v| v.as_str())
        .unwrap_or("")
        .to_string();
    Ok(json!({ "providers": providers, "primary_model": primary_model }))
}

#[tauri::command]
pub async fn save_provider(
    id: String,
    base_url: String,
    api_key: String,
    models: Vec<String>,
) -> Result<(), String> {
    let mut cfg = read_config()?;
    let providers = cfg
        .get_mut("providers")
        .and_then(|v| v.as_array_mut());

    let new_entry = json!({
        "id": id,
        "base_url": base_url,
        "api_key": api_key,
        "models": models,
        "is_official": true,
        "name": id
    });

    if let Some(arr) = providers {
        if let Some(pos) = arr.iter().position(|p| p.get("id").and_then(|v| v.as_str()) == Some(&id)) {
            arr[pos] = new_entry;
        } else {
            arr.push(new_entry);
        }
    } else {
        cfg["providers"] = json!([new_entry]);
    }

    write_config(&cfg)
}

#[tauri::command]
pub async fn delete_provider(id: String) -> Result<(), String> {
    let mut cfg = read_config()?;
    if let Some(arr) = cfg.get_mut("providers").and_then(|v| v.as_array_mut()) {
        arr.retain(|p| p.get("id").and_then(|v| v.as_str()) != Some(&id));
    }
    write_config(&cfg)
}

#[tauri::command]
pub async fn set_primary_model(model: String) -> Result<(), String> {
    let mut cfg = read_config()?;
    cfg["primaryModel"] = json!(model);
    write_config(&cfg)
}

#[tauri::command]
pub async fn get_channels_config() -> Result<Value, String> {
    let cfg = read_config()?;
    Ok(cfg.get("channels").cloned().unwrap_or(json!({})))
}

#[tauri::command]
pub async fn save_channel_config(channel: String, config: Value) -> Result<(), String> {
    let mut cfg = read_config()?;
    if cfg.get("channels").is_none() {
        cfg["channels"] = json!({});
    }
    cfg["channels"][channel] = config;
    write_config(&cfg)
}

#[tauri::command]
pub async fn get_agents_list() -> Result<Vec<Agent>, String> {
    let cfg = read_config()?;
    let agents_val = cfg.get("agents").cloned().unwrap_or(json!([]));
    serde_json::from_value(agents_val).map_err(|e| e.to_string())
}

#[tauri::command]
pub async fn save_agent(agent: Agent) -> Result<(), String> {
    let mut cfg = read_config()?;
    let agents = cfg
        .get_mut("agents")
        .and_then(|v| v.as_array_mut());

    let new_val = serde_json::to_value(&agent).map_err(|e| e.to_string())?;
    if let Some(arr) = agents {
        if let Some(pos) = arr.iter().position(|a| {
            a.get("id").and_then(|v| v.as_str()) == Some(&agent.id)
        }) {
            arr[pos] = new_val;
        } else {
            arr.push(new_val);
        }
    } else {
        cfg["agents"] = json!([new_val]);
    }

    write_config(&cfg)
}

#[tauri::command]
pub async fn delete_agent(id: String) -> Result<(), String> {
    let mut cfg = read_config()?;
    if let Some(arr) = cfg.get_mut("agents").and_then(|v| v.as_array_mut()) {
        arr.retain(|a| a.get("id").and_then(|v| v.as_str()) != Some(&id));
    }
    write_config(&cfg)
}

#[tauri::command]
pub async fn set_default_agent(id: String) -> Result<(), String> {
    let mut cfg = read_config()?;
    if let Some(arr) = cfg.get_mut("agents").and_then(|v| v.as_array_mut()) {
        for agent in arr.iter_mut() {
            let is_this = agent.get("id").and_then(|v| v.as_str()) == Some(&id);
            agent["is_primary"] = json!(is_this);
        }
    }
    write_config(&cfg)
}

#[tauri::command]
pub async fn get_skills_list() -> Result<Vec<Skill>, String> {
    use std::process::Command;
    // Try to get installed skills from openclaw
    let output = Command::new("openclaw")
        .args(["skills", "list", "--json"])
        .output();

    match output {
        Ok(out) if out.status.success() => {
            let text = String::from_utf8_lossy(&out.stdout);
            serde_json::from_str(&text).map_err(|e| e.to_string())
        }
        _ => {
            // Return a sample skill list for demo
            Ok(vec![
                Skill {
                    name: "web-search".to_string(),
                    version: "1.0.0".to_string(),
                    description: "Search the web using DuckDuckGo or Google".to_string(),
                    category: "builtin".to_string(),
                    installed: true,
                    enabled: true,
                    config_schema: vec![
                        SkillConfigField {
                            key: "engine".to_string(),
                            label: "Search Engine".to_string(),
                            field_type: "select".to_string(),
                            default_value: "duckduckgo".to_string(),
                            options: vec!["duckduckgo".to_string(), "google".to_string()],
                            required: false,
                        }
                    ],
                    config_values: HashMap::new(),
                },
                Skill {
                    name: "code-exec".to_string(),
                    version: "1.0.0".to_string(),
                    description: "Execute code in sandboxed environments".to_string(),
                    category: "builtin".to_string(),
                    installed: true,
                    enabled: false,
                    config_schema: vec![],
                    config_values: HashMap::new(),
                },
                Skill {
                    name: "file-manager".to_string(),
                    version: "1.2.0".to_string(),
                    description: "Read and write files on the local filesystem".to_string(),
                    category: "official".to_string(),
                    installed: false,
                    enabled: false,
                    config_schema: vec![
                        SkillConfigField {
                            key: "allowed_paths".to_string(),
                            label: "Allowed Base Paths".to_string(),
                            field_type: "text".to_string(),
                            default_value: "~/Documents".to_string(),
                            options: vec![],
                            required: false,
                        }
                    ],
                    config_values: HashMap::new(),
                },
                Skill {
                    name: "image-gen".to_string(),
                    version: "0.9.0".to_string(),
                    description: "Generate images with DALL-E or Stable Diffusion".to_string(),
                    category: "official".to_string(),
                    installed: false,
                    enabled: false,
                    config_schema: vec![
                        SkillConfigField {
                            key: "provider".to_string(),
                            label: "Image Provider".to_string(),
                            field_type: "select".to_string(),
                            default_value: "dalle".to_string(),
                            options: vec!["dalle".to_string(), "stable-diffusion".to_string()],
                            required: true,
                        },
                        SkillConfigField {
                            key: "api_key".to_string(),
                            label: "API Key".to_string(),
                            field_type: "password".to_string(),
                            default_value: "".to_string(),
                            options: vec![],
                            required: true,
                        },
                    ],
                    config_values: HashMap::new(),
                },
                Skill {
                    name: "email-sender".to_string(),
                    version: "1.0.3".to_string(),
                    description: "Send emails via SMTP or SendGrid".to_string(),
                    category: "community".to_string(),
                    installed: false,
                    enabled: false,
                    config_schema: vec![
                        SkillConfigField {
                            key: "smtp_host".to_string(),
                            label: "SMTP Host".to_string(),
                            field_type: "text".to_string(),
                            default_value: "smtp.gmail.com".to_string(),
                            options: vec![],
                            required: true,
                        },
                        SkillConfigField {
                            key: "smtp_port".to_string(),
                            label: "SMTP Port".to_string(),
                            field_type: "text".to_string(),
                            default_value: "587".to_string(),
                            options: vec![],
                            required: true,
                        },
                        SkillConfigField {
                            key: "username".to_string(),
                            label: "Username".to_string(),
                            field_type: "text".to_string(),
                            default_value: "".to_string(),
                            options: vec![],
                            required: true,
                        },
                        SkillConfigField {
                            key: "password".to_string(),
                            label: "Password / App Password".to_string(),
                            field_type: "password".to_string(),
                            default_value: "".to_string(),
                            options: vec![],
                            required: true,
                        },
                    ],
                    config_values: HashMap::new(),
                },
            ])
        }
    }
}

#[tauri::command]
pub async fn install_skill(name: String) -> Result<String, String> {
    use std::process::Command;
    let out = Command::new("openclaw")
        .args(["skills", "install", &name])
        .output()
        .map_err(|e| e.to_string())?;
    if out.status.success() {
        Ok(format!("Skill '{}' installed", name))
    } else {
        Err(String::from_utf8_lossy(&out.stderr).to_string())
    }
}

#[tauri::command]
pub async fn uninstall_skill(name: String) -> Result<String, String> {
    use std::process::Command;
    let out = Command::new("openclaw")
        .args(["skills", "uninstall", &name])
        .output()
        .map_err(|e| e.to_string())?;
    if out.status.success() {
        Ok(format!("Skill '{}' uninstalled", name))
    } else {
        Err(String::from_utf8_lossy(&out.stderr).to_string())
    }
}

#[tauri::command]
pub async fn install_custom_skill(source: String) -> Result<String, String> {
    use std::process::Command;
    let out = Command::new("openclaw")
        .args(["skills", "install", "--source", &source])
        .output()
        .map_err(|e| e.to_string())?;
    if out.status.success() {
        Ok(format!("Custom skill from '{}' installed", source))
    } else {
        Err(String::from_utf8_lossy(&out.stderr).to_string())
    }
}

#[tauri::command]
pub async fn save_skill_config(name: String, config: Value) -> Result<(), String> {
    let mut cfg = read_config()?;
    if cfg.get("skill_configs").is_none() {
        cfg["skill_configs"] = json!({});
    }
    cfg["skill_configs"][name] = config;
    write_config(&cfg)
}
