package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/filmil/bzltool/internal/config"
	"github.com/filmil/bzltool/internal/git"
	"github.com/filmil/bzltool/internal/template"
	"github.com/spf13/cobra"
)

var (
	projectNameFlag string
	configFlag      string
)

// InitConfig represents the JSON configuration structure passed via --config.
type InitConfig struct {
	Init struct {
		ProjectName string `json:"project_name"`
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

		if configFlag != "" {
			data, err := os.ReadFile(configFlag)
			if err != nil {
				return fmt.Errorf("failed to read config file: %w", err)
			}
			var initCfg InitConfig
			if err := json.Unmarshal(data, &initCfg); err != nil {
				return fmt.Errorf("failed to parse config file: %w", err)
			}
			if initCfg.Init.ProjectName != "" {
				projName = initCfg.Init.ProjectName
			}
		}

		params := TemplateParams{
			ProjectName: projName,
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
