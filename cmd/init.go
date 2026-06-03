package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/filmil/bzltool/internal/config"
	"github.com/filmil/bzltool/internal/git"
	"github.com/filmil/bzltool/internal/template"
	"github.com/spf13/cobra"
)

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

		params := struct{}{}

		if err := template.ProcessFragments(checkedOutDirs, cwd, params); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
