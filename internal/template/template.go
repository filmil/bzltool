package template

import (
	"bytes"
	"os"
	"path/filepath"
	"sort"
	"text/template"
)

// ProcessFragments searches for 'fragments' directories within the provided
// repository directories, sorts them alphanumerically, and evaluates all found
// file templates into the destDir directory using the provided parameters.
// Files with the same relative path are concatenated prior to evaluation.
func ProcessFragments(repoDirs []string, destDir string, params interface{}) error {
	var fragmentDirs []string
	for _, repoDir := range repoDirs {
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

	for _, fDir := range fragmentDirs {
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

			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			fragmentsMap[relPath] = append(fragmentsMap[relPath], string(content))
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

	return nil
}
