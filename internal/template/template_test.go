package template_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/filmil/bzltool/internal/template"
)

func TestProcessFragments(t *testing.T) {
	tmpDir := t.TempDir()
	repoDir := filepath.Join(tmpDir, "repo")
	destDir := filepath.Join(tmpDir, "dest")

	// Create test repo structure
	// repo/
	//   01.dir1/
	//     fragments/
	//       file1.txt (content: "Hello {{.ProjectName}}")
	//   02.dir2/
	//     fragments/
	//       file1.txt (content: "!")
	//       dir1/
	//         file2.txt (content: "Another")

	err := os.MkdirAll(filepath.Join(repoDir, "01.dir1", "fragments"), 0755)
	if err != nil {
		t.Fatal(err)
	}
	err = os.MkdirAll(filepath.Join(repoDir, "02.dir2", "fragments", "dir1"), 0755)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(filepath.Join(repoDir, "01.dir1", "fragments", "file1.txt"), []byte("Hello {{.ProjectName}}"), 0644)
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(filepath.Join(repoDir, "02.dir2", "fragments", "file1.txt"), []byte("!"), 0644)
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(filepath.Join(repoDir, "02.dir2", "fragments", "dir1", "file2.txt"), []byte("Another"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	params := struct {
		ProjectName string
	}{
		ProjectName: "World",
	}

	err = template.ProcessFragments([]string{repoDir}, destDir, params)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify merged content
	content1, err := os.ReadFile(filepath.Join(destDir, "file1.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(content1) != "Hello World!" {
		t.Errorf("expected 'Hello World!', got '%s'", string(content1))
	}

	content2, err := os.ReadFile(filepath.Join(destDir, "dir1", "file2.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(content2) != "Another" {
		t.Errorf("expected 'Another', got '%s'", string(content2))
	}
}
