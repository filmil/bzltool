package template

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
)

type TemplateConfig struct {
	Ignore     []string            `json:"ignore"`
	Raw        []string            `json:"raw"`
	Conditions map[string][]string `json:"conditions"`
}

func matchesAny(relPath string, patterns []string) bool {
	for _, pattern := range patterns {
		// Simple wildcard match on the filename
		if matched, _ := filepath.Match(pattern, filepath.Base(relPath)); matched {
			return true
		}
		// Directory or exact match
		if strings.HasPrefix(relPath, pattern) || relPath == strings.TrimSuffix(pattern, "/") {
			return true
		}
	}
	return false
}

func loadTemplateConfig(repoDir string) TemplateConfig {
	var cfg TemplateConfig
	data, err := os.ReadFile(filepath.Join(repoDir, "template.json"))
	if err == nil {
		json.Unmarshal(data, &cfg)
	}
	return cfg
}

// ProcessFragments searches for 'fragments' directories within the provided
// repository directories, sorts them alphanumerically, and evaluates all found
// file templates into the destDir directory using the provided parameters.
// Files with the same relative path are concatenated prior to evaluation.
// activeConditions controls which conditionally-included fragments are evaluated.
func ProcessFragments(repoDirs []string, destDir string, params interface{}, activeConditions []string) error {
	var fragmentDirs []string
	configs := make(map[string]TemplateConfig)
	activeCondMap := make(map[string]bool)
	for _, c := range activeConditions {
		activeCondMap[c] = true
	}

	for _, repoDir := range repoDirs {
		configs[repoDir] = loadTemplateConfig(repoDir)

		entries, err := os.ReadDir(repoDir)
		if err != nil {
			return err
		}
		for _, entry := range entries {
			if entry.IsDir() {
				fDir := filepath.Join(repoDir, entry.Name(), "fragments")
				if _, err := os.Stat(fDir); err == nil {
					fragmentDirs = append(fragmentDirs, fDir)
				}
			}
		}
	}

	sort.Strings(fragmentDirs)

	fragmentsMap := make(map[string][]string)
	rawMap := make(map[string][]byte) // Stores latest raw file content

	for _, fDir := range fragmentDirs {
		repoDir := filepath.Dir(filepath.Dir(fDir)) // fDir is <repoDir>/<module>/fragments
		cfg := configs[repoDir]

		err := filepath.Walk(fDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}

			relPath, err := filepath.Rel(fDir, path)
			if err != nil {
				return err
			}

			if matchesAny(relPath, cfg.Ignore) {
				return nil // Skip ignored file
			}

			// Check conditionals
			skipConditionally := false
			for condName, patterns := range cfg.Conditions {
				if matchesAny(relPath, patterns) {
					if !activeCondMap[condName] {
						skipConditionally = true
						break
					}
				}
			}
			if skipConditionally {
				return nil
			}

			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			if matchesAny(relPath, cfg.Raw) {
				// Raw files overwrite previous raw files (no concatenation)
				rawMap[relPath] = content
			} else {
				fragmentsMap[relPath] = append(fragmentsMap[relPath], string(content))
			}
			return nil
		})
		if err != nil {
			return err
		}
	}

	for relPath, fragments := range fragmentsMap {
		var combined string
		for _, frag := range fragments {
			combined += frag
		}

		tmpl, err := template.New(relPath).Parse(combined)
		if err != nil {
			return err
		}

		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, params); err != nil {
			return err
		}

		outPath := filepath.Join(destDir, relPath)
		if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
			return err
		}

		if err := os.WriteFile(outPath, buf.Bytes(), 0644); err != nil {
			return err
		}
	}

	for relPath, rawContent := range rawMap {
		outPath := filepath.Join(destDir, relPath)
		if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
			return err
		}

		if err := os.WriteFile(outPath, rawContent, 0644); err != nil {
			return err
		}
	}

	return nil
}
