package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"path/filepath"

	"github.com/filmil/bzltool/internal/config"
	"github.com/filmil/bzltool/internal/git"
	"github.com/filmil/bzltool/internal/template"
	"github.com/filmil/bzltool/internal/tui"
	"github.com/spf13/cobra"
)

var (
	projectNameFlag string
	configFlag      string
	templateFlag    string
)

// Toolchain represents a project toolchain configuration.
type Toolchain struct {
	Lang    string `json:"lang"`
	Version string `json:"version"`
}

// Module represents a project module configuration.
type Module struct {
	Lang    string `json:"lang"`
	Name    string `json:"name"`
	Type    string `json:"type,omitempty"`
	Version string `json:"version,omitempty"`
}

type InitSection struct {
	ProjectName string      `json:"project_name"`
	Languages   []string    `json:"languages,omitempty"`
	Toolchains  []Toolchain `json:"toolchains,omitempty"`
	Modules     []Module    `json:"modules,omitempty"`
}

var commonConfigs = map[string]InitSection{
	"Standard Go Server": {
		Languages:  []string{"go"},
		Toolchains: []Toolchain{{Lang: "go", Version: "1.21"}},
		Modules:    []Module{{Lang: "go", Name: "server"}},
	},
	"Python Data Science Project": {
		Languages:  []string{"python"},
		Toolchains: []Toolchain{{Lang: "python", Version: "3.11"}},
		Modules:    []Module{{Lang: "python", Name: "data_science"}},
	},
	"C++ Library": {
		Languages:  []string{"cpp"},
		Toolchains: []Toolchain{{Lang: "cpp", Version: "17"}},
		Modules:    []Module{{Lang: "cpp", Name: "lib"}},
	},
	"Protobuf": {
		Languages: []string{"protobuf"},
		Modules: []Module{
			{Lang: "bcr", Name: "protobuf", Version: "35.0"},
			{Lang: "bcr", Name: "rules_proto", Version: "7.1.0"},
		},
	},
	"BCR: abseil-cpp": {
		Modules: []Module{{Lang: "bcr", Name: "abseil-cpp", Version: "20260107.1"}},
	},
	"BCR: grpc": {
		Modules: []Module{{Lang: "bcr", Name: "grpc", Version: "1.81.0"}},
	},
	"BCR: rules_cc": {
		Modules: []Module{{Lang: "bcr", Name: "rules_cc", Version: "0.2.19"}},
	},
	"BCR: rules_go": {
		Modules: []Module{{Lang: "bcr", Name: "rules_go", Version: "0.61.0"}},
	},
	"BCR: rules_python": {
		Modules: []Module{{Lang: "bcr", Name: "rules_python", Version: "2.0.2"}},
	},
	"BCR: rules_rust": {
		Modules: []Module{{Lang: "bcr", Name: "rules_rust", Version: "0.70.0"}},
	},
}

// InitConfig represents the JSON configuration structure passed via --config.
type InitConfig struct {
	Init InitSection `json:"init"`
}

// TemplateParams contains the parameters passed to text templates during fragment processing.
type TemplateParams struct {
	ProjectName string
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize project from template repos",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig()
		if err != nil {
			return err
		}

		dir, err := config.GetConfigDir()
		if err != nil {
			return err
		}

		reposDir := filepath.Join(dir, "repos")
		if err := os.MkdirAll(reposDir, 0755); err != nil {
			return err
		}

		var repoUrls []string
		for _, repo := range cfg.TemplateRepos {
			repoUrls = append(repoUrls, repo.URL)
		}

		if err := git.CheckoutRepos(repoUrls, reposDir); err != nil {
			return err
		}

		var checkedOutDirs []string
		for i, repo := range cfg.TemplateRepos {
			checkedOutDirs = append(checkedOutDirs, filepath.Join(reposDir, fmt.Sprintf("repo_%d", i), repo.Subdir))
		}

		cwd, err := os.Getwd()
		if err != nil {
			return err
		}

		projName := projectNameFlag
		var initCfg InitConfig

		if configFlag != "" {
			data, err := os.ReadFile(configFlag)
			if err != nil {
				return fmt.Errorf("failed to read config file: %w", err)
			}
			if err := json.Unmarshal(data, &initCfg); err != nil {
				return fmt.Errorf("failed to parse config file: %w", err)
			}
			if initCfg.Init.ProjectName != "" {
				projName = initCfg.Init.ProjectName
			}
		} else {
			templateStr := templateFlag
			var selectedTemplates []string

			// Only prompt for template interactively if neither flag was provided
			if templateStr == "" && projName == "" {
				// Launch TUI to pick a template
				choices := []string{
					"Standard Go Server", 
					"Python Data Science Project", 
					"C++ Library",
					"Protobuf",
					"BCR: abseil-cpp",
					"BCR: grpc",
					"BCR: rules_cc",
					"BCR: rules_go",
					"BCR: rules_python",
					"BCR: rules_rust",
				}
				var err error
				selectedTemplates, err = tui.PromptTemplates(choices)
				if err != nil {
					return err
				}
			} else if templateStr != "" {
				for _, t := range strings.Split(templateStr, ",") {
					selectedTemplates = append(selectedTemplates, strings.TrimSpace(t))
				}
			}

			for _, templateName := range selectedTemplates {
				if common, ok := commonConfigs[templateName]; ok {
					initCfg.Init.Languages = append(initCfg.Init.Languages, common.Languages...)
					initCfg.Init.Toolchains = append(initCfg.Init.Toolchains, common.Toolchains...)
					initCfg.Init.Modules = append(initCfg.Init.Modules, common.Modules...)
				} else {
					return fmt.Errorf("unknown template: %s", templateName)
				}
			}
		}

		if projName == "" {
			var err error
			projName, err = tui.PromptProjectName()
			if err != nil {
				return err
			}
		}

		// Update the struct with the final project name
		initCfg.Init.ProjectName = projName

		params := TemplateParams{
			ProjectName: projName,
		}

		// Persist the configuration to the project directory
		projectConfigDir := filepath.Join(cwd, ".config", "bzltool")
		if err := os.MkdirAll(projectConfigDir, 0755); err != nil {
			return fmt.Errorf("failed to create project config directory: %w", err)
		}

		configData, err := json.MarshalIndent(initCfg, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal project config: %w", err)
		}

		if err := os.WriteFile(filepath.Join(projectConfigDir, "project_config.json"), configData, 0644); err != nil {
			return fmt.Errorf("failed to write project config: %w", err)
		}

		activeConds := make(map[string]bool)
		for _, l := range initCfg.Init.Languages {
			activeConds[l] = true
		}
		for _, t := range initCfg.Init.Toolchains {
			activeConds[t.Lang] = true
		}
		for _, m := range initCfg.Init.Modules {
			activeConds[m.Lang] = true
		}
		var activeConditions []string
		for k := range activeConds {
			activeConditions = append(activeConditions, k)
		}

		if err := template.ProcessFragments(checkedOutDirs, cwd, params, activeConditions); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	initCmd.Flags().StringVar(&projectNameFlag, "project_name", "", "Name of the project to initialize")
	initCmd.Flags().StringVar(&configFlag, "config", "", "Path to the JSON configuration file")
	rootCmd.AddCommand(initCmd)
}
