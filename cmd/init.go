package cmd

import (
	"encoding/json"
	"fmt"
	"os"
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
)

// Toolchain represents a project toolchain configuration.
type Toolchain struct {
	Lang    string `json:"lang"`
	Version string `json:"version"`
}

// Module represents a project module configuration.
type Module struct {
	Lang string `json:"lang"`
	Name string `json:"name"`
	Type string `json:"type,omitempty"`
}

// InitConfig represents the JSON configuration structure passed via --config.
type InitConfig struct {
	Init struct {
		ProjectName string      `json:"project_name"`
		Languages   []string    `json:"languages,omitempty"`
		Toolchains  []Toolchain `json:"toolchains,omitempty"`
		Modules     []Module    `json:"modules,omitempty"`
	} `json:"init"`
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

		if err := git.CheckoutRepos(cfg.TemplateRepos, reposDir); err != nil {
			return err
		}

		var checkedOutDirs []string
		for i := range cfg.TemplateRepos {
			checkedOutDirs = append(checkedOutDirs, filepath.Join(reposDir, fmt.Sprintf("repo_%d", i)))
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

		if err := template.ProcessFragments(checkedOutDirs, cwd, params); err != nil {
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
