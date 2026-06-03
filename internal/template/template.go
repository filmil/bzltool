package template

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"
)

type TemplateConfig struct {
	Ignore          []string                `json:"ignore"`
	Raw             []string                `json:"raw"`
	Conditions      map[string][]string     `json:"conditions"`
	MergeStrategies map[string]string       `json:"merge_strategies"`
	Hooks           map[string][][]string   `json:"hooks"`
}

func deepMergeMap(dest, src map[string]interface{}) map[string]interface{} {
	if dest == nil {
		dest = make(map[string]interface{})
	}
	for k, v := range src {
		if srcMap, ok := v.(map[string]interface{}); ok {
			if destMap, ok := dest[k].(map[string]interface{}); ok {
				dest[k] = deepMergeMap(destMap, srcMap)
				continue
			}
		} else if srcArr, ok := v.([]interface{}); ok {
			if destArr, ok := dest[k].([]interface{}); ok {
				dest[k] = append(destArr, srcArr...)
				continue
			}
		}
		dest[k] = v
	}
	return dest
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
	strategyMap := make(map[string]string)
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

			// Track merge strategy
			for pattern, strat := range cfg.MergeStrategies {
				if matchesAny(relPath, []string{pattern}) {
					strategyMap[relPath] = strat
				}
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
		var finalContent []byte
		strategy := strategyMap[relPath]

		if strategy == "json_deep_merge" || strategy == "yaml_deep_merge" {
			var mergedMap map[string]interface{}
			for _, frag := range fragments {
				tmpl, err := template.New(relPath).Parse(frag)
				if err != nil {
					return err
				}
				var buf bytes.Buffer
				if err := tmpl.Execute(&buf, params); err != nil {
					return err
				}

				var currentMap map[string]interface{}
				if strategy == "json_deep_merge" {
					if err := json.Unmarshal(buf.Bytes(), &currentMap); err != nil {
						return err
					}
				} else {
					if err := yaml.Unmarshal(buf.Bytes(), &currentMap); err != nil {
						return err
					}
				}

				if mergedMap == nil {
					mergedMap = currentMap
				} else {
					mergedMap = deepMergeMap(mergedMap, currentMap)
				}
			}

			var err error
			if strategy == "json_deep_merge" {
				finalContent, err = json.MarshalIndent(mergedMap, "", "  ")
			} else {
				finalContent, err = yaml.Marshal(mergedMap)
			}
			if err != nil {
				return err
			}
		} else if strategy == "override" {
			// Take the last fragment (highest priority)
			frag := fragments[len(fragments)-1]
			tmpl, err := template.New(relPath).Parse(frag)
			if err != nil {
				return err
			}

			var buf bytes.Buffer
			if err := tmpl.Execute(&buf, params); err != nil {
				return err
			}
			finalContent = buf.Bytes()
		} else {
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
			finalContent = buf.Bytes()
		}

		outPath := filepath.Join(destDir, relPath)
		if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
			return err
		}

		if err := os.WriteFile(outPath, finalContent, 0644); err != nil {
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

	// Execute post_merge hooks
	for _, cfg := range configs {
		if postMergeHooks, ok := cfg.Hooks["post_merge"]; ok {
			for _, hookCmd := range postMergeHooks {
				if len(hookCmd) == 0 {
					continue
				}

				// Basic templating on hook arguments to support {{.ProjectName}}
				var templatedArgs []string
				for _, arg := range hookCmd {
					tmpl, err := template.New("hook").Parse(arg)
					if err == nil {
						var buf bytes.Buffer
						if tmpl.Execute(&buf, params) == nil {
							templatedArgs = append(templatedArgs, buf.String())
							continue
						}
					}
					templatedArgs = append(templatedArgs, arg)
				}

				cmd := exec.Command(templatedArgs[0], templatedArgs[1:]...)
				cmd.Dir = destDir
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				if err := cmd.Run(); err != nil {
					return fmt.Errorf("hook %v failed: %w", hookCmd, err)
				}
			}
		}
	}

	return nil
}
