package dashboard

import (
	"encoding/json"
	"net/http"
	"gopkg.in/yaml.v3"
	"github.com/onehumancorp/mono/srcs/orchestration"
)

// CapabilityManifest defines the schema for a plugin manifest.
type CapabilityManifest struct {
	PluginID    string `yaml:"plugin_id" json:"plugin_id"`
	Name        string `yaml:"name" json:"name"`
	Version     string `yaml:"version" json:"version"`
	ManifestURL string `yaml:"manifest_url" json:"manifest_url"`
}

// handleImportPlugin handles the registration of a new plugin via manifest.
func (s *Server) handleImportPlugin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var manifest CapabilityManifest
	decoder := yaml.NewDecoder(r.Body)
	if err := decoder.Decode(&manifest); err != nil {
		http.Error(w, "Invalid plugin manifest schema", http.StatusBadRequest)
		return
	}

	if manifest.PluginID == "" || manifest.Name == "" || manifest.Version == "" || manifest.ManifestURL == "" {
		http.Error(w, "Missing required fields in manifest", http.StatusBadRequest)
		return
	}

	if sipdb, ok := s.hub.SIPDB().(*orchestration.SIPDB); ok && sipdb != nil {
		err := sipdb.RegisterCapabilityPlugin(r.Context(), manifest.PluginID, manifest.Name, manifest.Version, manifest.ManifestURL)
		if err != nil {
			http.Error(w, "Failed to register plugin", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "ACTIVE"})
}

// handleGetPlugins retrieves the registered plugins.
func (s *Server) handleGetPlugins(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if sipdb, ok := s.hub.SIPDB().(*orchestration.SIPDB); ok && sipdb != nil {
		plugins, err := sipdb.QueryCapabilityPlugins(r.Context())
		if err != nil {
			http.Error(w, "Failed to query plugins", http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(plugins)
	} else {
		json.NewEncoder(w).Encode([]string{})
	}
}
