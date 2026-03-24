// Prevents additional console window on Windows in release
#![cfg_attr(not(debug_assertions), windows_subsystem = "windows")]

mod commands;
mod models;
mod utils;

use commands::{config, diagnostics, installer, service};

fn main() {
    env_logger::init();
    tauri::Builder::default()
        .plugin(tauri_plugin_shell::init())
        .plugin(tauri_plugin_fs::init())
        .plugin(tauri_plugin_os::init())
        .plugin(tauri_plugin_dialog::init())
        .invoke_handler(tauri::generate_handler![
            // Service commands
            service::get_service_status,
            service::start_service,
            service::stop_service,
            service::restart_service,
            service::get_logs,
            // Config commands
            config::get_config,
            config::save_config,
            config::get_env_value,
            config::save_env_value,
            config::get_official_providers,
            config::get_ai_config,
            config::save_provider,
            config::delete_provider,
            config::set_primary_model,
            config::get_channels_config,
            config::save_channel_config,
            config::get_agents_list,
            config::save_agent,
            config::delete_agent,
            config::set_default_agent,
            config::get_skills_list,
            config::install_skill,
            config::uninstall_skill,
            config::install_custom_skill,
            config::save_skill_config,
            // Diagnostics commands
            diagnostics::run_doctor,
            diagnostics::test_ai_connection,
            diagnostics::test_channel,
            diagnostics::get_system_info,
            diagnostics::run_security_scan,
            diagnostics::fix_security_issues,
            // Installer commands
            installer::check_environment,
            installer::install_nodejs,
            installer::install_openclaw,
            installer::check_openclaw_update,
            installer::update_openclaw,
            installer::uninstall_openclaw,
        ])
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
