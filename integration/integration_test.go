package integration_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/filmil/bzltool/cmd"
)

func setupTestEnvironment(t *testing.T) (string, string) {
	t.Helper()
	tmpDir := t.TempDir()

	repoDir := filepath.Join(tmpDir, "template_repo")
	if err := os.MkdirAll(filepath.Join(repoDir, "01.core", "fragments"), 0755); err != nil {
		t.Fatal(err)
	}

	gitInit := exec.Command("git", "init")
	gitInit.Dir = repoDir
	if err := gitInit.Run(); err != nil {
		t.Fatal(err)
	}

	gitBranch := exec.Command("git", "branch", "-m", "main")
	gitBranch.Dir = repoDir
	gitBranch.Run() // Ignore error, branch might already be main

	if err := os.WriteFile(filepath.Join(repoDir, "01.core", "fragments", "README.md"), []byte("Project Name: {{.ProjectName}}"), 0644); err != nil {
		t.Fatal(err)
	}

	gitAdd := exec.Command("git", "add", ".")
	gitAdd.Dir = repoDir
	if err := gitAdd.Run(); err != nil {
		t.Fatal(err)
	}

	gitCommit := exec.Command("git", "-c", "user.name=Test", "-c", "user.email=test@example.com", "commit", "-m", "init")
	gitCommit.Dir = repoDir
	if err := gitCommit.Run(); err != nil {
		t.Fatal(err)
	}

	configDir := filepath.Join(tmpDir, ".config")
	bzltoolConfigDir := filepath.Join(configDir, "bzltool")
	if err := os.MkdirAll(bzltoolConfigDir, 0755); err != nil {
		t.Fatal(err)
	}

	configContent := `{"template_repos": ["` + repoDir + `"]}`
	if err := os.WriteFile(filepath.Join(bzltoolConfigDir, "config.json"), []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Set XDG_CONFIG_HOME
	os.Setenv("XDG_CONFIG_HOME", configDir)

	return tmpDir, repoDir
}

func TestE2E_ProjectNameFlag(t *testing.T) {
	tmpDir, _ := setupTestEnvironment(t)
	defer os.Unsetenv("XDG_CONFIG_HOME")

	workDir := filepath.Join(tmpDir, "workdir1")
	if err := os.MkdirAll(workDir, 0755); err != nil {
		t.Fatal(err)
	}
	
	// Change working directory for the command execution
	originalWd, _ := os.Getwd()
	os.Chdir(workDir)
	defer os.Chdir(originalWd)

	err := cmd.ExecuteWithArgs([]string{"init", "--project_name=FlagProject"})
	if err != nil {
		t.Fatalf("ExecuteWithArgs failed: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(workDir, "README.md"))
	if err != nil {
		t.Fatalf("failed to read generated README.md: %v", err)
	}

	if !strings.Contains(string(content), "Project Name: FlagProject") {
		t.Errorf("expected Project Name: FlagProject, got %s", string(content))
	}
}

func TestE2E_JSONConfig(t *testing.T) {
	tmpDir, _ := setupTestEnvironment(t)
	defer os.Unsetenv("XDG_CONFIG_HOME")

	workDir := filepath.Join(tmpDir, "workdir2")
	if err := os.MkdirAll(workDir, 0755); err != nil {
		t.Fatal(err)
	}
	
	originalWd, _ := os.Getwd()
	os.Chdir(workDir)
	defer os.Chdir(originalWd)

	configContent := `{"init": {"project_name": "JsonProject"}}`
	configPath := filepath.Join(workDir, "test_config.json")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	err := cmd.ExecuteWithArgs([]string{"init", "--config=" + configPath})
	if err != nil {
		t.Fatalf("ExecuteWithArgs failed: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(workDir, "README.md"))
	if err != nil {
		t.Fatalf("failed to read generated README.md: %v", err)
	}

	if !strings.Contains(string(content), "Project Name: JsonProject") {
		t.Errorf("expected Project Name: JsonProject, got %s", string(content))
	}
}
