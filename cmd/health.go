package cmd

import (
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"
)

var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Check the health of external dependencies",
	RunE: func(cmd *cobra.Command, args []string) error {
		tools := []string{"git", "bazel"}
		allGood := true

		fmt.Println("bzltool health check:")
		fmt.Println("---------------------")

		for _, tool := range tools {
			path, err := exec.LookPath(tool)
			if err != nil {
				fmt.Printf("[FAIL] %s is missing from PATH\n", tool)
				allGood = false
			} else {
				fmt.Printf("[ OK ] %s found at %s\n", tool, path)
			}
		}

		fmt.Println("---------------------")
		if allGood {
			fmt.Println("System is healthy!")
		} else {
			fmt.Println("System is missing required dependencies.")
			// We return an error so the CLI exits with non-zero code.
			return fmt.Errorf("health check failed")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(healthCmd)
}
