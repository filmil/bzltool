package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// TemplateRepo represents a template repository source.
type TemplateRepo struct {
	URL    string `json:"url"`
	Subdir string `json:"subdir,omitempty"`
}

// UnmarshalJSON implements custom JSON unmarshaling to support both string URLs and object structures.
func (tr *TemplateRepo) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		tr.URL = s
		return nil
	}

	type Alias TemplateRepo
	var a Alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	*tr = TemplateRepo(a)
	return nil
}

// Config represents the user-level configuration stored in the XDG config directory.
type Config struct {
	TemplateRepos []TemplateRepo `json:"template_repos"`
}

// GetConfigDir returns the directory path where bzltool's configuration is stored, typically following XDG base directory specifications.
func GetConfigDir() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "bzltool"), nil
}

// LoadConfig reads and parses the configuration file from the config directory. It returns an empty config if the file does not exist.
func LoadConfig() (*Config, error) {
	dir, err := GetConfigDir()
	if err != nil {
		return nil, err
	}
	path := filepath.Join(dir, "config.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
