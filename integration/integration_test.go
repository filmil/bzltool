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

	// Write a template.json to repo root
	templateJSON := `{"ignore": ["ignored.txt", "ignored_dir/"], "raw": ["*.sh"], "conditions": {"python": ["*.py"]}, "merge_strategies": {"*.json": "json_deep_merge", "*.yaml": "override"}, "hooks": {"post_merge": [["touch", "hook_ran.txt"]]}}`
	if err := os.WriteFile(filepath.Join(repoDir, "template.json"), []byte(templateJSON), 0644); err != nil {
		t.Fatal(err)
	}

	// Create ignored file and raw file
	if err := os.WriteFile(filepath.Join(repoDir, "01.core", "fragments", "ignored.txt"), []byte("this should be ignored"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(repoDir, "01.core", "fragments", "ignored_dir"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repoDir, "01.core", "fragments", "ignored_dir", "foo.txt"), []byte("this should be ignored"), 0644); err != nil {
		t.Fatal(err)
	}
	
	// A raw file that contains {{ }} but won't be templated
	if err := os.WriteFile(filepath.Join(repoDir, "01.core", "fragments", "script.sh"), []byte("echo {{.ProjectName}}"), 0644); err != nil {
		t.Fatal(err)
	}

	// A conditional file
	if err := os.WriteFile(filepath.Join(repoDir, "01.core", "fragments", "server.py"), []byte("import sys"), 0644); err != nil {
		t.Fatal(err)
	}

	// Two json fragments to merge
	if err := os.WriteFile(filepath.Join(repoDir, "01.core", "fragments", "test.json"), []byte(`{"a": 1, "b": {"c": 2}}`), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(repoDir, "02.ext", "fragments"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repoDir, "02.ext", "fragments", "test.json"), []byte(`{"b": {"d": 3}, "e": "{{.ProjectName}}"}`), 0644); err != nil {
		t.Fatal(err)
	}

	// Two yaml fragments to override
	if err := os.WriteFile(filepath.Join(repoDir, "01.core", "fragments", "test.yaml"), []byte("first: 1"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repoDir, "02.ext", "fragments", "test.yaml"), []byte("second: 2"), 0644); err != nil {
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

	// Verify ignored files don't exist
	if _, err := os.Stat(filepath.Join(workDir, "ignored.txt")); !os.IsNotExist(err) {
		t.Errorf("ignored.txt should not exist")
	}
	if _, err := os.Stat(filepath.Join(workDir, "ignored_dir")); !os.IsNotExist(err) {
		t.Errorf("ignored_dir should not exist")
	}

	// Verify conditional file doesn't exist because "python" is not an active condition
	if _, err := os.Stat(filepath.Join(workDir, "server.py")); !os.IsNotExist(err) {
		t.Errorf("server.py should not exist because python condition wasn't met")
	}

	// Verify raw file wasn't templated
	content, err = os.ReadFile(filepath.Join(workDir, "script.sh"))
	if err != nil {
		t.Fatalf("failed to read script.sh: %v", err)
	}
	if string(content) != "echo {{.ProjectName}}" {
		t.Errorf("expected script.sh to remain raw 'echo {{.ProjectName}}', got %s", string(content))
	}

	// Verify deep merged JSON
	content, err = os.ReadFile(filepath.Join(workDir, "test.json"))
	if err != nil {
		t.Fatalf("failed to read test.json: %v", err)
	}
	if !strings.Contains(string(content), `"a": 1`) || !strings.Contains(string(content), `"c": 2`) || !strings.Contains(string(content), `"d": 3`) || !strings.Contains(string(content), `"e": "JsonProject"`) {
		t.Errorf("JSON deep merge failed. Got: %s", string(content))
	}

	// Verify override YAML
	content, err = os.ReadFile(filepath.Join(workDir, "test.yaml"))
	if err != nil {
		t.Fatalf("failed to read test.yaml: %v", err)
	}
	if string(content) != "second: 2" {
		t.Errorf("expected override to give 'second: 2', got: %s", string(content))
	}

	// Verify post_merge hook ran
	if _, err := os.Stat(filepath.Join(workDir, "hook_ran.txt")); os.IsNotExist(err) {
		t.Errorf("hook_ran.txt was not created, meaning post_merge hook did not run")
	}
}

func TestE2E_SubdirConfig(t *testing.T) {
	tmpDir, repoDir := setupTestEnvironment(t)
	defer os.Unsetenv("XDG_CONFIG_HOME")

	// Update XDG config to use a subdir object instead of a string
	bzltoolConfigDir := filepath.Join(tmpDir, ".config", "bzltool")
	// The repo actually doesn't have a "templates" subdir, the fragments are right at root ("01.core/fragments").
	// So let's create a new structure in the repo.
	
	err := os.MkdirAll(filepath.Join(repoDir, "templates", "01.subdir", "fragments"), 0755)
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(filepath.Join(repoDir, "templates", "01.subdir", "fragments", "SUBDIR.md"), []byte("Subdir Project: {{.ProjectName}}"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	gitAdd := exec.Command("git", "add", ".")
	gitAdd.Dir = repoDir
	gitAdd.Run()

	gitCommit := exec.Command("git", "-c", "user.name=Test", "-c", "user.email=test@example.com", "commit", "-m", "add templates subdir")
	gitCommit.Dir = repoDir
	gitCommit.Run()

	configContent := `{"template_repos": [{"url": "` + repoDir + `", "subdir": "templates"}]}`
	if err := os.WriteFile(filepath.Join(bzltoolConfigDir, "config.json"), []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	workDir := filepath.Join(tmpDir, "workdir3")
	if err := os.MkdirAll(workDir, 0755); err != nil {
		t.Fatal(err)
	}
	
	originalWd, _ := os.Getwd()
	os.Chdir(workDir)
	defer os.Chdir(originalWd)

	err = cmd.ExecuteWithArgs([]string{"init", "--project_name=SubdirProject", "--config="})
	if err != nil {
		t.Fatalf("ExecuteWithArgs failed: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(workDir, "SUBDIR.md"))
	if err != nil {
		t.Fatalf("failed to read generated SUBDIR.md: %v", err)
	}

	if !strings.Contains(string(content), "Subdir Project: SubdirProject") {
		t.Errorf("expected Subdir Project: SubdirProject, got %s", string(content))
	}
}

func TestE2E_Health(t *testing.T) {
	// Execute health command
	err := cmd.ExecuteWithArgs([]string{"health"})
	// In the bazel sandbox, `bazel` might not be in PATH, so it may fail.
	// We just ensure it executed the command logic without panicking.
	if err != nil {
		if err.Error() != "health check failed" {
			t.Fatalf("Expected health check failed error in sandbox, got: %v", err)
		}
	}
}

func TestE2E_ToolchainsWorkflows(t *testing.T) {
	tmpDir, repoDir := setupTestEnvironment(t)
	defer os.Unsetenv("XDG_CONFIG_HOME")

	// Create fragments that use .Toolchains and .Languages
	if err := os.MkdirAll(filepath.Join(repoDir, "01.core", "fragments", ".github", "workflows"), 0755); err != nil {
		t.Fatal(err)
	}
	// MODULE.bazel template
	moduleTmpl := `module(name = "{{.ProjectName}}")
{{range .Toolchains}}
go_sdk.download(version = "{{.Version}}")
{{end}}
`
	if err := os.WriteFile(filepath.Join(repoDir, "01.core", "fragments", "MODULE.bazel"), []byte(moduleTmpl), 0644); err != nil {
		t.Fatal(err)
	}

	// .github/workflows/ci.yaml template
	ciTmpl := `{{range .Languages}}
name: CI for {{.}}
{{end}}
`
	if err := os.WriteFile(filepath.Join(repoDir, "01.core", "fragments", ".github", "workflows", "ci.yaml"), []byte(ciTmpl), 0644); err != nil {
		t.Fatal(err)
	}

	gitAdd := exec.Command("git", "add", ".")
	gitAdd.Dir = repoDir
	if err := gitAdd.Run(); err != nil {
		t.Fatal(err)
	}

	gitCommit := exec.Command("git", "-c", "user.name=Test", "-c", "user.email=test@example.com", "commit", "-m", "add toolchains templates")
	gitCommit.Dir = repoDir
	if err := gitCommit.Run(); err != nil {
		t.Fatal(err)
	}

	bzltoolConfigDir := filepath.Join(tmpDir, ".config", "bzltool")
	configContent := `{"template_repos": ["` + repoDir + `"]}`
	if err := os.WriteFile(filepath.Join(bzltoolConfigDir, "config.json"), []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	workDir := filepath.Join(tmpDir, "workdir_tc")
	if err := os.MkdirAll(workDir, 0755); err != nil {
		t.Fatal(err)
	}
	originalWd, _ := os.Getwd()
	os.Chdir(workDir)
	defer os.Chdir(originalWd)

	// Provide a JSON config that configures languages and toolchains
	projectJSON := `{
		"init": {
			"project_name": "TcProject",
			"languages": ["go"],
			"toolchains": [{"lang": "go", "version": "1.20"}],
			"modules": []
		}
	}`
	projConfigPath := filepath.Join(tmpDir, "init.json")
	if err := os.WriteFile(projConfigPath, []byte(projectJSON), 0644); err != nil {
		t.Fatal(err)
	}

	err := cmd.ExecuteWithArgs([]string{"init", "--config=" + projConfigPath})
	if err != nil {
		t.Fatalf("ExecuteWithArgs failed: %v", err)
	}

	// Verify Toolchains integration
	content, err := os.ReadFile(filepath.Join(workDir, "MODULE.bazel"))
	if err != nil {
		t.Fatalf("failed to read MODULE.bazel: %v", err)
	}
	if !strings.Contains(string(content), `go_sdk.download(version = "1.20")`) {
		t.Errorf("MODULE.bazel missing toolchain. Got: %s", string(content))
	}

	// Verify GitHub Workflows integration
	content, err = os.ReadFile(filepath.Join(workDir, ".github", "workflows", "ci.yaml"))
	if err != nil {
		t.Fatalf("failed to read ci.yaml: %v", err)
	}
	if !strings.Contains(string(content), `name: CI for go`) {
		t.Errorf("ci.yaml missing github workflow language. Got: %s", string(content))
	}
}

func TestE2E_CommonConfigs(t *testing.T) {
	tmpDir, repoDir := setupTestEnvironment(t)
	defer os.Unsetenv("XDG_CONFIG_HOME")

	// Create fragments that use .Toolchains and .Languages
	if err := os.MkdirAll(filepath.Join(repoDir, "01.core", "fragments", ".github", "workflows"), 0755); err != nil {
		t.Fatal(err)
	}
	// MODULE.bazel template
	moduleTmpl := `module(name = "{{.ProjectName}}")
{{range .Toolchains}}
go_sdk.download(version = "{{.Version}}")
{{end}}
`
	if err := os.WriteFile(filepath.Join(repoDir, "01.core", "fragments", "MODULE.bazel"), []byte(moduleTmpl), 0644); err != nil {
		t.Fatal(err)
	}

	gitAdd := exec.Command("git", "add", ".")
	gitAdd.Dir = repoDir
	if err := gitAdd.Run(); err != nil {
		t.Fatal(err)
	}

	gitCommit := exec.Command("git", "-c", "user.name=Test", "-c", "user.email=test@example.com", "commit", "-m", "add templates")
	gitCommit.Dir = repoDir
	if err := gitCommit.Run(); err != nil {
		t.Fatal(err)
	}

	bzltoolConfigDir := filepath.Join(tmpDir, ".config", "bzltool")
	configContent := `{"template_repos": ["` + repoDir + `"]}`
	if err := os.WriteFile(filepath.Join(bzltoolConfigDir, "config.json"), []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	workDir := filepath.Join(tmpDir, "workdir_cc")
	if err := os.MkdirAll(workDir, 0755); err != nil {
		t.Fatal(err)
	}
	originalWd, _ := os.Getwd()
	os.Chdir(workDir)
	defer os.Chdir(originalWd)

	err := cmd.ExecuteWithArgs([]string{"init", "--template=Standard Go Server", "--project_name=TemplateProject", "--config="})
	if err != nil {
		t.Fatalf("ExecuteWithArgs failed: %v", err)
	}

	// Verify Toolchains integration
	content, err := os.ReadFile(filepath.Join(workDir, "MODULE.bazel"))
	if err != nil {
		t.Fatalf("failed to read MODULE.bazel: %v", err)
	}
	if !strings.Contains(string(content), `go_sdk.download(version = "1.21")`) {
		t.Errorf("MODULE.bazel missing toolchain version 1.21 from template. Got: %s", string(content))
	}
}
